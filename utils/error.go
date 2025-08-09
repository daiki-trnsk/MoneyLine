package utils

import (
	"log"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func HandleError(ctxStr string, err error) (linebot.SendingMessage, error) {
	log.Printf("Error in %s: %v", ctxStr, err)
	return linebot.NewTextMessage("エラーが発生しました。"), nil
}
