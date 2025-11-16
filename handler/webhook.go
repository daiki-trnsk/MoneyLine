package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/usecase"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

// WebhookHandler LINEのWebhookイベントをDTO変換して処理し、返信を実行するハンドラー
func WebhookHandler(bot *linebot.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		events, err := bot.ParseRequest(c.Request())
		if err != nil {
			log.Printf("Error parsing request: %v", err)
			return c.NoContent(http.StatusOK)
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage || event.Type == linebot.EventTypeJoin || event.Type == linebot.EventTypeLeave {
				in := dto.ToIncoming(event)
				reply := usecase.HandleEvent(c.Request().Context(), bot, in)
				if reply != nil {
					if _, err := bot.ReplyMessage(in.ReplyToken, reply).Do(); err != nil {
						log.Printf("Error replying message: %v", err)
						return c.NoContent(http.StatusOK)
					}
				}
			}

			if event.Type == linebot.EventTypeFollow {
				subject := "New Friend Added"
				profile, err := bot.GetProfile(event.Source.UserID).Do()
				if err != nil {
					log.Printf("Error fetching user profile: %v", err)
					return c.NoContent(http.StatusOK)
				}

				displayName := profile.DisplayName
				body := "A new friend has been added to the LINE bot: " + displayName
				if err := utils.SendEmail(subject, body); err != nil {
					log.Printf("Error sending notification email: %v", err)
				}
			}
		}
		return c.NoContent(http.StatusOK)
	}
}
