package usecase

import (
	"DBForum/internal/app/models"
	serviceRepo "DBForum/internal/app/service/repository"
)

type UseCase struct {
	repo serviceRepo.Repository
}

func NewUseCase(repo serviceRepo.Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

func (u *UseCase) ClearDB() error {
	err := u.repo.ClearDB()
	if err != nil {
		return err
	}
	return nil
}

func (u *UseCase) Status() (models.NumRecords, error) {
	numRecords, err := u.repo.Status()
	if err != nil {
		return models.NumRecords{}, err
	}
	return numRecords, nil
}
