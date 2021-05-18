package usecase

import (
	"DBForum/internal/app/models"
	userRepo "DBForum/internal/app/user/repository"
)

type UseCase struct {
	repo userRepo.Repository
}

func NewUseCase(repo userRepo.Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

func (u *UseCase) CreateUser(user models.User) error {
	err := u.repo.CreateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (u *UseCase) GetUsersByNickAndEmail(nickname string, email string) ([]models.User, error) {
	users, err := u.repo.GetUsersByNickAndEmail(nickname, email)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UseCase) GetUserInfo(nickname string) (*models.User, error) {
	user, err := u.repo.GetUserByNick(nickname)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UseCase) ChangeUser(user *models.User) error {
	err := u.repo.ChangeUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (u *UseCase) GetUserNickByEmail(email string) (string, error) {
	nickname, err := u.repo.GetUserNickByEmail(email)
	if err != nil {
		return "", err
	}
	return nickname, nil
}
