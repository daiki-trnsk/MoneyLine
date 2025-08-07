package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	
    bot, err := linebot.New(
        os.Getenv("LINE_CHANNEL_SECRET"),
        os.Getenv("LINE_CHANNEL_TOKEN"),
    )
    if err != nil {
        log.Fatal("Error initializing LINE bot:", err)
    }

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "MoneyLine Server is running!")
	})

    e.POST("/webhook", func(c echo.Context) error {
        events, err := bot.ParseRequest(c.Request())
        if err != nil {
            return c.NoContent(http.StatusBadRequest)
        }

        for _, event := range events {
            if event.Type == linebot.EventTypeMessage {
                if msg, ok := event.Message.(*linebot.TextMessage); ok {
					fmt.Printf("moneyline userid: %s\n", msg.Mention.Mentionees[0].UserID)
                    bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("受信したメッセージ: "+msg.Text)).Do()
                }
            }
        }

        return c.String(http.StatusOK, "ok")
    })
	e.Logger.Fatal(e.Start(":8000"))
}
