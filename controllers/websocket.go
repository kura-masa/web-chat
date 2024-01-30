package controllers

import (
	"log"
	"net/http"
	"strconv"
	"web-chat/models"
	"web-chat/websocket"

	"github.com/gin-gonic/gin"
)

func ServeRoomWs(c *gin.Context) {
	// pathparamのgroupIdを取得
	// groupID->*hubを取得
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "Invalid room ID"})
		return
	}
	hub := models.RoomToHub[uint(roomID)]
	// 連想配列(2次元配列、[2][任意]）  const arrays[2][10]
	// key  : roomID1,roomID2,roomID3,
	// value: hub1   ,hub2   ,hub3   ,
	ServeWs(hub, c.Writer, c.Request)
	return
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *websocket.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &websocket.Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
