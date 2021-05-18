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

func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (u *UseCase) GetPostInfoByID(id uint64, related []string) (models.PostInfo, error) {
	postInfo := &models.PostInfo{}
	var err error
	postInfo.Post, err = u.postRepo.GetPostByID(id)
	if err != nil {
		return models.PostInfo{}, err
	}

	if Find(related, "user") {
		postInfo.Author, err = u.userRepo.GetUserByNick(postInfo.Post.Author)
		if err != nil {
			return models.PostInfo{}, err
		}
	}

	if Find(related, "thread") {
		postInfo.Thread, err = u.threadRepo.FindThreadByID(postInfo.Post.Thread)
		if err != nil {
			return models.PostInfo{}, err
		}
	}

	if Find(related, "forum") {
		postInfo.Forum, err = u.forumRepo.FindBySlug(postInfo.Post.Forum)
		if err != nil {
			return models.PostInfo{}, err
		}
	}
	return *postInfo, nil
}

func (u *UseCase) ChangeMessage(post models.Post) (*models.Post, error) {
	oldPost, err := u.postRepo.GetPostByID(post.ID)
	if err != nil {
		return nil, err
	}
	oldPost.Message = post.Message
	oldPost.IsEdited = true
	err = u.postRepo.ChangePost(oldPost)
	if err != nil {
		return nil, err
	}
	return oldPost, nil
}
