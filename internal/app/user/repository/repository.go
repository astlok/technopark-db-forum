package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"database/sql"
	"github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	//TODO: навесить индекс на никнейм
	selectIDByNickname = "SELECT id FROM dbforum.users WHERE nickname = $1"

	selectUsersByForumSlugDesc = "SELECT fu.nickname, fu.fullname, fu.about, fu.email " +
		"FROM dbforum.forum_users AS fu " +
		"WHERE fu.forum_slug = $1 AND CASE WHEN $2 != '' THEN fu.nickname < $2 ELSE TRUE END " +
		"ORDER BY fu.nickname DESC " +
		"LIMIT $3"

	selectUsersByForumSlug = "SELECT fu.nickname, fu.fullname, fu.about, fu.email " +
		"FROM dbforum.forum_users AS fu " +
		"WHERE fu.forum_slug = $1 AND CASE WHEN $2 != '' THEN fu.nickname > $2 ELSE TRUE END " +
		"ORDER BY fu.nickname " +
		"LIMIT $3"

	insertUser = `INSERT INTO dbforum.users (
							   nickname, 
							   fullname, 
							   about, 
							   email
                           ) 
                           VALUES (
                                   :nickname,
                                   :fullname,
                                   :about,
                                   :email)`

	selectUsersByNickAndEmail = "SELECT nickname, fullname, about, email FROM dbforum.users WHERE nickname = $1 OR email = $2"

	selectByNickname = "SELECT nickname, fullname, about, email FROM dbforum.users WHERE nickname = $1"

	updateUser = `UPDATE dbforum.users SET 
							 fullname=:fullname,
							 about=:about,
							 email=:email
                         WHERE nickname=:nickname`

	selectNickByEmail = "SELECT nickname FROM dbforum.users WHERE email = $1"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CheckUserExists(nickname string) (uint64, error) {
	var userID uint64
	if err := r.db.Get(&userID, selectIDByNickname, nickname); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, customErr.ErrUserNotFound
		}
		return 0, err
	}
	return userID, nil
}

func (r *Repository) GetForumUsers(forumSlug string, limit int64, since string, desc bool) ([]models.User, error) {
	var users []models.User
	var err error
	if desc {
		err = r.db.Select(&users, selectUsersByForumSlugDesc, forumSlug, since, limit)
	} else {
		err = r.db.Select(&users, selectUsersByForumSlug, forumSlug, since, limit)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, customErr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) CreateUser(user models.User) error {
	_, err := r.db.NamedExec(insertUser, &user)
	if driverErr, ok := err.(pgx.PgError); ok {
		if driverErr.Code == "23505" {
			return customErr.ErrDuplicate
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetUsersByNickAndEmail(nickname string, email string) ([]models.User, error) {
	var users []models.User
	err := r.db.Select(&users, selectUsersByNickAndEmail, nickname, email)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) GetUserByNick(nickname string) (*models.User, error) {
	var user models.User
	if err := r.db.Get(&user, selectByNickname, nickname); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) ChangeUser(user *models.User) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	var oldUser models.User
	if err := tx.Get(&oldUser, selectByNickname, user.Nickname); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return customErr.ErrUserNotFound
		}
		return err
	}
	if user.Fullname == "" {
		user.Fullname = oldUser.Fullname
	}
	if user.About == "" {
		user.About = oldUser.About
	}
	if user.Email == "" {
		user.Email = oldUser.Email
	}
	_, err = tx.NamedExec(updateUser, &user)
	if driverErr, ok := err.(pgx.PgError); ok {
		if driverErr.Code == "23505" {
			_ = tx.Rollback()
			return customErr.ErrConflict
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

func (r *Repository) GetUserNickByEmail(email string) (string, error) {
	var nickname string
	if err := r.db.Get(&nickname, selectNickByEmail, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", customErr.ErrUserNotFound
		}
		return "", err
	}
	return nickname, nil
}
