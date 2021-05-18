package repository

import (
	"DBForum/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) ClearDB() error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`TRUNCATE dbforum.users CASCADE;`)
	_, err = tx.Exec(`TRUNCATE dbforum.forum_users CASCADE;`)
	_, err = tx.Exec(`TRUNCATE dbforum.thread CASCADE;`)
	_, err = tx.Exec(`TRUNCATE dbforum.post CASCADE;`)
	_, err = tx.Exec(`TRUNCATE dbforum.forum CASCADE;`)
	_, err = tx.Exec(`TRUNCATE dbforum.votes CASCADE;`)

	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Status() (models.NumRecords, error) {
	var numRec models.NumRecords
	tx, err := r.db.Beginx()
	if err != nil {
		return models.NumRecords{}, err
	}
	err = tx.Get(&numRec.Post, "SELECT COUNT(*) as post_count FROM dbforum.post")
	err = tx.Get(&numRec.User, "SELECT COUNT(*) as user_count FROM dbforum.users")
	err = tx.Get(&numRec.Forum, "SELECT COUNT(*) as forum_count FROM dbforum.forum")
	err = tx.Get(&numRec.Thread, "SELECT COUNT(*) as thread_count FROM dbforum.thread")
	if err != nil {
		_ = tx.Rollback()
		return models.NumRecords{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.NumRecords{}, err
	}
	return numRec, nil
}
