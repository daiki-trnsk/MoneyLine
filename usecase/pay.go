package usecase

import (
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func Pay(bot *linebot.Client, groupID string, text string, mentionees []*linebot.Mentionee) (*linebot.TextMessage, error) {
	return linebot.NewTextMessage("Pay method called with message: " + text), nil
}

func Summary(bot *linebot.Client, groupID string, text string, mentionees []*linebot.Mentionee) (*linebot.TextMessage, error) {
	return linebot.NewTextMessage("Summary method called with message: " + text), nil
}

func History(bot *linebot.Client, groupID string, text string, mentionees []*linebot.Mentionee) (*linebot.TextMessage, error) {
	return linebot.NewTextMessage("History method called with message: " + text), nil
}
