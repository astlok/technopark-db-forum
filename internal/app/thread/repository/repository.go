package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"github.com/jackc/pgx"
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
                                   NULLIF($5,''), 
                                   $6) RETURNING ID`

	selectThreadBySlug = "SELECT id, forum_slug, author_nickname, title, message, votes,  COALESCE(slug, '') as slug, created FROM dbforum.thread WHERE slug = $1"

	selectThreadsByForumSlugSinceDesc = "SELECT id, forum_slug, author_nickname, title, message, votes, COALESCE(slug, '') as slug, created FROM dbforum.thread WHERE forum_slug = $1 AND created <= $2 ORDER BY created DESC LIMIT $3"

	selectThreadsByForumSlugSince = "SELECT id, forum_slug, author_nickname, title, message, votes, COALESCE(slug, '') as slug, created FROM dbforum.thread WHERE forum_slug = $1 AND created >= $2 ORDER BY created LIMIT $3"

	selectThreadsByForumSlugDesc = "SELECT id, forum_slug, author_nickname, title, message, votes, COALESCE(slug, '') as slug, created FROM dbforum.thread WHERE forum_slug = $1 ORDER BY created DESC LIMIT $2"

	selectThreadsByForumSlug = "SELECT id, forum_slug, author_nickname, title, message, votes, COALESCE(slug, '') as slug, created FROM dbforum.thread WHERE forum_slug = $1 ORDER BY created LIMIT $2"

	selectThreadByID = "SELECT id, forum_slug, author_nickname, title, message, votes, COALESCE(slug,'') as slug, created from dbforum.thread WHERE id = $1"

	updateThreadBySlug = "UPDATE dbforum.thread SET title=$1, message=$2 WHERE slug=$3"

	updateThreadByID = "UPDATE dbforum.thread SET title=$1, message=$2 WHERE id=$3"

	selectVoteInfo = "SELECT nickname, voice FROM dbforum.votes WHERE thread_id = $1 AND nickname = $2"

	updateThreadVoteBySlug = "UPDATE dbforum.thread SET votes=$1 WHERE slug=$2"

	updateThreadVoteByID = "UPDATE dbforum.thread SET votes=$1 WHERE id=$2"

	intertVote = "INSERT INTO dbforum.votes(nickname, voice, thread_id) VALUES ($1, $2, $3)"

	updateUserVote = "UPDATE dbforum.votes SET voice=$1 WHERE thread_id = $2 AND nickname = $3"

	selectSlugBySlug = "SELECT slug  as slug FROM dbforum.forum WHERE slug = $1"

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

func (r *Repository) CreateThread(thread *models.Thread) (*models.Thread, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(selectThreadBySlug, thread.Slug)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if rows.Next() {
		err = rows.Scan(
			&thread.ID,
			&thread.Forum,
			&thread.Author,
			&thread.Title,
			&thread.Message,
			&thread.Votes,
			&thread.Slug,
			&thread.Created)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		_ = tx.Rollback()
		rows.Close()
		return thread, customErr.ErrDuplicate
	}
	rows.Close()

	var slug string
	rows, err = tx.Query(selectSlugBySlug, thread.Forum)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if !rows.Next() {
		_ = tx.Rollback()
		return nil, customErr.ErrForumNotFound
	}
	err = rows.Scan(&slug)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	rows.Close()
	thread.Forum = slug

	var nickname string
	rows, err = tx.Query(selectNicknameByNickname, thread.Author)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if !rows.Next() {
		_ = tx.Rollback()
		return nil, customErr.ErrUserNotFound
	}
	err = rows.Scan(&nickname)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	rows.Close()

	thread.Author = nickname

	err = tx.QueryRow(
		insertThread,
		thread.Forum,
		thread.Author,
		thread.Title,
		thread.Message,
		thread.Slug,
		thread.Created).Scan(&thread.ID)

	if driverErr, ok := err.(pgx.PgError); ok {
		if driverErr.Code == "23505" {
			_ = tx.Rollback()
			return thread, customErr.ErrDuplicate
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	return thread, nil
}

func (r *Repository) FindThreadBySlug(threadSlug string) (*models.Thread, error) {
	thread := models.Thread{}
	rows, err := r.db.Query(selectThreadBySlug, threadSlug)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, customErr.ErrForumNotFound
	}
	err = rows.Scan(
		&thread.ID,
		&thread.Forum,
		&thread.Author,
		&thread.Title,
		&thread.Message,
		&thread.Votes,
		&thread.Slug,
		&thread.Created)
	if err != nil {
		return nil, err
	}
	rows.Close()
	return &thread, nil
}

func (r *Repository) FindThreadByID(id uint64) (*models.Thread, error) {
	thread := models.Thread{}
	rows, err := r.db.Query(selectThreadByID, id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, customErr.ErrForumNotFound
	}
	err = rows.Scan(
		&thread.ID,
		&thread.Forum,
		&thread.Author,
		&thread.Title,
		&thread.Message,
		&thread.Votes,
		&thread.Slug,
		&thread.Created)
	if err != nil {
		return nil, err
	}
	rows.Close()
	return &thread, nil
}

func (r *Repository) GetForumThreads(forumSlug string, limit int64, since string, desc bool) ([]models.Thread, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	var threads []models.Thread
	row, err := tx.Query("SELECT 1 FROM dbforum.forum WHERE slug = $1 LIMIT 1", forumSlug)
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
			row, err = tx.Query(selectThreadsByForumSlugDesc, forumSlug, limit)
		} else {
			row, err = tx.Query(selectThreadsByForumSlug, forumSlug, limit)
		}
	} else {
		if desc {
			row, err = tx.Query(selectThreadsByForumSlugSinceDesc, forumSlug, since, limit)
		} else {
			row, err = tx.Query(selectThreadsByForumSlugSince, forumSlug, since, limit)
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	for row.Next() {
		th := models.Thread{}
		err := row.Scan(
			&th.ID,
			&th.Forum,
			&th.Author,
			&th.Title,
			&th.Message,
			&th.Votes,
			&th.Slug,
			&th.Created)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		threads = append(threads, th)
	}
	row.Close()
	if threads == nil {
		_ = tx.Rollback()
		return nil, nil
	}
	_ = tx.Commit()
	return threads, nil
}

func (r *Repository) UpdateThreadBySlug(threadSlug string, thread models.Thread) (models.Thread, error) {
	var oldThread models.Thread
	tx, err := r.db.Begin()
	if err != nil {
		return models.Thread{}, err
	}
	rows, err := tx.Query(selectThreadBySlug, threadSlug)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if rows.Next() {
		err = rows.Scan(
			&oldThread.ID,
			&oldThread.Forum,
			&oldThread.Author,
			&oldThread.Title,
			&oldThread.Message,
			&oldThread.Votes,
			&oldThread.Slug,
			&oldThread.Created)
	} else {
		_ = tx.Rollback()
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	rows.Close()

	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if thread.Message != "" {
		oldThread.Message = thread.Message
	}
	if thread.Title != "" {
		oldThread.Title = thread.Title
	}
	_, err = tx.Exec(updateThreadBySlug, &oldThread.Title, &oldThread.Message, &oldThread.Slug)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	return oldThread, nil
}

func (r *Repository) UpdateThreadByID(threadID uint64, thread models.Thread) (models.Thread, error) {
	var oldThread models.Thread
	tx, err := r.db.Begin()
	if err != nil {
		return models.Thread{}, err
	}
	rows, err := tx.Query(selectThreadByID, threadID)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if !rows.Next() {
		_ = tx.Rollback()
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	err = rows.Scan(
		&oldThread.ID,
		&oldThread.Forum,
		&oldThread.Author,
		&oldThread.Title,
		&oldThread.Message,
		&oldThread.Votes,
		&oldThread.Slug,
		&oldThread.Created)
	rows.Close()
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if thread.Message != "" {
		oldThread.Message = thread.Message
	}
	if thread.Title != "" {
		oldThread.Title = thread.Title
	}
	_, err = tx.Exec(updateThreadByID, &oldThread.Title, &oldThread.Message, &oldThread.ID)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	return oldThread, nil
}

func (r *Repository) VoteThreadBySlug(slug string, vote models.Vote) (models.Thread, error) {
	var thread models.Thread
	tx, err := r.db.Begin()
	if err != nil {
		return models.Thread{}, err
	}
	rows, err := tx.Query(selectThreadBySlug, slug)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if !rows.Next() {
		_ = tx.Rollback()
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	err = rows.Scan(
		&thread.ID,
		&thread.Forum,
		&thread.Author,
		&thread.Title,
		&thread.Message,
		&thread.Votes,
		&thread.Slug,
		&thread.Created)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	rows.Close()
	row, err := tx.Query("SELECT 1 FROM dbforum.users WHERE nickname=$1 LIMIT 1", vote.Nickname)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if !row.Next() {
		_ = tx.Rollback()
		return models.Thread{}, customErr.ErrUserNotFound
	}
	row.Close()
	curVote := models.Vote{}
	rows, err = tx.Query(selectVoteInfo, thread.ID, vote.Nickname)
	if err != nil {

		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if !rows.Next() {
		thread.Votes += vote.Voice
		_, err = tx.Exec(updateThreadVoteBySlug, thread.Votes, slug)
		_, err = tx.Exec(intertVote, vote.Nickname, vote.Voice, thread.ID)
		if err != nil {
			rows.Close()
			_ = tx.Rollback()
			return models.Thread{}, err
		}
		if err := tx.Commit(); err != nil {
			rows.Close()
			_ = tx.Rollback()
			return models.Thread{}, err
		}
		rows.Close()
		return thread, nil
	}
	err = rows.Scan(
		&curVote.Nickname,
		&curVote.Voice)
	rows.Close()
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	thread.Votes -= curVote.Voice
	thread.Votes += vote.Voice
	_, err = tx.Exec(updateThreadVoteBySlug, thread.Votes, slug)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
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
	tx, err := r.db.Begin()
	if err != nil {
		return models.Thread{}, err
	}
	rows, err := tx.Query(selectThreadByID, id)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if !rows.Next() {
		_ = tx.Rollback()
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	err = rows.Scan(
		&thread.ID,
		&thread.Forum,
		&thread.Author,
		&thread.Title,
		&thread.Message,
		&thread.Votes,
		&thread.Slug,
		&thread.Created)
	rows.Close()
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	row, err := tx.Query("SELECT 1 FROM dbforum.users WHERE nickname=$1 LIMIT 1", vote.Nickname)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if !row.Next() {
		_ = tx.Rollback()
		return models.Thread{}, customErr.ErrUserNotFound
	}
	row.Close()
	curVote := models.Vote{}
	rows, err = tx.Query(selectVoteInfo, thread.ID, vote.Nickname)
	if err != nil {

		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if !rows.Next() {
		thread.Votes += vote.Voice
		_, err = tx.Exec(updateThreadVoteByID, thread.Votes, id)
		if err != nil {
			_ = tx.Rollback()
			return models.Thread{}, err
		}
		_, err = tx.Exec(intertVote, vote.Nickname, vote.Voice, thread.ID)
		if err != nil {
			_ = tx.Rollback()
			return models.Thread{}, err
		}
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return models.Thread{}, err
		}
		rows.Close()
		return thread, nil
	}
	err = rows.Scan(
		&curVote.Nickname,
		&curVote.Voice)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	rows.Close()
	if curVote.Voice == vote.Voice {
		_ = tx.Rollback()
		return thread, nil
	}
	thread.Votes -= curVote.Voice
	thread.Votes += vote.Voice
	_, err = tx.Exec(updateThreadVoteByID, thread.Votes, id)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
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
