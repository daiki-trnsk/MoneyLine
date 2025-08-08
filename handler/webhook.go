package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/constants"
	"github.com/daiki-trnsk/MoneyLine/usecase"
)

func WebhookHandler(bot *linebot.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		events, err := bot.ParseRequest(c.Request())
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		for _, event := range events {
			var replyMessage *linebot.TextMessage

			switch event.Type {
			case linebot.EventTypeJoin:
				replyMessage = linebot.NewTextMessage(constants.JoinMessage)
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

				// メンションされていない場合はここでスキップ
				if !found {
					continue
				}

				var groupID string
				if event.Source != nil {
					groupID = event.Source.GroupID
				}

				// 以下、各処理をusecaseから呼び出すreplyMessage, err = usecase.hogefuga
				replyMessage, err = usecase.TestMention(bot, groupID, msg)
				if err != nil {
					log.Println("Error in TestMention:", err)
					replyMessage = linebot.NewTextMessage("エラーが発生しました。")
				}
				if replyMessage == nil {
					replyMessage = linebot.NewTextMessage("無効なメッセージです。")
				}
			}

			if replyMessage != nil {
				if _, err := bot.ReplyMessage(event.ReplyToken, replyMessage).Do(); err != nil {
					log.Println("Reply Error:", err)
				}
			}
		}

		return c.String(http.StatusOK, "ok")
	}
}
