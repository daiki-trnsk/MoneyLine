package usecase

import (
	"fmt"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func TestMention(msg *linebot.TextMessage) (*linebot.TextMessage, error) {
	var mentionees []linebot.Mentionee
	replyText := "Mentioned user:"
	index := len(replyText)

	for i, m := range msg.Mention.Mentionees {
		botUserID := os.Getenv("MONEYLINE_BOT_ID")
		if m.UserID == botUserID {
			continue
		}
		name := fmt.Sprintf(" @メンバー%d", i+1)
		fmt.Printf("DEBUG: Mentionee UserID=%s, name=%s, Index=%d, Length=%d\n", m.UserID, name, index+1, len(name)-1)
		replyText += name
		mentionees = append(mentionees, linebot.Mentionee{
			Index:  index + 1,     // 空白の次から
			Length: len([]rune(name)) - 1, // 空白分を除く
			UserID: m.UserID,
		})
		index += len([]rune(name))
	}

	if len(mentionees) == 0 {
		fmt.Println("DEBUG: No mentionees found")
		return nil, nil
	}

	mentioneesPtr := make([]*linebot.Mentionee, len(mentionees))
	for i := range mentionees {
		mentioneesPtr[i] = &mentionees[i]
	}

	fmt.Printf("DEBUG: replyText=%s\n", replyText)

	replyMessage := linebot.NewTextMessage(replyText)
	replyMessage.Mention = &linebot.Mention{
		Mentionees: mentioneesPtr,
	}
	return replyMessage, nil
}
