package models

import "gorm.io/gorm"

type Chat_history struct {
	gorm.Model
	UserID  uint
	Content string `form:"content" validate:"required,excludesall= "`
	RoomID  uint
	Rooms   Rooms `gorm:"foreignKey:RoomID"`
}

type Req_reseiver struct {
	Content string `json:"content"`
	RoomID  uint   `json:"room_id"`
}
