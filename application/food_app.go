package application

import (
	"learning-golang-ddd/domain/entity"
	"learning-golang-ddd/domain/repository"
)

type foodApp struct {
	fr repository.FoodRepository
}

var _ FoodAppInterface = &foodApp{}

type FoodAppInterface interface {
	SaveFood(*entity.Food) (*entity.Food, map[string]string)
	GetAllFood() ([]entity.Food, error)
	GetFood(uint64) (*entity.Food, error)
	UpdateFood(*entity.Food) (*entity.Food, map[string]string)
	DeleteFood(uint64) error
}

func (fApp *foodApp) SaveFood(food *entity.Food) (*entity.Food, map[string]string) {
	return fApp.fr.SaveFood(food)
}

func (fApp *foodApp) GetAllFood() ([]entity.Food, error) {
	return fApp.fr.GetAllFood()
}

func (fApp *foodApp) GetFood(foodId uint64) (*entity.Food, error) {
	return fApp.fr.GetFood(foodId)
}

func (fApp *foodApp) UpdateFood(food *entity.Food) (*entity.Food, map[string]string) {
	return fApp.fr.UpdateFood(food)
}

func (fApp *foodApp) DeleteFood(foodId uint64) error {
	return fApp.fr.DeleteFood(foodId)
}
