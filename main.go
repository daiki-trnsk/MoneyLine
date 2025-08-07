package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "MoneyLine Server is running!")
	})

	e.POST("/webhook", func(c echo.Context) error {
		body := make([]byte, c.Request().ContentLength)
		c.Request().Body.Read(body)
		fmt.Println("LINEからの受信:", string(body))
		return c.String(http.StatusOK, "ok")
	})
	e.Logger.Fatal(e.Start(":8000"))
}
