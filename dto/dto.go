package dto

import (
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// DTO
type Incoming struct {
	EventType  string
	SourceType string
	GroupID    string
	SenderID   string
	Text       string
	Mentionees []*linebot.Mentionee
	ReplyToken string
}

func ToIncoming(ev *linebot.Event) Incoming {
	in := Incoming{
		EventType:  string(ev.Type),
		ReplyToken: ev.ReplyToken,
	}
	if ev.Source != nil {
		in.SourceType = string(ev.Source.Type)
		in.GroupID = ev.Source.GroupID
		in.SenderID = ev.Source.UserID
	}
	if msg, ok := ev.Message.(*linebot.TextMessage); ok {
		in.Text = msg.Text
		if msg.Mention != nil {
			in.Mentionees = msg.Mention.Mentionees
		} else {
			in.Mentionees = []*linebot.Mentionee{}
		}
	}
	return in
}
