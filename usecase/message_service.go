package usecase

import (
	"context"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/constants"
	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

// HandleEvent リクエストを解析し各処理呼び出し、返信メッセージ返す
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

		// TODO: utilsで解析、処理分岐実装
		switch utils.DetectCommand(in.Text) {
		case utils.CmdSummary:
			return Summary(bot, in.GroupID, in.Text, in.Mentionees)
		case utils.CmdHistory:
			return History(bot, in.GroupID, in.Text, in.Mentionees)
		case utils.CmdHelp:
			return linebot.NewTextMessage(constants.HelpMessage), nil
		default:
			// マネリン以外のメンション+数字でPay処理
			// 例: @マネリン 1000
			if len(in.Mentionees) > 1 && utils.ContainsNumber(in.Text) {
				return Pay(bot, in.GroupID, in.Text, in.Mentionees)
			}
			return linebot.NewTextMessage(constants.InvalidMessage), nil
		}

		// replyMessage, err := TestMention(bot, in.GroupID, in.Mentionees)
		// if err != nil {
		// 	return handleError("TestMention", err)
		// }

		// if replyMessage == nil {
		// 	return linebot.NewTextMessage("無効なメッセージです。"), nil
		// }
	}
	return nil, nil
}
