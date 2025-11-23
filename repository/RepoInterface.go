//the interface package defines the functions other services can use

package repository

import (
	"Users/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	DeleteUser(user *models.User) error
	FetchAllUsers() ([]models.User, error)
	FetchUserByID(id string) (*models.User, error)
	UpdateUserStatus(id string, status string) error
}
