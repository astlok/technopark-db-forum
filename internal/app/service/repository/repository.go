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

	_, err = tx.Exec("truncPost")
	_, err = tx.Exec("truncForumUsers")
	_, err = tx.Exec("truncThread")
	_, err = tx.Exec("truncVotes")
	_, err = tx.Exec("truncForum")
	_, err = tx.Exec("truncUsers")

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
	err = tx.QueryRow("countPost").Scan(&numRec.Post)
	err = tx.QueryRow("countUsers").Scan(&numRec.User)
	err = tx.QueryRow("countForum").Scan(&numRec.Forum)
	err = tx.QueryRow("countThread").Scan(&numRec.Thread)
	if err != nil {
		_ = tx.Rollback()
		return models.NumRecords{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.NumRecords{}, err
	}
	return numRec, nil
}

func (r *Repository) Prepare() error {
	_, err := r.db.Prepare("truncPost", `TRUNCATE dbforum.post CASCADE`)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("truncForumUsers", `TRUNCATE dbforum.forum_users CASCADE`)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("truncThread", `TRUNCATE dbforum.thread CASCADE`)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("truncVotes", `TRUNCATE dbforum.votes CASCADE`)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("truncForum", `TRUNCATE dbforum.forum CASCADE`)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("truncUsers", `TRUNCATE dbforum.users CASCADE`)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("countPost", "SELECT COUNT(*) as post_count FROM dbforum.post")
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("countUsers", "SELECT COUNT(*) as user_count FROM dbforum.users")
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("countForum", "SELECT COUNT(*) as forum_count FROM dbforum.forum")
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("countThread", "SELECT COUNT(*) as thread_count FROM dbforum.thread")
	if err != nil {
		return err
	}

	return nil
}
