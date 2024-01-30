package models

import (
	"gorm.io/gorm"
)

type Friends struct {
	gorm.Model
	UserID1    uint
	UserID2    uint
	RoomsRefer uint
	Rooms      Rooms `gorm:"foreignKey:RoomsRefer"`
}
