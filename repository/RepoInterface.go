//the interface package defines the functions other services can use

package repository

import (
	"Users/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	UpdateUser(user *models.User)error 
}
