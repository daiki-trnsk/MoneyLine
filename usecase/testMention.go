package usecase

import (
	"os"
	"fmt"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func TestMention(msg *linebot.TextMessage) (*linebot.TextMessage, error) {
	var mentionees []linebot.Mentionee
	text := msg.Text
	offset := 0

	for _, m := range msg.Mention.Mentionees {

		botUserID := os.Getenv("MONEYLINE_BOT_ID")
		if m.UserID == botUserID {
			continue // 自分（Bot）はスキップ
		}

		// メンション表記（仮）を挿入する（例: @メンバー1 @メンバー2）
		name := fmt.Sprintf("@メンバー%d", offset+1) // 表示名ではなく仮名
		start := len(text)
		text += " " + name
		mentionees = append(mentionees, linebot.Mentionee{
			Index:  start + 1, // 空白の次から
			Length: len(name),
			UserID: m.UserID,
		})
		offset++
	}

	if len(mentionees) == 0 {
		return nil, nil
	}

	// 返信メッセージを構築
	replyText := fmt.Sprintf("Mentioned user: %s", text[len(msg.Text):]) // 元のメッセージ以降を表示
	// Convert []linebot.Mentionee to []*linebot.Mentionee
	mentioneesPtr := make([]*linebot.Mentionee, len(mentionees))
	for i := range mentionees {
		mentioneesPtr[i] = &mentionees[i]
	}

	replyMessage := &linebot.TextMessage{
		Text: replyText,
		Mention: &linebot.Mention{
			Mentionees: mentioneesPtr,
		},
	}
	return replyMessage, nil
}
