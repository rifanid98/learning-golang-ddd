package repository

import "learning-golang-ddd/domain/entity"

type UserRepository interface {
	SaveUser(*entity.User) (*entity.User, map[string]string)
	GetUser(uint64) (*entity.User, error)
	GetUsers() ([]entity.User, error)
	GetUserByFilter(*entity.User) (*entity.User, map[string]string)
}
