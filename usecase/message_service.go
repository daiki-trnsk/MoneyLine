package usecase

import (
	"context"
	"os"
	"strings"

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

		// メッセージの最初に@マネリンがいない場合はエラー返す
		fields := strings.Fields(in.Text)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], "@") || fields[0] != "@マネリン|金銭管理Bot" {
			return linebot.NewTextMessage("@マネリンを文頭でメンションしてください。"), nil
		}

		// メッセージ解析、処理分岐
		switch utils.DetectCommand(in.Text) {
		case utils.CmdSummary:
			return Summary(bot, in)
		case utils.CmdHistory:
			return History(bot, in)
		case utils.CmdHelp:
			return linebot.NewTextMessage(constants.HelpMessage), nil
		default:
			// マネリン以外のメンション+数字でPay処理
			// 例: @マネリン 1000
			if len(in.Mentionees) > 1 && utils.ContainsNumber(in.Text) {
				return Pay(bot, in)
			}
			return linebot.NewTextMessage(constants.InvalidMessage), nil
		}
	}
	return nil, nil
}
