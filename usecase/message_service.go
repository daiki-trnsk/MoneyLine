package usecase

import (
	"context"
	"log"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/constants"
	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

func HandleEvent(ctx context.Context, bot *linebot.Client, in dto.Incoming) (linebot.SendingMessage, error) {
	switch in.EventType {
	case string(linebot.EventTypeJoin):
		return linebot.NewTextMessage(constants.JoinMessage), nil
	case string(linebot.EventTypeMessage):
		// グループ以外は処理しない
		if in.SourceType == "" || in.GroupID == "" {
			return linebot.NewTextMessage(constants.PrivateChatMessage), nil
		}

		// メンションされていない場合はスキップ
		botUserID := os.Getenv("MONEYLINE_BOT_ID")
		if !utils.IsMentioned(in.Mentionees, botUserID) {
			return nil, nil
		}

		// TestMentionの引数をdto.Incomingに合わせて修正
		replyMessage, err := TestMention(bot, in.GroupID, in.Mentionees)
		if err != nil {
			log.Println("Error in TestMention:", err)
			return linebot.NewTextMessage("エラーが発生しました。"), nil
		}

		if replyMessage == nil {
			return linebot.NewTextMessage("無効なメッセージです。"), nil
		}
		return replyMessage, nil
	}
	return nil, nil
}

func ReplyMessage(ctx context.Context, in dto.Incoming) error {

	return nil
}
