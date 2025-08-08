package usecase

import (
	"fmt"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func TestMention(msg *linebot.TextMessage) (*linebot.TextMessage, error) {
	var mentionees []linebot.Mentionee
	replyText := "Mentioned user:"
	offset := 0
	index := len(replyText)

	for _, m := range msg.Mention.Mentionees {
		botUserID := os.Getenv("MONEYLINE_BOT_ID")
		if m.UserID == botUserID {
			continue // Botはスキップ
		}
		name := fmt.Sprintf(" @メンバー%d", offset+1)
		replyText += name
		mentionees = append(mentionees, linebot.Mentionee{
			Index:  index + 1,     // 空白の次から
			Length: len(name) - 1, // 空白分を除く
			UserID: m.UserID,
		})
		index += len(name)
		offset++
	}

	if len(mentionees) == 0 {
		return nil, nil
	}

	mentioneesPtr := make([]*linebot.Mentionee, len(mentionees))
	for i := range mentionees {
		mentioneesPtr[i] = &mentionees[i]
	}

	replyMessage := linebot.NewTextMessage(replyText)
	replyMessage.Mention = &linebot.Mention{
		Mentionees: mentioneesPtr,
	}
	return replyMessage, nil
}
