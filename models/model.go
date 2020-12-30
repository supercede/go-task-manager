package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
)

type priority string
type password string

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
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" sql:"index"`
	Username  string     `gorm:"type:varchar(50);not null;unique" json:"username" validate:"required,gte=3"`
	Password  string     `gorm:"not null" json:"password" validate:"required"`
	Tasks     []*Task    `gorm:"many2many:user_tasks;" json:"-"`
}

// Avoid returning Password
func (u User) MarshalJSON() ([]byte, error) {
	var tmp struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
	}
	tmp.Username = u.Username
	tmp.ID = u.ID
	return json.Marshal(&tmp)
}

type Task struct {
	gorm.Model
	Title       string `gorm:"type:varchar(50);not null" json:"title" validate:"required"`
	Description string `gorm:"type:varchar(200);not null" json:"description" validate:"required,lte=200"`

	// ENUM not supported in postgres
	// Priority  string `gorm:"type:ENUM(1', '2', '3');default:'1'" json:"priority"`
	Priority  priority `sql:"type:priority" gorm:"default:'1'" json:"priority"`
	Completed bool     `json:"completed" gorm:"default:'false'"`
	Users     []*User  `gorm:"many2many:user_tasks;" json:"users,omitempty"`
}

type UpdateTask struct {
	Priority  string `json:"priority" validate:"omitempty,oneof=1 2 3"`
	Completed bool   `json:"completed" validate:"omitempty"`
}
