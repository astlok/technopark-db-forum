package usecase

import (
	forumRepository "DBForum/internal/app/forum/repository"
	"DBForum/internal/app/models"
	postRepository "DBForum/internal/app/post/repository"
	threadRepository "DBForum/internal/app/thread/repository"
	userRepository "DBForum/internal/app/user/repository"
)

type UseCase struct {
	postRepo   postRepository.Repository
	userRepo   userRepository.Repository
	threadRepo threadRepository.Repository
	forumRepo  forumRepository.Repository
}

func NewUseCase(postRepo postRepository.Repository,
	userRepo userRepository.Repository,
	threadRepo threadRepository.Repository,
	forumRepo forumRepository.Repository) *UseCase {
	return &UseCase{
		postRepo:   postRepo,
		userRepo:   userRepo,
		threadRepo: threadRepo,
		forumRepo:  forumRepo,
	}
}

func (u *UseCase) GetPostInfoByID(id uint64, related []string) (models.PostInfo, error) {
	postInfo, err := u.postRepo.GetPostInfoByID(id, related)
	if err != nil {
		return models.PostInfo{}, err
	}
	return *postInfo, nil
}

func (u *UseCase) ChangeMessage(post models.Post) (*models.Post, error) {
	post, err := u.postRepo.ChangePost(&post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}
