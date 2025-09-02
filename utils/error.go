package utils

import (
	"fmt"
	"log"
	"runtime"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// エラーログ出力とエラーメッセージ返信
func LogAndReplyError(err error, in dto.Incoming, title string) linebot.SendingMessage {
	_, file, line, ok := runtime.Caller(1)
	loc := "unknown"
	if ok {
		loc = fmt.Sprintf("%s:%d", file, line)
	}

	// ログ出力
	log.Printf("[ERROR] title=%s err=%v caller=%s incoming=%+v",
		title, err, loc, in,
	)

	// メール通知
	subject := fmt.Sprintf("[ERROR] %s", title)
	body := fmt.Sprintf(
		"Error occurred:\n\nTitle: %s\nError: %v\nLocation: %s\nIncoming: %+v",
		title, err, loc, in,
	)
	if emailErr := SendEmail(subject, body); emailErr != nil {
		log.Printf("[ERROR] Failed to send error notification email: %v", emailErr)
	}

	return linebot.NewTextMessage("エラーが発生しました。")
}
