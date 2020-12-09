package database

import (
	"fmt"
	"todo-app/models"
	"todo-app/util"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
)

type Store struct {
	DB *gorm.DB
}

func New() (*Store, error) {
	conf, err := util.GetConfig()
	if err != nil {
		return &Store{}, errors.Wrap(err, "Failed to Read config file")
	}

	credentials := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", conf.PGHost, conf.PGUsername, conf.PGName, conf.PGPassword)

	db, err := gorm.Open("postgres", credentials)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	// ENUM not working with GORM Postgres
	db.Raw("CREATE TYPE priority AS ENUM ('1', '2', '3')").Row()
	db.AutoMigrate(&models.User{}, &models.Task{})
	return &Store{
		DB: db,
	}, nil
}

func (s *Store) AddUser(u models.User) (*models.User, error) {
	result := s.DB.Create(&u)
	if result.Error != nil {
		return nil, result.Error
	}

	return &u, nil
}

func (s *Store) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := s.DB.Where("username=?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
