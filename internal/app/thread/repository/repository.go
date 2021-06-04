package repository

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/models"
	"github.com/jackc/pgx"
	"strconv"
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

	updateThreadBySlug = "UPDATE dbforum.thread SET title=COALESCE(NULLIF($1, ''), title), message=COALESCE(NULLIF($2, ''), message) WHERE slug=$3 RETURNING *"

	updateThreadByID = "UPDATE dbforum.thread SET title=COALESCE(NULLIF($1, ''), title), message=COALESCE(NULLIF($2, ''), message) WHERE id=$3 RETURNING *"

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

	rows, err := tx.Query("selectThreadBySlug", thread.Slug)
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
	rows, err = tx.Query("selectSlugBySlug", thread.Forum)
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
	rows, err = tx.Query("selectNicknameByNickname", thread.Author)
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
		"insertThread",
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
	rows, err := r.db.Query("selectThreadBySlug", threadSlug)
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
	rows, err := r.db.Query("selectThreadByID", id)
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

func (r *Repository) GetForumThreads(forumSlug string, limit int, since string, desc bool) ([]models.Thread, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	var threads []models.Thread
	row, err := tx.Query("checkForum", forumSlug)
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
			row, err = tx.Query("selectThreadsByForumSlugDesc", forumSlug, limit)
		} else {
			row, err = tx.Query("selectThreadsByForumSlug", forumSlug, limit)
		}
	} else {
		if desc {
			row, err = tx.Query("selectThreadsByForumSlugSinceDesc", forumSlug, since, limit)
		} else {
			row, err = tx.Query("selectThreadsByForumSlugSince", forumSlug, since, limit)
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
	tx, err := r.db.Begin()
	if err != nil {
		return models.Thread{}, err
	}
	err = tx.QueryRow("updateThreadBySlug", &thread.Title, &thread.Message, &threadSlug).Scan(
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
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	return thread, nil
}

func (r *Repository) UpdateThreadByID(threadID uint64, thread models.Thread) (models.Thread, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return models.Thread{}, err
	}
	err = tx.QueryRow("updateThreadByID", &thread.Title, &thread.Message, &threadID).Scan(
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
		return models.Thread{}, customErr.ErrThreadNotFound
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	return thread, nil
}

func (r *Repository) VoteThreadByID(idOrSlug string, vote models.Vote) (models.Thread, error) {
	var thread models.Thread
	tx, err := r.db.Begin()
	if err != nil {
		return models.Thread{}, err
	}
	var rows *pgx.Rows
	var id uint64
	if id, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		rows, err = tx.Query("selectThreadBySlug", idOrSlug)
	} else {
		rows, err = tx.Query("selectThreadByID", id)
	}
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
	curVote := models.Vote{}
	rows, err = tx.Query("selectVoteInfo", thread.ID, vote.Nickname)
	if err != nil {
		_ = tx.Rollback()
		return models.Thread{}, err
	}
	if !rows.Next() {
		thread.Votes += vote.Voice
		_, err = tx.Exec("intertVote", vote.Nickname, vote.Voice, thread.ID)
		if err != nil {
			_ = tx.Rollback()
			return models.Thread{}, customErr.ErrUserNotFound
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
	_, err = tx.Exec("updateUserVote", vote.Voice, thread.ID, vote.Nickname)
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

func (r *Repository) Prepare() error {
	_, err := r.db.Prepare("selectThreadBySlug", selectThreadBySlug)
	if err != nil {
		return err
	}
	_, err = r.db.Prepare("selectSlugBySlug", selectSlugBySlug)
	if err != nil {
		return err
	}
	_, err = r.db.Prepare("selectNicknameByNickname", selectNicknameByNickname)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("insertThread", insertThread)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("updateThreadBySlug", updateThreadBySlug)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("updateThreadByID", updateThreadByID)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectThreadByID", selectThreadByID)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectVoteInfo", selectVoteInfo)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("intertVote", intertVote)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("updateUserVote", updateUserVote)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectThreadsByForumSlugDesc", selectThreadsByForumSlugDesc)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectThreadsByForumSlug", selectThreadsByForumSlug)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectThreadsByForumSlugSinceDesc", selectThreadsByForumSlugSinceDesc)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("selectThreadsByForumSlugSince", selectThreadsByForumSlugSince)
	if err != nil {
		return err
	}

	_, err = r.db.Prepare("checkForum", "SELECT 1 FROM dbforum.forum WHERE slug = $1 LIMIT 1")
	if err != nil {
		return err
	}

	return nil
}
