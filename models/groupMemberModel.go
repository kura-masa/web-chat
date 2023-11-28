package models

import (
	"gorm.io/gorm"
)

type Groups struct {
	gorm.Model
	UsersRefer uint
	Users      Users `gorm:"foreignKey:UsersRefer"`
	RoomsRefer uint
	Rooms      Rooms `gorm:"foreignKey:RoomsRefer"`
}
