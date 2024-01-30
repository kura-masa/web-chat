package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"web-chat/initializers"
	"web-chat/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *gin.Context) {
	// Get the cookie off req
	// 指定した名前のCookieのValueを取得
	tokenString, err := c.Cookie("Authorization")

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	// Decode/validate it
	// Parseはjwtトークンから元になった認証情報を復元
	// 第一引数は暗号化された文字列
	// 第二引数は署名方法が指定したものであるかを確認して復元のためのキーを返す関数
	token, err := jwt.Parse(tokenString,
		func(token *jwt.Token) (interface{}, error) {
			// 与えられたjwtトークンがHMAC署名メソッドで生成されたものであるかの確認
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(os.Getenv("SECRET")), nil
		})

	// 復元後のトークンのClaim（情報）を取得
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check the exp
		// 使用期限を確認
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// Find the user with token sub
		// DBから値を取得
		var user models.Users
		initializers.DB.First(&user, claims["sub"])

		if user.ID == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// Attach to req
		// gin.Context.Setで保存した情報を、gin.Context.Getで受け取ることができる
		c.Set("user", user)

		// Continue
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

}
