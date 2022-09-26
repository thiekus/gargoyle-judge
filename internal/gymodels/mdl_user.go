package gymodels

import (
	"gorm.io/gorm"
)

type User struct {
	Id uint `gorm:"autoIncrement; primaryKey; unique; not null"`
}

type UserModel struct {
	db *gorm.DB
}
