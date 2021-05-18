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
	insertThread = `INSERT INTO dbforum.thread(
							   forum_slug, 
							   author_nickname, 
							   title, 
							   message, 
							   slug, 
							   created
                           ) 
                           VALUES (
                                   $1, 
                                   $2, 
                                   $3, 
                                   $4, 
                                   $5, 
                                   $6) RETURNING ID`

	selectThreadBySlug = "SELECT * FROM dbforum.thread WHERE slug = $1"

	selectThreadsByForumSlugDesc = "SELECT * FROM dbforum.thread WHERE forum_slug = $1 AND CASE WHEN $2 != '' THEN created <= $2 ELSE TRUE END ORDER BY created DESC LIMIT $3"

	selectThreadsByForumSlug = "SELECT * FROM dbforum.thread WHERE forum_slug = $1 AND CASE WHEN $2 != '' THEN created >= $2 ELSE TRUE END ORDER BY created LIMIT $3"

	selectThreadByID = "SELECT * from dbforum.thread WHERE id = $1"

	updateThreadBySlug = `UPDATE dbforum.thread SET 
							 title=:title,
							 message=:message
                         WHERE slug=:slug`

	updateThreadByID = `UPDATE dbforum.thread SET 
							 title=:title,
							 message=:message
                         WHERE id=:id`

	selectVoteInfo = "SELECT nickname, voice FROM dbforum.votes WHERE thread_id = $1 AND nickname = $2"

	updateThreadVoteBySlug = "UPDATE dbforum.thread SET votes=$1 WHERE slug=$2"

	updateThreadVoteByID = "UPDATE dbforum.thread SET votes=$1 WHERE id=$2"

	intertVote = "INSERT INTO dbforum.votes(nickname, voice, thread_id) VALUES ($1, $2, $3)"

	updateUserVote = "UPDATE dbforum.votes SET voice=$1 WHERE thread_id = $2 AND nickname = $3"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateThread(thread models.Thread) (uint64, error) {
	var threadID uint64
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	err = tx.QueryRowx(
		insertThread,
		thread.Forum,
		thread.Author,
		thread.Title,
		thread.Message,
		thread.Slug,
		thread.Created).Scan(&threadID)

	if driverErr, ok := err.(pgx.PgError); ok {
		if driverErr.Code == "23505" {
			_ = tx.Rollback()
			return 0, customErr.ErrDuplicate
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	//query := fmt.Sprintf("INSERT INTO dbforum.forum_users(forum_slug, nickname, fullname, about, email) "+
	//	"SELECT '%s', nickname, fullname, about, email FROM dbforum.users "+
	//	"WHERE nickname = '%s'", thread.Forum, thread.Author)
	//if _, err = tx.Exec(query); err != nil {
	//	if driverErr, ok := err.(pgx.PgError); ok {
	//		if driverErr.Code == "23505" {
	//			_ = tx.Rollback()
	//			err = r.db.QueryRowx(
	//				insertThread,
	//				thread.Forum,
	//				thread.Author,
	//				thread.Title,
	//				thread.Message,
	//				thread.Slug,
	//				thread.Created).Scan(&threadID)
	//			if err != nil {
	//				return 0, err
	//			}
	//			return threadID, nil
	//		}
	//	}
	//	_ = tx.Rollback()
	//	return 0, err
	//}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return threadID, nil
}

func (r *Repository) FindThreadBySlug(threadSlug string) (*models.Thread, error) {
	thread := models.Thread{}
	if err := r.db.Get(&thread, selectThreadBySlug, threadSlug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrForumNotFound
		}
		return nil, err
	}
	return &thread, nil
}

func (r *Repository) FindThreadByID(id uint64) (*models.Thread, error) {
	thread := models.Thread{}
	if err := r.db.Get(&thread, selectThreadByID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrForumNotFound
		}
		return nil, err
	}
	return &thread, nil
}

