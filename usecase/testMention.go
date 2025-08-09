package usecase

import (
	"fmt"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func TestMention(bot *linebot.Client, groupID string, mentionees []*linebot.Mentionee) (*linebot.TextMessage, error) {
	var newMentionees []linebot.Mentionee
	replyText := "Mentioned user:"
	index := len([]rune(replyText))

	for _, m := range mentionees {
		botUserID := os.Getenv("MONEYLINE_BOT_ID")
		if m.UserID == botUserID {
			continue
		}
		// グループメンバーの表示名を取得
		profile, err := bot.GetGroupMemberProfile(groupID, m.UserID).Do()
		if err != nil {
			continue
		}
		name := fmt.Sprintf(" %s", profile.DisplayName)
		replyText += name
		newMentionees = append(newMentionees, linebot.Mentionee{
			Index:  index,
			Length: len([]rune(name)),
			UserID: m.UserID,
		})
		index += len([]rune(name))
	}

	if len(newMentionees) == 0 {
		return nil, nil
	}

	mentioneesPtr := make([]*linebot.Mentionee, len(newMentionees))
	for i := range newMentionees {
		mentioneesPtr[i] = &newMentionees[i]
	}

	fmt.Printf("DEBUG: replyText=%s\n", replyText)

	replyMessage := linebot.NewTextMessage(replyText)
	replyMessage.Mention = &linebot.Mention{
		Mentionees: mentioneesPtr,
	}
	return replyMessage, nil
}
