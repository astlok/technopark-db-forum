package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"database/sql"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const (
	insertPost = `INSERT INTO dbforum.post(author_nickname, forum_slug, thread_id, parent, created, message)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING ID`

	selectByThreadIDFlatDesc = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN id < $2 ELSE TRUE END ORDER BY id DESC LIMIT $3"

	selectByThreadIDFlat = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN id > $2 ELSE TRUE END ORDER BY id LIMIT $3"

	selectByThreadIDTreeDesc = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN tree < (SELECT tree FROM dbforum.post WHERE id=$2) ELSE TRUE END ORDER BY tree DESC LIMIT $3"

	selectByThreadIDTree = "SELECT * FROM dbforum.post WHERE thread_id=$1 AND CASE WHEN $2 > 0 THEN tree > (SELECT tree FROM dbforum.post WHERE id=$2) ELSE TRUE END ORDER BY tree LIMIT $3"

	selectByThreadIDParentTreeDesc = "SELECT * FROM dbforum.post WHERE tree[1] IN (SELECT id FROM dbforum.post WHERE thread_id = $1 AND parent = 0 AND CASE WHEN $3 > 0 THEN tree[1] < (SELECT tree[1] FROM dbforum.post WHERE id=$3) ELSE TRUE END ORDER BY id DESC LIMIT $2) ORDER BY tree[1] DESC, tree, id"

	selectByThreadIDParentTree = "SELECT * FROM dbforum.post WHERE tree[1] IN (SELECT id FROM dbforum.post WHERE thread_id = $1 AND parent = 0  AND CASE WHEN $3 > 0 THEN tree[1] > (SELECT tree[1] FROM dbforum.post WHERE id=$3) ELSE TRUE END ORDER BY id LIMIT $2) ORDER BY tree, id"

	selectPostByID = "SELECT * FROM dbforum.post WHERE id=$1"

	updatePost = `UPDATE dbforum.post SET message=COALESCE(NULLIF($1, ''), message),
                	is_edited = CASE WHEN $1 = '' OR message = $1 THEN is_edited ELSE true END
					WHERE id=$2 
					RETURNING id, author_nickname, forum_slug, thread_id, message, parent, is_edited, created`
)

type Repository struct {
	db *pgx.ConnPool
}

func NewRepo(db *pgx.ConnPool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreatePosts(idOrSlug string, posts []models.Post) ([]models.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		posts = append(posts, models.Post{Forum: idOrSlug})
	}
	var threadID uint64
	var forumSlug string
	if threadID, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		rows, err := tx.Query("selectThreadIDAndForumSlug", idOrSlug)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if !rows.Next() {
			_ = tx.Rollback()
			return nil, customErr.ErrThreadNotFound
		}
		err = rows.Scan(&threadID, &forumSlug)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		rows.Close()
	} else {
		rows, err := tx.Query("selectForumSlug", threadID)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if !rows.Next() {
			_ = tx.Rollback()
			return nil, customErr.ErrThreadNotFound
		}
		err = rows.Scan(&forumSlug)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		rows.Close()
	}
	err = nil
	if posts[0].Parent != 0 {
		var parent uint64
		rows, err := tx.Query("selectThreadIDFromPost", posts[0].Parent)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if rows.Next() {
			err := rows.Scan(&parent)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		}
		if parent != threadID {
			_ = tx.Rollback()
			rows.Close()
			return nil, customErr.ErrNoParent
		}
		rows.Close()
	}

	created := strfmt.DateTime(time.Now())
	query := "INSERT INTO dbforum.post(author_nickname, forum_slug, thread_id, parent, created, message) VALUES "
	var args []interface{}
	for i, post := range posts {
		posts[i].Created = created
		posts[i].Thread = threadID
		posts[i].Forum = forumSlug
		if post.Author != "" {
			row, err := tx.Query("selectPostAuthor", post.Author)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			if !row.Next() {
				_ = tx.Rollback()
				return nil, errors.Wrap(customErr.ErrUserNotFound, post.Author)
			}
			row.Close()
		} else {
			_ = tx.Rollback()
			return nil, nil
		}

		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)
		if i != len(posts)-1 {
			query += ","
		} else {
			query += " RETURNING ID"
		}
		args = append(args, post.Author, forumSlug, threadID, post.Parent, created, post.Message)
	}
	rows, err := tx.Query(query, args...)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	index := 0
	for rows.Next() {
		err = rows.Scan(&posts[index].ID)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		index++
	}
	rows.Close()
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return posts, nil
}

