package controllers

import (
	"log"
	"net/http"
	"strconv"
	"web-chat/initializers"
	"web-chat/models"
	"web-chat/websocket"

	"github.com/gin-gonic/gin"
)

// チャットをデータベースに登録
func CreateChat(c *gin.Context) {
	user, _ := c.Get("user")
	content := &models.Req_reseiver{}
	if err := c.Bind(content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// チャットデータを構造体に格納
	chat := &models.Chat_history{
		Content: content.Content,
		RoomID:  content.RoomID,
		UserID:  user.(models.Users).ID,
	}
	// chat_historyテーブルにデータを挿入
	result := initializers.DB.Create(&chat)
	// エラー処理
	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "chat.html", gin.H{
			"title":  user.(models.Users).UserName,
			"result": "エラーが発生しました",
		})
		return
	}
	c.JSON(http.StatusOK, chat)
}

// チャットリストをフロントに返す
func ChatList(c *gin.Context) {
	user, _ := c.Get("user")

	var chatroom []models.Friends
	var room1 []models.Rooms
	var grouproom []models.Groups
	var room2 []models.Rooms

	initializers.DB.Where("user_id1 = ? OR user_id2 = ?", user.(models.Users).ID, user.(models.Users).ID).Find(&chatroom)
	initializers.DB.Where("users_refer = ?", user.(models.Users).ID).Find(&grouproom)

	// 配列(スライス)の要素毎のループ処理
	// for インデックス番号, 値の変数 := range 配列orスライス{繰り返し処理}
	for _, v := range chatroom {
		var room models.Rooms
		initializers.DB.Where("id = ?", v.RoomsRefer).Find(&room)
		room1 = append(room1, room) // append(スライス,追加する値)
	}

	for _, v := range grouproom {
		var room models.Rooms
		initializers.DB.Where("id = ?", v.RoomsRefer).Find(&room)
		//変更部分
		if room.RoomName != "" {
			room2 = append(room2, room)
		}
	}

	c.HTML(http.StatusOK, "chatlist.html", gin.H{
		"title": user.(models.Users).UserName + "'s Chat",
		"room1": room1,
		"room2": room2,
	})
}

// ルームのチャット履歴を表示（フロントに返す）show chat history at a room
func ListChatHistory(c *gin.Context) {
	user, _ := c.Get("user")
	userID := user.(models.Users).ID
	var chat_history []models.Chat_history
	var room models.Rooms
	room_id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		log.Fatalln(err)
	}
	// 部外者がroomには入れないようにする
	var frooms []models.Friends
	var grooms []models.Groups
	initializers.DB.Where("user_id1 = ? OR user_id2 = ?", userID, userID).Find(&frooms)
	initializers.DB.Where("users_refer = ?", userID).Find(&grooms)
	j := false
	for _, froom := range frooms {
		if froom.RoomsRefer == uint(room_id) {
			j = true
			break
		}
	}
	for _, groom := range grooms {
		if groom.RoomsRefer == uint(room_id) {
			j = true
			break
		}
	}
	if !j {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "無効なリクエストです",
		})
		return
	}

	// Hubを作成
	if models.RoomToHub[uint(room_id)] == nil {
		h := websocket.NewHub()
		go h.Run()
		models.RoomToHub[uint(room_id)] = h
	}
	// initializers.DB.Find(&chat_history, "room_id = ?", user.(models.Chat_history).ID)
	initializers.DB.Where("room_id = ?", room_id).Find(&chat_history)
	initializers.DB.Where("id = ?", room_id).Find(&room)
	// fmt.Print(chat_history)
	c.HTML(http.StatusOK, "chat.html", gin.H{
		"title":        room.RoomName + "'s Chat History",
		"chat_history": chat_history,
		"room":         room,
	})
}
