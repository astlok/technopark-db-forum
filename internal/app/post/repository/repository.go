package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strconv"
)

const (
	insertPost = `INSERT INTO dbforum.post(author_nickname, forum_slug, thread_id, parent, created, tree, message)
				VALUES ($1, $2, $3, $4, $5,
						CONCAT(cast($6 as text), CAST((SELECT currval(pg_get_serial_sequence('dbforum.post', 'id'))) as text)), $7)
				RETURNING ID`

	selectByThreadIDFlatDesc = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN id < $2 ELSE TRUE END ORDER BY id DESC LIMIT $3"

	selectByThreadIDFlat = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN id > $2 ELSE TRUE END ORDER BY id LIMIT $3"

	selectByThreadIDTreeDesc = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN id < $2 ELSE TRUE END ORDER BY split_part(tree, '.', 1), tree DESC LIMIT $3"

	selectByThreadIDTree = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2> 0 THEN id > $2 ELSE TRUE END ORDER BY split_part(tree, '.', 1), tree LIMIT $3"

	selectByThreadIDParentTreeDesc = "SELECT * FROM dbforum.post WHERE cast(split_part(tree, '.', 1) AS BIGINT) IN (SELECT id FROM dbforum.post WHERE thread_id = $1 AND parent = 0 LIMIT $2) AND CASE WHEN $3 > 0 THEN id < $3 ELSE TRUE END ORDER BY split_part(tree, '.', 1) DESC, tree, id"

	selectByThreadIDParentTree = "SELECT * FROM dbforum.post WHERE cast(split_part(tree, '.', 1) AS BIGINT) IN (SELECT id FROM dbforum.post WHERE thread_id = $1 AND parent = 0 LIMIT $2) AND CASE WHEN $3 > 0 THEN id > $3 ELSE TRUE END ORDER BY split_part(tree, '.', 1), tree,  id"

	selectPostByID = "SELECT * FROM dbforum.post WHERE id=$1"

	updatePost = `UPDATE dbforum.post SET
					message=:message,
                    is_edited=:is_edited
					WHERE id=:id`
)

type Repository struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreatePost(idOrSlug string, post models.Post) (models.Post, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return models.Post{}, err
	}
	var threadID uint64
	if threadID, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		err = tx.Get(&post, "SELECT id as thread_id, forum_slug FROM dbforum.thread WHERE slug=$1 LIMIT 1", idOrSlug)
	} else {
		post.Thread = threadID
		err = tx.Get(&post, "SELECT forum_slug FROM dbforum.thread WHERE id=$1 LIMIT 1", threadID)
	}
	if errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return models.Post{}, customErr.ErrThreadNotFound
	}
	row, err := tx.Query("SELECT 1 FROM dbforum.users WHERE nickname=$1 LIMIT 1", post.Author)
	if err != nil {
		_ = tx.Rollback()
		return models.Post{}, err
	}
	if !row.Next() {
		_ = tx.Rollback()
		return models.Post{}, errors.Wrap(customErr.ErrUserNotFound, post.Author)
	}
	if err := row.Close(); err != nil {
		_ = tx.Rollback()
		return models.Post{}, err
	}
	if post.Parent != 0 {
		row, err = tx.Query("SELECT tree FROM dbforum.post WHERE id=$1 AND thread_id=$2 LIMIT 1", post.Parent, post.Thread)
		if err != nil {
			_ = tx.Rollback()
			return models.Post{}, err
		}
		if !row.Next() {
			_ = tx.Rollback()
			return models.Post{}, customErr.ErrNoParent
		}
		err = row.Scan(&post.Tree)
		if err := row.Close(); err != nil {
			_ = tx.Rollback()
			return models.Post{}, err
		}
	}
	if post.Tree != "" {
		post.Tree += "."
	}
	err = tx.QueryRowx(
		insertPost,
		post.Author,
		post.Forum,
		post.Thread,
		post.Parent,
		post.Created,
		post.Tree,
		post.Message).Scan(&post.ID)
	row, err = tx.Query("SELECT 1 FROM dbforum.forum_users WHERE forum_slug=$1 AND nickname=$2", post.Forum, post.Author)
	if err != nil {
		_ = tx.Rollback()
		return models.Post{}, err
	}
	if !row.Next() {
		query := fmt.Sprintf("INSERT INTO dbforum.forum_users(forum_slug, nickname, fullname, about, email) "+
			"SELECT '%s', nickname, fullname, about, email FROM dbforum.users "+
			"WHERE nickname = '%s'", post.Forum, post.Author)
		if _, err := tx.Exec(query); err != nil {
			_ = tx.Rollback()
			return models.Post{}, err
		}
	}
	if err := row.Close(); err != nil {
		return models.Post{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Post{}, err
	}
	return post, nil
}

func (r *Repository) GetPosts(idOrSlug string, limit int64, since int64, desc bool, sort string) ([]models.Post, error) {
	var posts []models.Post
	tx, err := r.db.Beginx()
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
		if err := rows.Close(); err != nil {
			return nil, err
		}
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
		if err := rows.Close(); err != nil {
			return nil, err
		}
	}

	if desc {
		switch sort {
		case "flat":
			err = tx.Select(&posts, selectByThreadIDFlatDesc, threadID, since, limit)
		case "tree":
			err = tx.Select(&posts, selectByThreadIDTreeDesc, threadID, since, limit)
		case "parent_tree":
			err = tx.Select(&posts, selectByThreadIDParentTreeDesc, threadID, limit, since)
		default:
			err = tx.Select(&posts, selectByThreadIDFlatDesc, threadID, since, limit)
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
			err = tx.Select(&posts, selectByThreadIDFlat, threadID, since, limit)
		case "tree":
			err = tx.Select(&posts, selectByThreadIDTree, threadID, since, limit)
		case "parent_tree":
			err = tx.Select(&posts, selectByThreadIDParentTree, threadID, limit, since)
		default:
			err = tx.Select(&posts, selectByThreadIDFlat, threadID, since, limit)
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
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *Repository) GetPostByID(id uint64) (*models.Post, error) {
	post := models.Post{}
	if err := r.db.Get(&post, selectPostByID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *Repository) ChangePost(post *models.Post) error {
	_, err := r.db.NamedExec(updatePost, &post)
	if err != nil {
		return err
	}
	return nil
}
