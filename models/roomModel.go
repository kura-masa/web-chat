package models

import (
	"web-chat/websocket"

	"gorm.io/gorm"
)

type Rooms struct {
	gorm.Model //gormで定義されている構造体
	RoomName   string
}

var RoomToHub = map[uint]*websocket.Hub{}
