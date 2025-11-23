package services

import (
	"Users/models"
	"Users/repository"
	"Users/utils"
	"Users/validation"
)

type createServiceImpl struct {
	createUser repository.UserRepository
}

func NewCreateService(createUser repository.UserRepository) CreateInterface {
	return &createServiceImpl{createUser: createUser}

}

func (c *createServiceImpl) CreateUser(user *models.User) error {
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
	//hashing password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user = &models.User{
		Username: user.Username,
		Email:    user.Email,
		Password: hashedPassword,
	}
	//calling the repository layer to create user
	err = c.createUser.CreateUser(user)
	if err != nil {
		return err

	}
	return nil
}
