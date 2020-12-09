package models

import (
	"database/sql/driver"

	"github.com/jinzhu/gorm"
)

type priority string

const (
	low  priority = "1"
	mid  priority = "2"
	high priority = "3"
)

func (p *priority) Scan(value interface{}) error {
	*p = priority(value.([]byte))
	return nil
}

func (p priority) Value() (driver.Value, error) {
	return string(p), nil
}

type User struct {
	gorm.Model `json:"-"`
	Username   string  `gorm:"type:varchar(50);not null;unique" json:"username" validate:"required,gte=3"`
	Password   string  `gorm:"not null" json:"password" validate:"required"`
	Tasks      []*Task `gorm:"many2many:user_tasks;" json:"-"`
}

// Avoid returning Password
type returnedUser struct {
	Username string `json:"username"`
}

func (u *User) ReturnUser() *returnedUser {
	return &returnedUser{
		Username: u.Username,
	}
}

type Task struct {
	gorm.Model  `json:"-"`
	Title       string `gorm:"type:varchar(50);not null" json:"title" validate:"required"`
	Description string `gorm:"type:varchar(200);not null" json:"description" validate:"required,lte=200"`

	// ENUM not supported in postgres
	// Priority  string `gorm:"type:ENUM(1', '2', '3');default:'1'" json:"priority"`
	Priority  priority `sql:"type:priority" gorm:"default:'1'" json:"priority"`
	Completed bool     `json:"completed"`
}

func (t *Task) Complete() {
	t.Completed = true
}

func (t *Task) Undo() {
	t.Completed = false
}
