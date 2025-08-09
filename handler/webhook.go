package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/usecase"
	"github.com/daiki-trnsk/MoneyLine/dto"
)

func WebhookHandler(bot *linebot.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		events, err := bot.ParseRequest(c.Request())
		if err != nil {
			log.Printf("Error parsing request: %v", err)
			return c.NoContent(http.StatusOK)
		}

		for _, event := range events {
			in := dto.ToIncoming(event)
			reply, err := usecase.HandleEvent(c.Request().Context(), bot, in)
			if err != nil {
				log.Printf("Error handling event: %v", err)
				return c.NoContent(http.StatusOK)
			}
			if reply != nil {
				if _, err := bot.ReplyMessage(in.ReplyToken, reply).Do(); err != nil {
					log.Printf("Error replying message: %v", err)
					return c.NoContent(http.StatusOK)
				}
			}
		}
		return c.NoContent(http.StatusOK)
	}
}