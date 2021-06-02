package usecase

import (
	"DBForum/internal/app/models"
	postRepo "DBForum/internal/app/post/repository"
	threadRepo "DBForum/internal/app/thread/repository"
	"strconv"
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
	thread, err := u.threadRepo.VoteThreadByID(idOrSlug, vote)
	if err != nil {
		return models.Thread{}, err
	}
	return thread, nil
}

func (u *UseCase) CreatePosts(idOrSlug string, posts []models.Post) ([]models.Post, error) {
	posts, err := u.postRepo.CreatePosts(idOrSlug, posts)
	if err != nil {
		return nil, err
	}
	if posts == nil {
		return []models.Post{}, err
	}
	return posts, nil
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
