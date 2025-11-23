package services

import (
	"Users/models"
	"Users/repository"
	"Users/validation"
)

type updateServiceImpl struct {
	update repository.UserRepository
}

func NewUpdateService(update repository.UserRepository) UpdateInterface {
	return &updateServiceImpl{update: update}
}

func (u updateServiceImpl) UpdateUser(user *models.User) error {

	//validating email
	if user.Email != "" {
		err := validation.ValidateEmail(user.Email)
		if err != nil {
			return err
		}
	}
	if user.Password != "" {
		err := validation.ValidatePassword(user.Password)
		if err != nil {
			return err
		}
	}
	if user.Username != "" {
		err := validation.ValidateUsername(user.Username)
		if err != nil {
			return err
		}
	}
	//calling the repository layer to update user
	err := u.update.UpdateUser(user)
	if err != nil {
		return err
	}
	return nil
}
