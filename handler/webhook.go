package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/constants"
	"github.com/daiki-trnsk/MoneyLine/usecase"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

func WebhookHandler(bot *linebot.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		events, err := bot.ParseRequest(c.Request())
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		for _, event := range events {
			replyMessage := handleEvent(bot, event)
			if replyMessage != nil {
				if _, err := bot.ReplyMessage(event.ReplyToken, replyMessage).Do(); err != nil {
					log.Println("Reply Error:", err)
				}
			}
		}

		return c.String(http.StatusOK, "ok")
	}
}

func handleEvent(bot *linebot.Client, event *linebot.Event) *linebot.TextMessage {
	switch event.Type {
	case linebot.EventTypeJoin:
		return linebot.NewTextMessage(constants.JoinMessage)
	case linebot.EventTypeMessage:
		// グループ以外は処理しない
		if event.Source == nil || event.Source.GroupID == "" {
			return linebot.NewTextMessage(constants.PrivateChatMessage)
		}

		msg, ok := event.Message.(*linebot.TextMessage)
		if !ok {
			return nil
		}

		// メンションされていない場合はスキップ
		botUserID := os.Getenv("MONEYLINE_BOT_ID")
		if !utils.IsMentioned(msg, botUserID) {
			return nil
		}

		replyMessage, err := usecase.TestMention(bot, event.Source.GroupID, msg)
		if err != nil {
			log.Println("Error in TestMention:", err)
			return linebot.NewTextMessage("エラーが発生しました。")
		}
		if replyMessage == nil {
			return linebot.NewTextMessage("無効なメッセージです。")
		}
		return replyMessage
	}
	return nil
}

