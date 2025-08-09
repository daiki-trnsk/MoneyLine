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
			fmt.Printf("DEBUG: GetGroupMemberProfile error: %v\n", err)
			continue
		}
		name := fmt.Sprintf(" %s", profile.DisplayName)
		fmt.Printf("DEBUG: Mentionee UserID=%s, name=%s, Index=%d, Length=%d\n", m.UserID, name, index, len([]rune(name)))
		replyText += name
		newMentionees = append(newMentionees, linebot.Mentionee{
			Index:  index,
			Length: len([]rune(name)),
			UserID: m.UserID,
		})
		index += len([]rune(name))
	}

	if len(newMentionees) == 0 {
		fmt.Println("DEBUG: No mentionees found")
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
