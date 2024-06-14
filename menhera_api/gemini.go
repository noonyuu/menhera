package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/api/option"
)

var models *genai.GenerativeModel

func GeminiInit() {
	// Gemini API の API キー
	err := error(nil)
	models, err = GeminiAI(os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}
}

func GeminiAI(apiKey string) (*genai.GenerativeModel, error) {
	c := context.Background()
	client, err := genai.NewClient(c, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	// モデルを初期化
	model := client.GenerativeModel("gemini-1.0-pro")
	return model, nil
}

// AIに問い合わせ
func QuestionAI(que string) (linebot.TextMessage, error) {
	c := context.Background()
	//モデル取得
	if models == nil {
		return linebot.TextMessage{
			Text: "モデルが初期化されていません",
		}, fmt.Errorf("models is not initialized")
	}

	var message linebot.TextMessage
	//生成
	query := "あなたは重度のメンヘラです。\n" + que + "\nこのメッセージに日本語を使用して50文字以内で返答してください" // ここで文字列を正しく結合
	var response, err = models.GenerateContent(c, genai.Text(query))
	if err != nil {
		log.Println(err)
		return linebot.TextMessage{
			Text: "エラーが発生しました",
		}, err
	}
	ans := printResponse(response)

	message = linebot.TextMessage{
		Text: ans,
	}
	return message, nil
}

func printResponse(resp *genai.GenerateContentResponse) string {
	resultText := ""
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				resultText += fmt.Sprint(part)
			}
		}
	}
	return resultText
}

// AIに聞く
func CallAI(message string) (linebot.TextMessage, error) {
	//AI生成
	aiGenerate, err := QuestionAI(message)
	if err != nil {
		log.Println(err)
		return linebot.TextMessage{}, err
	}

	return aiGenerate, nil
}
