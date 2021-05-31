package usecase

import (
	"DBForum/internal/app/models"
	postRepo "DBForum/internal/app/post/repository"
	threadRepo "DBForum/internal/app/thread/repository"
	"github.com/go-openapi/strfmt"
	"strconv"
	"time"
)

type UseCase struct {
	threadRepo threadRepo.Repository
	postRepo   postRepo.Repository
}

func NewUseCase(threadRepo threadRepo.Repository, postRepo postRepo.Repository) *UseCase {
	return &UseCase{
		threadRepo: threadRepo,
		postRepo:   postRepo,
	}
}

func (u *UseCase) ThreadInfo(idOrSlug string) (*models.Thread, error) {
	var id uint64
	var err error
	if id, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		thread, err := u.threadRepo.FindThreadBySlug(idOrSlug)
		if err != nil {
			return nil, err
		}
		return thread, nil
	}
	thread, err := u.threadRepo.FindThreadByID(id)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (u *UseCase) ChangeThread(idOrSlug string, thread models.Thread) (models.Thread, error) {
	var id uint64
	var err error
	if id, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		thread, err = u.threadRepo.UpdateThreadBySlug(idOrSlug, thread)
		if err != nil {
			return models.Thread{}, err
		}
		return thread, nil
	}
	thread, err = u.threadRepo.UpdateThreadByID(id, thread)
	if err != nil {
		return models.Thread{}, err
	}
	return thread, nil
}

func (u *UseCase) VoteThread(idOrSlug string, vote models.Vote) (models.Thread, error) {
	var id uint64
	var err error
	if id, err = strconv.ParseUint(idOrSlug, 10, 64); err != nil {
		thread, err := u.threadRepo.VoteThreadBySlug(idOrSlug, vote)
		if err != nil {
			return models.Thread{}, err
		}
		return thread, nil
	}
	thread, err := u.threadRepo.VoteThreadByID(id, vote)
	if err != nil {
		return models.Thread{}, err
	}
	return thread, nil
}

func (u *UseCase) CreatePosts(idOrSlug string, posts []models.Post) ([]models.Post, error) {
	created := strfmt.DateTime(time.Now())
	result := make([]models.Post, 0, len(posts))
	if len(posts) == 0 {
		posts = append(posts, models.Post{Forum: idOrSlug})
	}
	for _, post := range posts {
		post.Created = time.Time(created)
		post, err := u.postRepo.CreatePost(idOrSlug, post)
		if err != nil {
			return nil, err
		}
		if post.Author != "" {
			result = append(result, post)
		}
	}
	return result, nil
}

func (u *UseCase) GetPosts(idOrSlug string, limit int64, since int64, sort string, desc bool) ([]models.Post, error) {
	posts, err := u.postRepo.GetPosts(idOrSlug, limit, since, desc, sort)
	if err != nil {
		return nil, err
	}
	if posts == nil {
		return []models.Post{}, nil
	}
	return posts, nil
}
