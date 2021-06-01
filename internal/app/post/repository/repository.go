package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"database/sql"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"strconv"
)

const (
	insertPost = `INSERT INTO dbforum.post(author_nickname, forum_slug, thread_id, parent, created, tree, message)
				VALUES ($1, $2, $3, $4, $5,
						$6 || ARRAY [(SELECT currval(pg_get_serial_sequence('dbforum.post', 'id')))], $7)
				RETURNING ID`

	selectByThreadIDFlatDesc = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN id < $2 ELSE TRUE END ORDER BY id DESC LIMIT $3"

	selectByThreadIDFlat = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN id > $2 ELSE TRUE END ORDER BY id LIMIT $3"

	selectByThreadIDTreeDesc = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN tree < (SELECT tree FROM dbforum.post WHERE id=$2) ELSE TRUE END ORDER BY tree DESC LIMIT $3"

	selectByThreadIDTree = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN tree > (SELECT tree FROM dbforum.post WHERE id=$2) ELSE TRUE END ORDER BY tree LIMIT $3"

	selectByThreadIDParentTreeDesc = "SELECT * FROM dbforum.post WHERE tree[1] IN (SELECT id FROM dbforum.post WHERE thread_id = $1 AND parent = 0 AND CASE WHEN $3 > 0 THEN tree[1] < (SELECT tree[1] FROM dbforum.post WHERE id=$3) ELSE TRUE END ORDER BY id DESC LIMIT $2) ORDER BY tree[1] DESC, tree, id"

	selectByThreadIDParentTree = "SELECT * FROM dbforum.post WHERE tree[1] IN (SELECT id FROM dbforum.post WHERE thread_id = $1 AND parent = 0  AND CASE WHEN $3 > 0 THEN tree[1] > (SELECT tree[1] FROM dbforum.post WHERE id=$3) ELSE TRUE END ORDER BY id LIMIT $2) ORDER BY tree, id"

	selectPostByID = "SELECT * FROM dbforum.post WHERE id=$1"

	updatePost = `UPDATE dbforum.post SET
					message=$1,
                    is_edited=$2
					WHERE id=$3`
)

type Repository struct {
	db *pgx.ConnPool
}

func NewRepo(db *pgx.ConnPool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreatePost(idOrSlug string, post models.Post) (models.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return models.Post{}, err
	}
	var threadID uint64
	if threadID, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		rows, err := tx.Query("SELECT id as thread_id, forum_slug FROM dbforum.thread WHERE slug=$1 LIMIT 1", idOrSlug)
		if err != nil {
			_ = tx.Rollback()
			return models.Post{}, err
		}
		if !rows.Next() {
			_ = tx.Rollback()
			return models.Post{}, customErr.ErrThreadNotFound
		}
		err = rows.Scan(&post.Thread, &post.Forum)
		rows.Close()
	} else {
		post.Thread = threadID
		rows, err := tx.Query("SELECT forum_slug FROM dbforum.thread WHERE id=$1 LIMIT 1", threadID)
		if err != nil {
			_ = tx.Rollback()
			return models.Post{}, err
		}
		if !rows.Next() {
			_ = tx.Rollback()
			return models.Post{}, customErr.ErrThreadNotFound
		}
		err = rows.Scan(&post.Forum)
		rows.Close()
	}
	if post.Author != "" {
		row, err := tx.Query("SELECT 1 FROM dbforum.users WHERE nickname=$1 LIMIT 1", post.Author)
		if err != nil {
			_ = tx.Rollback()
			return models.Post{}, err
		}
		if !row.Next() {
			_ = tx.Rollback()
			return models.Post{}, errors.Wrap(customErr.ErrUserNotFound, post.Author)
		}
		row.Close()
	} else {
		_ = tx.Rollback()
		return models.Post{}, nil
	}

	if post.Parent != 0 {
		row, err := tx.Query("SELECT tree FROM dbforum.post WHERE id=$1 AND thread_id=$2 LIMIT 1", post.Parent, post.Thread)
		if err != nil {
			_ = tx.Rollback()
			return models.Post{}, err
		}
		if !row.Next() {
			_ = tx.Rollback()
			return models.Post{}, customErr.ErrNoParent
		}
		err = row.Scan(&post.Tree)
		row.Close()
	}
	err = tx.QueryRow(
		insertPost,
		post.Author,
		post.Forum,
		post.Thread,
		post.Parent,
		post.Created,
		post.Tree,
		post.Message).Scan(&post.ID)
	if err != nil {
		_ = tx.Rollback()
		return models.Post{}, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return models.Post{}, err
	}
	return post, nil
}

func (r *Repository) GetPosts(idOrSlug string, limit int64, since int64, desc bool, sort string) ([]models.Post, error) {
	var posts []models.Post
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	var threadID uint64
	if threadID, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		rows, err := tx.Query("SELECT id FROM dbforum.thread WHERE slug=$1 LIMIT 1", idOrSlug)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if !rows.Next() {
			_ = tx.Rollback()
			return nil, customErr.ErrThreadNotFound
		}
		err = rows.Scan(&threadID)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		rows.Close()
	} else {
		rows, err := tx.Query("SELECT 1 FROM dbforum.thread WHERE id=$1 LIMIT 1", threadID)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if !rows.Next() {
			_ = tx.Rollback()
			return nil, customErr.ErrThreadNotFound
		}
		rows.Close()
	}

	var rows *pgx.Rows
	if desc {
		switch sort {
		case "flat":
			rows, err = tx.Query(selectByThreadIDFlatDesc, threadID, since, limit)
		case "tree":
			rows, err = tx.Query(selectByThreadIDTreeDesc, threadID, since, limit)
		case "parent_tree":
			rows, err = tx.Query(selectByThreadIDParentTreeDesc, threadID, limit, since)
		default:
			rows, err = tx.Query(selectByThreadIDFlatDesc, threadID, since, limit)
		}
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return []models.Post{}, nil
		}
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		switch sort {
		case "flat":
			rows, err = tx.Query(selectByThreadIDFlat, threadID, since, limit)
		case "tree":
			rows, err = tx.Query(selectByThreadIDTree, threadID, since, limit)
		case "parent_tree":
			rows, err = tx.Query(selectByThreadIDParentTree, threadID, limit, since)
		default:
			rows, err = tx.Query(selectByThreadIDFlat, threadID, since, limit)
		}
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return []models.Post{}, nil
		}
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	for rows.Next() {
		p := models.Post{}
		err := rows.Scan(
			&p.ID,
			&p.Author,
			&p.Forum,
			&p.Thread,
			&p.Message,
			&p.Parent,
			&p.IsEdited,
			&p.Created,
			&p.Tree)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		posts = append(posts, p)
	}
	rows.Close()
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	return posts, nil
}

func (r *Repository) GetPostByID(id uint64) (*models.Post, error) {
	post := models.Post{}
	rows, err := r.db.Query(selectPostByID, id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, customErr.ErrPostNotFound
	}
	err = rows.Scan(
		&post.ID,
		&post.Author,
		&post.Forum,
		&post.Thread,
		&post.Message,
		&post.Parent,
		&post.IsEdited,
		&post.Created,
		&post.Tree)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *Repository) ChangePost(post *models.Post) error {
	_, err := r.db.Exec(updatePost, &post.Message, &post.IsEdited, &post.ID)
	if err != nil {
		return err
	}
	return nil
}
