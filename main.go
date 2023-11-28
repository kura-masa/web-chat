package main

import (
	"net/http"
	"web-chat/controllers"
	"web-chat/initializers"
	"web-chat/middleware"
	"web-chat/websocket"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/**/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth.html", gin.H{
			"title": "Auth",
		})
	})

	// シナリオ説明
	r.GET("/scenario_lsits", func(c *gin.Context) {
		c.File("description/scenario_lists.html")
	})
	// シナリオ1
	r.GET("/scenario/scenario1", func(c *gin.Context) {
		c.File("description/scenario/scenario1.html")
	})
	// シナリオ：グループ削除
	r.GET("/scenario/delete_group", func(c *gin.Context) {
		c.File("description/scenario/delete_group.html")
	})
	// HTMLのリスト
	r.GET("/html_lists", func(c *gin.Context) {
		c.File("description/descriptions/html_lists.html")
	})
	// HTMLのformの説明
	r.GET("/html/form_description", func(c *gin.Context) {
		c.File("description/descriptions/html_descriptions/html_form.html")
	})
	// goの説明リスト
	r.GET("/go_lists", func(c *gin.Context) {
		c.File("description/descriptions/go_lists.html")
	})
	// gormのCRUDの説明
	r.GET("/go/gorm_description", func(c *gin.Context) {
		c.File("description/descriptions/go_descriptions/gorm.html")
	})
	// 論理削除・物理削除の説明
	r.GET("/logical_physical_deletion", func(c *gin.Context) {
		c.File("description/descriptions/go_descriptions/deletion_ways.html")
	})
	// auth.htmlに戻る
	r.GET("/back2auth", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth.html", gin.H{
			"title": "Auth",
		})
	})

	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)
	r.POST("/logout", controllers.Logout)

	// Home
	r.GET("/home", middleware.RequireAuth, controllers.Lists)
	// search friend
	r.POST("/home/search_friend", middleware.RequireAuth, controllers.SearchFriend)
	// add friend
	r.POST("/home/add_friend", middleware.RequireAuth, controllers.AddFriend)
	// create group
	r.POST("/home/create_group", middleware.RequireAuth, controllers.CreateGroup)
	// delete friend
	r.POST("/home/destroy_friend", middleware.RequireAuth, controllers.DeleteFriend)
	// delete room
	r.POST("/home/destroy_group", middleware.RequireAuth, controllers.DeleteGroup)

	//
	//
	// chat history
	r.GET("/chat/chat", middleware.RequireAuth, controllers.ListChatHistory)
	// chat list
	r.GET("/chat/chatlist", middleware.RequireAuth, controllers.ChatList)
	r.POST("/chat/create", middleware.RequireAuth, controllers.CreateChat)
	// error
	r.GET("/error", func(c *gin.Context) {
		c.File("templates/home/error.html")
	})

	// ws
	hub := websocket.NewHub()
	go hub.Run()
	r.GET("/ws/:id", func(c *gin.Context) {
		controllers.ServeRoomWs(c)
	})
	r.Run()
}