func (r *Repository) GetPosts(idOrSlug string, limit int64, since int64, desc bool, sort string) ([]models.Post, error) {
	var posts []models.Post
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	var threadID uint64
	if threadID, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		rows, err := tx.Query("selectIDFromThread", idOrSlug)
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
		rows, err := tx.Query("checkThreadExists", threadID)
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
			rows, err = tx.Query("selectByThreadIDFlatDesc", threadID, since, limit)
		case "tree":
			rows, err = tx.Query("selectByThreadIDTreeDesc", threadID, since, limit)
		case "parent_tree":
			rows, err = tx.Query("selectByThreadIDParentTreeDesc", threadID, limit, since)
		default:
			rows, err = tx.Query("selectByThreadIDFlatDesc", threadID, since, limit)
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
			rows, err = tx.Query("selectByThreadIDFlat", threadID, since, limit)
		case "tree":
			rows, err = tx.Query("selectByThreadIDTree", threadID, since, limit)
		case "parent_tree":
			rows, err = tx.Query("selectByThreadIDParentTree", threadID, limit, since)
		default:
			rows, err = tx.Query("selectByThreadIDFlat", threadID, since, limit)
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

func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (r *Repository) GetPostInfoByID(id uint64, related []string) (*models.PostInfo, error) {
	tx, _ := r.db.Begin()
	postInfo := models.PostInfo{
		Post: &models.Post{},
	}
	rows, err := tx.Query("selectPostByID", id)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if !rows.Next() {
		_ = tx.Rollback()
		return nil, customErr.ErrPostNotFound
	}
	err = rows.Scan(
		&postInfo.Post.ID,
		&postInfo.Post.Author,
		&postInfo.Post.Forum,
		&postInfo.Post.Thread,
		&postInfo.Post.Message,
		&postInfo.Post.Parent,
		&postInfo.Post.IsEdited,
		&postInfo.Post.Created,
		&postInfo.Post.Tree)
	rows.Close()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if Find(related, "user") {
		rows, err := tx.Query("selectByNickname", postInfo.Post.Author)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if rows.Next() {
			postInfo.Author = &models.User{}
			err = rows.Scan(
				&postInfo.Author.Nickname,
				&postInfo.Author.Fullname,
				&postInfo.Author.About,
				&postInfo.Author.Email)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		}
		rows.Close()
	}
	if Find(related, "thread") {
		rows, err := tx.Query("selectThreadByID", postInfo.Post.Thread)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if rows.Next() {
			postInfo.Thread = &models.Thread{}
			err = rows.Scan(
				&postInfo.Thread.ID,
				&postInfo.Thread.Forum,
				&postInfo.Thread.Author,
				&postInfo.Thread.Title,
				&postInfo.Thread.Message,
				&postInfo.Thread.Votes,
				&postInfo.Thread.Slug,
				&postInfo.Thread.Created)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		}
		rows.Close()
	}

	if Find(related, "forum") {
		rows, err := tx.Query("selectForumBySlug", postInfo.Post.Forum)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if rows.Next() {
			postInfo.Forum = &models.Forum{}
			err = rows.Scan(
				&postInfo.Forum.User,
				&postInfo.Forum.Title,
				&postInfo.Forum.Slug,
				&postInfo.Forum.Posts,
				&postInfo.Forum.Threads)
			rows.Close()
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return &postInfo, nil
}

func (r *Repository) ChangePost(post *models.Post) (models.Post, error) {
	err := r.db.QueryRow("updatePost", &post.Message, &post.ID).Scan(
		&post.ID,
		&post.Author,
		&post.Forum,
		&post.Thread,
		&post.Message,
		&post.Parent,
		&post.IsEdited,
		&post.Created)
	if err != nil {
		return models.Post{}, customErr.ErrPostNotFound
	}
	return *post, nil
}

func (r *Repository) Prepare() error {
	_, err := r.db.Prepare("insertPost", insertPost)
	if err != nil {
		return err
	}
	_, err = r.db.Prepare("selectThreadIDAndForumSlug", "SELECT id, forum_slug FROM dbforum.thread WHERE slug=$1 LIMIT 1")
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectForumSlug", "SELECT forum_slug FROM dbforum.thread WHERE id=$1 LIMIT 1")
	if err != nil {
		return err
	}
	_, err = r.db.Prepare("selectThreadIDFromPost", "SELECT thread_id FROM dbforum.post WHERE id = $1")
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectPostAuthor", "SELECT 1 FROM dbforum.users WHERE nickname=$1 LIMIT 1")
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectIDFromThread", "SELECT id FROM dbforum.thread WHERE slug=$1 LIMIT 1")

	if err != nil {
		return err
	}

	_, err = r.db.Prepare("checkThreadExists", "SELECT 1 FROM dbforum.thread WHERE id=$1 LIMIT 1")
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("updatePost", updatePost)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectPostByID", selectPostByID)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByThreadIDFlatDesc", selectByThreadIDFlatDesc)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByThreadIDTreeDesc", selectByThreadIDTreeDesc)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByThreadIDParentTreeDesc", selectByThreadIDParentTreeDesc)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByThreadIDFlatDesc", selectByThreadIDFlatDesc)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByThreadIDFlat", selectByThreadIDFlat)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByThreadIDTree", selectByThreadIDTree)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByThreadIDParentTree", selectByThreadIDParentTree)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByThreadIDFlat", selectByThreadIDFlat)
	if err != nil {
		return err
	}

	return nil
}
