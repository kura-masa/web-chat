package initializers

import "web-chat/models"

func SyncDatabase() {
	// DBの初期化。カラムの物理削除はできない。
	DB.AutoMigrate(&models.Users{}, &models.Rooms{}, &models.Friends{}, &models.Groups{}, &models.Chat_history{})
}
