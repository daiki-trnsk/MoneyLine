package handler

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var (
	mentionPattern = regexp.MustCompile(`@([^\s]+)\s*@([^\s]+)\s*(\d+)\s*(.*)`)
)

func WebhookHandler(bot *linebot.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		events, err := bot.ParseRequest(c.Request())
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				msg, ok := event.Message.(*linebot.TextMessage)
				if !ok {
					continue
				}

				// メンションが2人（Bot + 債務者）含まれているか確認
				if msg.Mention == nil || len(msg.Mention.Mentionees) < 2 {
					log.Println("メンションが足りない or 無効")
					continue
				}

				// 債権者（送信者）
				creditorID := event.Source.UserID

				// 債務者
				debtorID := msg.Mention.Mentionees[1].UserID

				// 金額とメモを抽出
				tokens := strings.Fields(msg.Text)
				if len(tokens) < 3 {
					log.Println("金額やメモが足りません")
					continue
				}

				amount, err := strconv.Atoi(tokens[2])
				if err != nil {
					log.Println("金額が数値でない:", tokens[2])
					continue
				}

				memo := strings.Join(tokens[3:], " ")

				log.Printf("債権者（送信者）: %s, 債務者（メンション）: %s, 金額: %d, メモ: %s", creditorID, debtorID, amount, memo)

				if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("記録しました！")).Do(); err != nil {
					log.Println("返信エラー:", err)
				}
			}
		}

		return c.String(http.StatusOK, "ok")
	}
}