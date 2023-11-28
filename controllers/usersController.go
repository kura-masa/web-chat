package controllers

import (
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
	"web-chat/initializers"
	"web-chat/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	// Get the username/pass off req body
	var body struct {
		UserName string
		Password string
	}

	body.UserName = c.PostForm("username")
	body.Password = c.PostForm("password")

	username := body.UserName
	password := body.Password

	// ユーザー名のエラー処理
	if !(strings.Contains(username, "@")) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ユーザー名には@が必要です",
		})
		return
	}

	// パスの条件
	// 文字が7文字以上。数字・大文字・句読点と記号が1文字以上。空白なし。
	sevenOrMore, number, upper, special, space := VerifyPassword(password)
	if !sevenOrMore || !number || !upper || !special || space {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"title":  "Home",
			"result": "passwordの条件を満たしてません",
		})
		return
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})

		return
	}

	// Create the user
	user := models.Users{UserName: body.UserName, Password: string(hash)}
	result := initializers.DB.Create(&user) // pass pointer of data to Create

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})

		return
	}

	// Respond
	c.HTML(http.StatusOK, "auth.html", gin.H{
		"title":  "Auth",
		"result": "Hi " + body.UserName + ", Success to Signup",
	})
}

func Login(c *gin.Context) {
	// Get the username/pass off req body
	var body struct {
		UserName string
		Password string
	}

	// UserNameとpasswordを取得
	// if c.Bind(&body) != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"error": "Failed to read body",
	// 	})
	// 	return
	// }
	body.UserName = c.PostForm("username")
	body.Password = c.PostForm("password")

	// Look up requested user
	var user models.Users
	initializers.DB.First(&user, "user_name = ?", body.UserName)
	// SELECT * FROM users WHERE user_name = ○○;
	// usernameはunique

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid username or password",
		})

		return
	}

	// Compare sent in pass with saved user pass hash
	// DBに保存されているhash化されたパスワードとログイン時に入力されたパスワードの比較
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid username or password",
		})

		return
	}

	// Generate a jwt token
	// Tokenとは使い勝手のいい形（コストが小さい）をした情報のかたまり
	// jwt.NewWithClaims()の第一引数は署名方法（指定した時点でalgとtypが設定される）
	// 第二引数はtokenの内容を指定（キー(claims））
	// ここでは指定したメソッドを使った単なる文字変換をしている
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	// ここで暗号化（.envファイルで指定したキーを使って）
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})

		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	// send it back
	c.Redirect(http.StatusMovedPermanently, "/home")
}

func Logout(c *gin.Context) {
	// Cookieを期限切れにする
	c.SetCookie("Authorization", "", -1000, "", "", false, true)
	c.HTML(http.StatusOK, "auth.html", gin.H{
		"title":  "Auth",
		"result": "Success to Logout",
	})
}

func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"message": user,
	})
}

func VerifyPassword(s string) (sevenOrMore, number, upper, special, space bool) {
	letters := 0
	for _, c := range s {
		switch {
		// 少なくとも一つ数字が入っている必要がある
		case unicode.IsNumber(c):
			number = true
		// 少なくとも一つ大文字が入っている必要がある
		case unicode.IsUpper(c):
			upper = true
			letters++
		// IsPunct:句読点,IsSymbol:記号
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c):
			letters++
		// スペースはだめ
		case c == ' ':
			space = true
		default:
			//return false, false, false, false
		}
	}
	// 少なくとも7文字以上含まれている必要がある
	sevenOrMore = letters >= 7
	return
}
