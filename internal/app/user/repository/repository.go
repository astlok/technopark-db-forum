package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"github.com/jackc/pgx"
)

const (
	selectIDByNickname = "SELECT id FROM dbforum.users WHERE nickname = $1"

	selectUsersByForumSlugSinceDesc = "SELECT fu.nickname, fu.fullname, fu.about, fu.email " +
		"FROM dbforum.forum_users AS fu " +
		"WHERE fu.forum_slug = $1 AND fu.nickname < $2 " +
		"ORDER BY fu.nickname DESC " +
		"LIMIT $3"

	selectUsersByForumSlugSince = "SELECT fu.nickname, fu.fullname, fu.about, fu.email " +
		"FROM dbforum.forum_users AS fu " +
		"WHERE fu.forum_slug = $1 " +
		"AND fu.nickname > $2 " +
		"ORDER BY fu.nickname " +
		"LIMIT $3"

	selectUsersByForumSlugDesc = "SELECT fu.nickname, fu.fullname, fu.about, fu.email " +
		"FROM dbforum.forum_users AS fu " +
		"WHERE fu.forum_slug = $1 " +
		"ORDER BY fu.nickname DESC " +
		"LIMIT $2"

	selectUsersByForumSlug = "SELECT fu.nickname, fu.fullname, fu.about, fu.email " +
		"FROM dbforum.forum_users AS fu " +
		"WHERE fu.forum_slug = $1 " +
		"ORDER BY fu.nickname " +
		"LIMIT $2"

	insertUser = `INSERT INTO dbforum.users (
							   nickname, 
							   fullname, 
							   about, 
							   email
                           ) 
                           VALUES (
                                   $1,
                                   $2,
                                   $3,
                                   $4)`

	selectUsersByNickAndEmail = "SELECT nickname, fullname, about, email FROM dbforum.users WHERE nickname = $1 OR email = $2"

	selectByNickname = "SELECT nickname, fullname, about, email FROM dbforum.users WHERE nickname = $1"

	updateUser = `UPDATE dbforum.users SET 
					fullname=COALESCE(NULLIF($1, ''), fullname),
					about=COALESCE(NULLIF($2, ''), about),
					email=COALESCE(NULLIF($3, ''), email)
					WHERE nickname=$4 RETURNING nickname, fullname, about, email`

	selectNickByEmail = "SELECT nickname FROM dbforum.users WHERE email = $1"
)

type Repository struct {
	db *pgx.ConnPool
}

func NewRepo(db *pgx.ConnPool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetForumUsers(forumSlug string, limit int, since string, desc bool) ([]models.User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	var users []models.User
	row, err := tx.Query("checkForumExist", forumSlug)

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if !row.Next() {
		_ = tx.Rollback()
		return nil, customErr.ErrForumNotFound
	}
	row.Close()
	if since == "" {
		if desc {
			row, err = r.db.Query("selectUsersByForumSlugDesc", forumSlug, limit)
		} else {
			row, err = r.db.Query("selectUsersByForumSlug", forumSlug, limit)
		}
	} else {
		if desc {
			row, err = r.db.Query("selectUsersByForumSlugSinceDesc", forumSlug, since, limit)
		} else {
			row, err = r.db.Query("selectUsersByForumSlugSince", forumSlug, since, limit)
		}
	}

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	for row.Next() {
		u := models.User{}
		err := row.Scan(
			&u.Nickname,
			&u.Fullname,
			&u.About,
			&u.Email)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		users = append(users, u)
	}
	row.Close()
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	return users, nil
}

func (r *Repository) CreateUser(user models.User) error {
	_, err := r.db.Exec("insertUser", &user.Nickname, &user.Fullname, &user.About, &user.Email)
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
	rows, err := r.db.Query(selectUsersByNickAndEmail, nickname, email)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		u := models.User{}
		err := rows.Scan(
			&u.Nickname,
			&u.Fullname,
			&u.About,
			&u.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	rows.Close()
	return users, nil
}

func (r *Repository) GetUserByNick(nickname string) (*models.User, error) {
	var user models.User
	rows, err := r.db.Query("selectByNickname", nickname)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, customErr.ErrUserNotFound
	}
	err = rows.Scan(
		&user.Nickname,
		&user.Fullname,
		&user.About,
		&user.Email)
	if err != nil {
		return nil, err
	}
	rows.Close()
	return &user, nil
}

func (r *Repository) ChangeUser(user *models.User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	err = tx.QueryRow("updateUser", &user.Fullname, &user.About, &user.Email, &user.Nickname).Scan(
		&user.Nickname,
		&user.Fullname,
		&user.About,
		&user.Email)
	if driverErr, ok := err.(pgx.PgError); ok {
		if driverErr.Code == "23505" {
			_ = tx.Rollback()
			return customErr.ErrConflict
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return customErr.ErrUserNotFound
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func (r *Repository) GetUserNickByEmail(email string) (string, error) {
	var nickname string
	rows, err := r.db.Query(selectNickByEmail, email)
	if err != nil {
		rows.Close()
		return "", err
	}
	if !rows.Next() {
		rows.Close()
		return "", customErr.ErrUserNotFound
	}
	err = rows.Scan(&nickname)
	rows.Close()
	if err != nil {
		return "", err
	}
	return nickname, nil
}

func (r *Repository) Prepare() error {
	_, err := r.db.Prepare("insertUser", insertUser)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("updateUser", updateUser)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectUsersByForumSlugSinceDesc", selectUsersByForumSlugSinceDesc)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectUsersByForumSlugSince", selectUsersByForumSlugSince)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectUsersByForumSlugDesc", selectUsersByForumSlugDesc)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectUsersByForumSlug", selectUsersByForumSlug)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("checkForumExist", "SELECT 1 FROM dbforum.forum WHERE slug = $1 LIMIT 1")
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectByNickname", selectByNickname)
	if err != nil {
		return err
	}

	return nil
}