func (r *Repository) GetForumThreads(forumSlug string, limit int64, since string, desc bool) ([]models.Thread, error) {
	var threads []models.Thread
	var err error
	if desc {
		err = r.db.Select(&threads, selectThreadsByForumSlugDesc, forumSlug, since, limit)
	} else {
		err = r.db.Select(&threads, selectThreadsByForumSlug, forumSlug, since, limit)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, customErr.ErrThreadNotFound
	}
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (r *Repository) UpdateThreadBySlug(threadSlug string, thread models.Thread) (models.Thread, error) {
	var oldThread models.Thread
	tx, err := r.db.Beginx()
	if err != nil {
		return models.Thread{}, err
	}
	err = tx.Get(&oldThread, selectThreadBySlug, threadSlug)
	if err != nil {
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	if thread.Message != "" {
		oldThread.Message = thread.Message
	}
	if thread.Title != "" {
		oldThread.Title = thread.Title
	}
	_, err = tx.NamedExec(updateThreadBySlug, oldThread)
	if err != nil {
		return models.Thread{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Thread{}, err
	}
	return oldThread, nil
}

func (r *Repository) UpdateThreadByID(threadID uint64, thread models.Thread) (models.Thread, error) {
	var oldThread models.Thread
	tx, err := r.db.Beginx()
	if err != nil {
		return models.Thread{}, err
	}
	err = tx.Get(&oldThread, selectThreadByID, threadID)
	if err != nil {
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	if thread.Message != "" {
		oldThread.Message = thread.Message
	}
	if thread.Title != "" {
		oldThread.Title = thread.Title
	}
	_, err = tx.NamedExec(updateThreadByID, oldThread)
	if err != nil {
		return models.Thread{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Thread{}, err
	}
	return oldThread, nil
}

func (r *Repository) VoteThreadBySlug(slug string, vote models.Vote) (models.Thread, error) {
	var thread models.Thread
	tx, err := r.db.Beginx()
	if err != nil {
		return models.Thread{}, err
	}
	err = tx.Get(&thread, selectThreadBySlug, slug)
	if errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	curVote := models.Vote{}
	err = tx.Get(&curVote, selectVoteInfo, thread.ID, vote.Nickname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			thread.Votes += vote.Voice
			_, err = tx.Exec(updateThreadVoteBySlug, thread.Votes, slug)
			_, err = tx.Exec(intertVote, vote.Nickname, vote.Voice, thread.ID)
			if err != nil {
				_ = tx.Rollback()
				return models.Thread{}, err
			}
			if err := tx.Commit(); err != nil {
				_ = tx.Rollback()
				return models.Thread{}, err
			}
			return thread, nil
		}
		return models.Thread{}, err
	}
	thread.Votes -= curVote.Voice
	thread.Votes += vote.Voice
	_, err = tx.Exec(updateThreadVoteBySlug, thread.Votes, slug)
	_, err = tx.Exec(updateUserVote, vote.Voice, thread.ID, vote.Nickname)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	return thread, nil
}

func (r *Repository) VoteThreadByID(id uint64, vote models.Vote) (models.Thread, error) {
	var thread models.Thread
	tx, err := r.db.Beginx()
	if err != nil {
		return models.Thread{}, err
	}
	err = tx.Get(&thread, selectThreadByID, id)
	if errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	curVote := models.Vote{}
	err = tx.Get(&curVote, selectVoteInfo, thread.ID, vote.Nickname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			thread.Votes += vote.Voice
			_, err = tx.Exec(updateThreadVoteByID, thread.Votes, id)
			_, err = tx.Exec(intertVote, vote.Nickname, vote.Voice, thread.ID)
			if err != nil {
				_ = tx.Rollback()
				return models.Thread{}, err
			}
			if err := tx.Commit(); err != nil {
				_ = tx.Rollback()
				return models.Thread{}, err
			}
			return thread, nil
		}
		return models.Thread{}, err
	}
	if curVote.Voice == vote.Voice {
		return thread, nil
	}
	thread.Votes -= curVote.Voice
	thread.Votes += vote.Voice
	_, err = tx.Exec(updateThreadVoteByID, thread.Votes, id)
	_, err = tx.Exec(updateUserVote, vote.Voice, thread.ID, vote.Nickname)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	return thread, nil
}

