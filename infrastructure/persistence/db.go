package persistence

import (
	"fmt"
	"learning-golang-ddd/domain/entity"
	"learning-golang-ddd/domain/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repositories struct {
	User repository.UserRepository
	Food repository.FoodRepository
	db   *gorm.DB
}

func NewRepositories(dbUser, dbPassword, dbPort, dbHost, dbName string) (*Repositories, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, dbPort, dbUser, dbName, dbPassword)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.Logger.LogMode(logger.LogLevel(4))

	return &Repositories{
		User: NewUserRepository(db),
		Food: NewFoodRepository(db),
		db:   db,
	}, nil
}

// closes the database connection
func (r *Repositories) Close() error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	db.Close()
	return nil
}

// this migrate all tables
func (r *Repositories) Automigrate() error {
	return r.db.AutoMigrate(&entity.User{}, &entity.Food{})
}
