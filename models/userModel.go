package models

import "gorm.io/gorm"

type Users struct {
	gorm.Model
	UserName string `gorm:"unique"`
	Password string
}
