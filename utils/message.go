package utils

import "github.com/line/line-bot-sdk-go/v7/linebot"

func IsMentioned(mentionees []*linebot.Mentionee, userID string) bool {
	for _, m := range mentionees {
		if m.UserID == userID {
			return true
		}
	}
	return false
}
