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

func (s *Store) GetUserById(id uint) (*models.User, error) {
	var user models.User
	result := s.DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (s *Store) CreateTask(u *models.User, t models.Task) (*models.Task, error) {
	result := s.DB.Create(&t)
	if result.Error != nil {
		return nil, result.Error
	}

	if err := s.DB.Model(u).Association("Tasks").Append(&t).Error; err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *Store) GetTasks(u *models.User, params map[string]interface{}) (*[]models.Task, error) {
	var tasks []models.Task
	if err := s.DB.Where(params).Model(u).Association("Tasks").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return &tasks, nil
}

func (s *Store) GetTask(u *models.User, id int) (*models.Task, error) {
	var task models.Task
	if err := s.DB.Model(u).Preload("Users").Where("ID = ?", id).Association("Tasks").Find(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) AddUserToTask(u *models.User, t models.Task, idUser, idTask int) (*models.User, *models.Task, error) {
	if err := s.DB.First(u, idUser).Error; err != nil {
		return nil, nil, err
	}
	if err := s.DB.First(&t, idTask).Error; err != nil {
		return nil, nil, err
	}

	if err := s.DB.Model(&t).Association("Users").Append(u).Error; err != nil {
		return nil, nil, err
	}

	return u, &t, nil
}

func (s *Store) RemoveUserFromTask(u *models.User, t models.Task, idUser, idTask int) (*models.User, *models.Task, error) {
	if err := s.DB.First(u, idUser).Error; err != nil {
		return nil, nil, err
	}
	if err := s.DB.First(&t, idTask).Error; err != nil {
		return nil, nil, err
	}

	if err := s.DB.Model(&t).Association("Users").Delete(u).Error; err != nil {
		return nil, nil, err
	}

	return u, &t, nil
}

func (s *Store) UpdateTask(u *models.User, t models.UpdateTask, idTask int) (*models.Task, error) {
	var task models.Task
	if err := s.DB.Model(u).Where("ID = ?", idTask).Association("Tasks").Find(&task).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Model(&task).Update(t).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *Store) DeleteTask(u *models.User, idTask int) (*models.Task, error) {
	var task models.Task
	if err := s.DB.Model(u).Where("ID = ?", idTask).Association("Tasks").Find(&task).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Delete(task).Error; err != nil {
		return nil, err
	}

	return &task, nil
}
