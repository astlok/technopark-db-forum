package repository

import (
	"DBForum/internal/app/models"
	"github.com/jackc/pgx"
)

type Repository struct {
	db *pgx.ConnPool
}

func NewRepo(db *pgx.ConnPool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) ClearDB() error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`TRUNCATE dbforum.post CASCADE`)
	_, err = tx.Exec(`TRUNCATE dbforum.forum_users CASCADE`)
	_, err = tx.Exec(`TRUNCATE dbforum.thread CASCADE`)
	_, err = tx.Exec(`TRUNCATE dbforum.votes CASCADE`)
	_, err = tx.Exec(`TRUNCATE dbforum.forum CASCADE`)
	_, err = tx.Exec(`TRUNCATE dbforum.users CASCADE`)

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
	tx, err := r.db.Begin()
	if err != nil {
		return models.NumRecords{}, err
	}
	err = tx.QueryRow("SELECT COUNT(*) as post_count FROM dbforum.post").Scan(&numRec.Post)
	err = tx.QueryRow("SELECT COUNT(*) as user_count FROM dbforum.users").Scan(&numRec.User)
	err = tx.QueryRow("SELECT COUNT(*) as forum_count FROM dbforum.forum").Scan(&numRec.Forum)
	err = tx.QueryRow("SELECT COUNT(*) as thread_count FROM dbforum.thread").Scan(&numRec.Thread)
	if err != nil {
		_ = tx.Rollback()
		return models.NumRecords{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.NumRecords{}, err
	}
	return numRec, nil
}
