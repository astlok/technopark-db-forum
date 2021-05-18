package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"database/sql"
	"errors"
	"github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
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

	selectIDBySlug = "SELECT id FROM dbforum.forum WHERE slug = $1"

	selectNicknameByNickname = "SELECT nickname FROM dbforum.users WHERE nickname = $1"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateForum(forum *models.Forum)  error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	var nickname string
	if err := tx.Get(&nickname, selectNicknameByNickname, forum.User); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return customErr.ErrUserNotFound
		}
		return err
	}
	forum.User = nickname
	_, err = tx.Exec(
		insertForum,
		forum.User,
		forum.Title,
		forum.Slug)

	if driverErr, ok := err.(pgx.PgError); ok {
		if driverErr.Code == "23505" {
			_ = tx.Rollback()
			return  customErr.ErrDuplicate
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) FindBySlug(slug string) (*models.Forum, error) {
	forum := models.Forum{}
	if err := r.db.Get(&forum, selectForumBySlug, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrForumNotFound
		}
		return nil, err
	}
	return &forum, nil
}

func (r *Repository) CheckForumExists(forumSlug string) (uint64, error) {
	var forumID uint64
	if err := r.db.Get(&forumID, selectIDBySlug, forumSlug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, customErr.ErrForumNotFound
		}
		return 0, err
	}
	return forumID, nil
}
