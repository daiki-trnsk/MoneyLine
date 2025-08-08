package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/handler"
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

	e.POST("/webhook", handler.WebhookHandler(bot))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
