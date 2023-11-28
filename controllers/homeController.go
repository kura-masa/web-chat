package controllers

import (
	"net/http"
	"strconv"
	"web-chat/initializers"
	"web-chat/models"

	"github.com/gin-gonic/gin"
)

func Lists(c *gin.Context) {
	// 自分のIDに該当するfriendsテーブルを取得
	user, _ := c.Get("user")
	var friends []models.Friends
	userID := user.(models.Users).ID
	result := initializers.DB.Where("user_id1 = ? OR user_id2 = ?", userID, userID).Find(&friends)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	// friendsテーブルから、友達のIDのみを抽出
	var friendsIDs []uint
	for _, friend := range friends {
		if friend.UserID1 == userID {
			friendsIDs = append(friendsIDs, friend.UserID2)
		} else {
			friendsIDs = append(friendsIDs, friend.UserID1)
		}
	}
	// 友達IDをもとにusersテーブルから友達の名前を取得。 IN:スライスを扱える。Pluck():スライス形式で渡せる
	var users []models.Users
	result = initializers.DB.Where("id IN ?", friendsIDs).Find(&users)

	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	// 自分のIDに該当するgroupsテーブルを取得し、ルーム名を取得する
	groups := []models.Groups{}
	rooms := []models.Rooms{}
	var rooms_refer []uint
	result = initializers.DB.Where("users_refer = ?", userID).Find(&groups)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	for _, r := range groups {
		rooms_refer = append(rooms_refer, r.RoomsRefer)
	}
	result = initializers.DB.Where("id IN ?", rooms_refer).Find(&rooms)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	c.HTML(http.StatusOK, "home.html", gin.H{
		"title": "Home",
		"users": users,
		"rooms": rooms,
	})
}

func SearchFriend(c *gin.Context) {
	user, _ := c.Get("user")
	user_name := (c.PostForm("searchBar"))
	if user_name == user.(models.Users).UserName {
		Result(c, "自分は追加できません", models.Users{})
		return
	}
	if user_name == "" {
		Result(c, "ユーザー名を入力してください", models.Users{})
		return
	}
	user1 := models.Users{}
	result := initializers.DB.Where("user_name = ?", user_name).First(&user1)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	if user1.UserName == "" {
		Result(c, "ユーザーが見つかりませんでした", models.Users{})
		return
	}
	// もう友達なら友達追加しない
	friend := models.Friends{}
	conditions := "(user_id1 = ? AND user_id2 = ?) OR (user_id1 = ? AND user_id2 = ?)"
	userID := user.(models.Users).ID
	result = initializers.DB.Where(conditions, userID, user1.ID, user1.ID, userID).First(&friend)
	if friend.UserID1 != 0 {
		Result(c, "すでにフレンドです", models.Users{})
		return
	}
	Result(c, "ユーザーが見つかりました", user1)
}

func AddFriend(c *gin.Context) {
	user, _ := c.Get("user")
	user_id2, _ := strconv.Atoi(c.PostForm("userID"))
	user_name := (c.PostForm("userName"))
	room_name := user.(models.Users).UserName + " & " + user_name
	room := models.Rooms{RoomName: room_name}
	//　部屋作成
	result := initializers.DB.Create(&room)
	if result.Error != nil {
		Result(c, "ユーザーを追加できませんでした", models.Users{})
		return
	}
	friend := models.Friends{
		UserID1:    user.(models.Users).ID,
		UserID2:    uint(user_id2),
		RoomsRefer: room.ID,
	}
	// フレンド追加
	result = initializers.DB.Create(&friend)
	if result.Error != nil {
		// 一度作った部屋を消す
		initializers.DB.Where("id = ?", room.ID).Delete(models.Rooms{})
		Result(c, "ユーザーを追加できませんでした", models.Users{})
		return
	}
	c.Redirect(http.StatusSeeOther, "/home")
}

func CreateGroup(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(models.Users)
	type GroupCreationRequest struct {
		GroupName       string   `json:"groupName"`
		SelectedFriends []string `json:"selectedFriends"`
	}
	var request GroupCreationRequest
	c.ShouldBindJSON(&request)
	room_name := request.GroupName
	var u_ids []uint
	u_ids = append(u_ids, u.ID)
	for _, u_id := range request.SelectedFriends {
		id, _ := strconv.Atoi(u_id)
		u_ids = append(u_ids, uint(id))
	}
	room := models.Rooms{RoomName: room_name}
	// ルーム作成
	result := initializers.DB.Debug().Create(&room)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	// グループメンバー作成
	for _, u := range u_ids {
		group_member := models.Groups{UsersRefer: u, RoomsRefer: room.ID}
		result = initializers.DB.Debug().Create(&group_member)

		if result.Error != nil {
			// 一度作った部屋を消す
			initializers.DB.Where("id = ?", room.ID).Delete(&models.Rooms{})
			// 一度作ったgroup_memberを消す
			initializers.DB.Where("users_refer IN ? AND rooms_refer = ?", u_ids, room.ID).Delete(&models.Groups{})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
			return
		}
	}
}

func DeleteFriend(c *gin.Context) {
	user, _ := c.Get("user")
	u_id := user.(models.Users).ID
	f_id, _ := strconv.Atoi(c.PostForm("userid"))
	friend := models.Friends{}
	conditions := "(user_id1 = ? AND user_id2 = ?) OR (user_id1 = ? AND user_id2 = ?)"
	result := initializers.DB.Debug().
		Where(conditions, f_id, u_id, u_id, f_id).
		Find(&friend)
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	result = initializers.DB.Debug().Where("id = ?", friend.RoomsRefer).Delete(&models.Rooms{})
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}

	result = initializers.DB.Debug().
		Where(conditions, f_id, u_id, u_id, f_id).
		Delete(&models.Friends{})
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/home")
}

func DeleteGroup(c *gin.Context) {
	room_id := c.PostForm("room_id")
	result := initializers.DB.Debug().
		Where("id = ?", room_id).Delete(&models.Rooms{})
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	result = initializers.DB.Debug().
		Where("rooms_refer = ?", room_id).Delete(&[]models.Groups{})
	if result.Error != nil {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "エラーが発生しました",
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/home")
}

func Result(c *gin.Context, title string, user models.Users) {
	c.HTML(http.StatusOK, "result.html", gin.H{
		"title": title,
		"user":  user,
	})
}
