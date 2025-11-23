package services

import (
	"Users/models"
)

type UpdateInterface interface {
	UpdateUser(user *models.User) error
}
type CreateInterface interface {
	CreateUser(user *models.User) error
}
