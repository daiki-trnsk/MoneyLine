package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/constants"
)

func WebhookHandler(bot *linebot.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		events, err := bot.ParseRequest(c.Request())
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		for _, event := range events {
			var replyMessage string

			switch event.Type {
			case linebot.EventTypeJoin:
				replyMessage = constants.JoinMessage
			case linebot.EventTypeMessage:
				msg, ok := event.Message.(*linebot.TextMessage)
				if !ok {
					continue
				}

				// マネリンがメンションされたか
				botUserID := os.Getenv("MONEYLINE_BOT_ID")
				found := false
				if msg.Mention != nil {
					for _, m := range msg.Mention.Mentionees {
						if m.UserID == botUserID {
							found = true
							break
						}
					}
				}
				if !found {
					continue
				}
				replyMessage = "メンションされました！"
			}

			if replyMessage != "" {
				if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
					log.Println("Reply Error:", err)
				}
			}
		}

		return c.String(http.StatusOK, "ok")
	}
}
