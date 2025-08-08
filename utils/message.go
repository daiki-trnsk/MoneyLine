package utils

import "github.com/line/line-bot-sdk-go/v7/linebot"

func IsMentioned(msg *linebot.TextMessage, userID string) bool {
	if msg.Mention == nil {
		return false
	}
	for _, m := range msg.Mention.Mentionees {
		if m.UserID == userID {
			return true
		}
	}
	return false
}
