package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"github.com/jackc/pgx"
)

const (
	insertForum = `INSERT INTO dbforum.forum (
							   user_nickname, 
							   title, 
							   slug
                           ) 
                           VALUES (
                                   $1,
                                   $2,
                                   $3
                           )`
	selectForumBySlug = "SELECT user_nickname, title, slug, posts, threads FROM dbforum.forum WHERE slug = $1"

	selectNicknameByNickname = "SELECT nickname FROM dbforum.users WHERE nickname = $1"
)

type Repository struct {
	db *pgx.ConnPool
}

func NewRepo(db *pgx.ConnPool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateForum(forum *models.Forum) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	rows, err := tx.Query("selectForumBySlug", &forum.Slug)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if rows.Next() {
		err = rows.Scan(
			&forum.User,
			&forum.Title,
			&forum.Slug,
			&forum.Posts,
			&forum.Threads)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		rows.Close()
		_ = tx.Rollback()
		return customErr.ErrDuplicate
	}

	rows.Close()
	var nickname string
	rows, err = tx.Query("selectNicknameByNickname", forum.User)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if !rows.Next() {
		_ = tx.Rollback()
		return customErr.ErrUserNotFound
	}
	err = rows.Scan(&nickname)
	rows.Close()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	forum.User = nickname
	_, err = tx.Exec(
		"insertForum",
		forum.User,
		forum.Title,
		forum.Slug)

	if driverErr, ok := err.(pgx.PgError); ok {
		if driverErr.Code == "23505" {
			_ = tx.Rollback()
			return customErr.ErrDuplicate
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func (r *Repository) FindBySlug(slug string) (*models.Forum, error) {
	forum := models.Forum{}
	rows, err := r.db.Query("selectForumBySlug", slug)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, customErr.ErrForumNotFound
	}
	err = rows.Scan(
		&forum.User,
		&forum.Title,
		&forum.Slug,
		&forum.Posts,
		&forum.Threads)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return &forum, nil
}

func (r *Repository) Prepare() error {
	_, err := r.db.Prepare("insertForum", insertForum)
	if err != nil {
		return err
	}
	_, err = r.db.Prepare("selectForumBySlug", selectForumBySlug)
	if err != nil {
		return err
	}
	_, err = r.db.Prepare("selectNicknameByNickname", selectNicknameByNickname)
	if err != nil {
		return err
	}
	return nil
}
