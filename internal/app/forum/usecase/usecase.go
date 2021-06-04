package usecase

import (
	forumRepo "DBForum/internal/app/forum/repository"
	"DBForum/internal/app/models"
	threadRepo "DBForum/internal/app/thread/repository"
	userRepo "DBForum/internal/app/user/repository"
)

type UseCase struct {
	forumRepo  forumRepo.Repository
	userRepo   userRepo.Repository
	threadRepo threadRepo.Repository
}

func NewUseCase(forumRepo forumRepo.Repository, userRepo userRepo.Repository, threadRepo threadRepo.Repository) *UseCase {
	return &UseCase{
		forumRepo:  forumRepo,
		userRepo:   userRepo,
		threadRepo: threadRepo,
	}
}

func (u *UseCase) CreateForum(forum *models.Forum) (*models.Forum, error) {
	err := u.forumRepo.CreateForum(forum)
	if err != nil {
		return forum, err
	}
	return forum, nil
}

func (u *UseCase) GetInfoBySlug(slug string) (*models.Forum, error) {
	forum, err := u.forumRepo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}
	return forum, nil
}

func (u *UseCase) CreateThread(thread *models.Thread) (*models.Thread, error) {
	thread, err := u.threadRepo.CreateThread(thread)
	if err != nil {
		return thread, err
	}
	return thread, nil
}

func (u *UseCase) GetForumUsers(forumSlug string, limit int, since string, desc bool) ([]models.User, error) {
	if limit == 0 {
		limit = 100
	}
	users, err := u.userRepo.GetForumUsers(forumSlug, limit, since, desc)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return []models.User{}, nil
	}
	return users, nil
}

func (u *UseCase) GetForumThreads(forumSlug string, limit int, since string, desc bool) ([]models.Thread, error) {
	threads, err := u.threadRepo.GetForumThreads(forumSlug, limit, since, desc)
	if err != nil {
		return nil, err
	}
	if threads == nil {
		return []models.Thread{}, nil
	}
	return threads, nil
}
