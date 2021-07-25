package application

import (
	"learning-golang-ddd/domain/entity"
	"learning-golang-ddd/domain/repository"
)

type userApp struct {
	ur repository.UserRepository
}

// userApp implements the UserAppInterface
var _ UserAppInterface = &userApp{}

type UserAppInterface interface {
	SaveUser(*entity.User) (*entity.User, map[string]string)
	GetUsers() ([]entity.User, error)
	GetUser(uint64) (*entity.User, error)
	GetUserByFilter(*entity.User) (*entity.User, map[string]string)
}

func (uApp *userApp) SaveUser(user *entity.User) (*entity.User, map[string]string) {
	return uApp.ur.SaveUser(user)
}

func (uApp *userApp) GetUsers() ([]entity.User, error) {
	return uApp.ur.GetUsers()
}

func (uApp *userApp) GetUser(uId uint64) (*entity.User, error) {
	return uApp.ur.GetUser(uId)
}

func (uApp *userApp) GetUserByFilter(user *entity.User) (*entity.User, map[string]string) {
	return uApp.ur.GetUserByFilter(user)
}
