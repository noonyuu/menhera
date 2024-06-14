package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

func loadEnv() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Panicln("読み込み出来ませんでした: %v", err)
	}
}

func main() {
	loadEnv()
	server := gin.Default()

	var bot *linebot.Client
	var err error

	jwt, err := Jwt()
	if err != nil {
		bot, err = linebot.New(
			os.Getenv("LINE_BOT_CHANNEL_SECRET"),
			os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
		)
	} else {
		bot, err = linebot.New(
			os.Getenv("LINE_BOT_CHANNEL_SECRET"),
			os.Getenv(jwt),
		)
	}

	server.GET("/", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	server.POST("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				log.Print(err)
				c.Status(http.StatusBadRequest)
			} else {
				log.Print(err)
				c.Status(http.StatusInternalServerError)
			}
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					// CallAIを呼び出し、その結果をaiResponseに代入
					aiResponse, err := CallAI(message.Text)
					if err != nil {
						log.Print(err)
						aiResponse = linebot.TextMessage{Text: message.Text}
					}

					// aiResponseのTextフィールドの内容を新しいテキストメッセージとして送信
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(aiResponse.Text)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
		c.Status(http.StatusOK)
	})

	server.Run(":3012")
}
