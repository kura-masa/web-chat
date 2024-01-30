package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"web-chat/initializers"
	"web-chat/models"

	"github.com/gin-gonic/gin"
)

func OpenAIContextTranslate(c *gin.Context) {
	room_id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		log.Fatalln(err)
	}
	var chat_history []models.Chat_history
	initializers.DB.Where("room_id = ?", room_id).Order("created_at DESC").Find(&chat_history)
	question := ""
	chat_length := 0
	u1 := chat_history[0].UserID
	println(u1)
	for _, chat := range chat_history {
		chat_length += len(chat.Content)
		if chat_length > 150 {
			break
		}
		if u1 == chat.UserID {
			question = "A:" + chat.Content + question
		} else {
			question = "B:" + chat.Content + question
		}
	}
	question += "というチャットの'" + chat_history[0].Content + "'の目的語が無い場合,補完した英文を出力して。また、代名詞は明確にした英文を出力して"

	println("AIに送った質問の内容")
	println(question)
	apiResponse, err := queryOpenAI(question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying OpenAI 1"})
		return
	}

	question = apiResponse + "日本語訳して"
	apiResponse, err = queryOpenAI(question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying OpenAI 2"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": apiResponse})
}

func queryOpenAI(question string) (string, error) {
	// OpenAI APIのエンドポイント モデルはgpt-3.5-turbo
	apiURL := "https://api.openai.com/v1/chat/completions"

	// 環境変数からOpenAI APIキーを取得
	apiKey := os.Getenv("OPENAI_API")
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not set")
	}

	// OpenAIにリクエストを送信
	requestData := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": question},
		},
		"max_tokens": 50,
		"model":      "gpt-4-0125-preview",
	}

	requestBody, err := initializers.JSONMarshal(requestData)
	if err != nil {
		return "", fmt.Errorf("Error marshaling request data")
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("Error creating request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error making request")
	}
	defer resp.Body.Close()

	var m map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {

		return "error after sending a request to api", err
	}
	if errVal, ok := m["error"].(map[string]interface{}); ok {

		errorCode := errVal["code"].(string)

		return errorCode, errors.New(errorCode)
	}

	responseContent, _ := initializers.JSONMarshal(m)
	fmt.Println("OpenAI Full Response:")
	fmt.Println(string(responseContent))

	content := m["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	return content, err
}
