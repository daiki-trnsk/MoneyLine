package usecase

import (
	"fmt"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func TestMention(bot *linebot.Client, groupID string, msg *linebot.TextMessage) (*linebot.TextMessage, error) {
	var mentionees []linebot.Mentionee
	replyText := "Mentioned user:"
	index := len([]rune(replyText))

	for _, m := range msg.Mention.Mentionees {
		botUserID := os.Getenv("MONEYLINE_BOT_ID")
		if m.UserID == botUserID {
			continue
		}
		// グループメンバーの表示名を取得
		profile, err := bot.GetGroupMemberProfile(groupID, m.UserID).Do()
		if err != nil {
			fmt.Printf("DEBUG: GetGroupMemberProfile error: %v\n", err)
			continue
		}
		name := fmt.Sprintf(" %s", profile.DisplayName)
		fmt.Printf("DEBUG: Mentionee UserID=%s, name=%s, Index=%d, Length=%d\n", m.UserID, name, index, len([]rune(name)))
		replyText += name
		mentionees = append(mentionees, linebot.Mentionee{
			Index:  index,
			Length: len([]rune(name)),
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
