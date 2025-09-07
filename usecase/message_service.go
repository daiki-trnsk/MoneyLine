package usecase

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/constants"
	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

// HandleEvent リクエストを解析し各処理呼び出し、返信メッセージ返す
func HandleEvent(ctx context.Context, bot *linebot.Client, in dto.Incoming) linebot.SendingMessage {
	switch in.EventType {
	case string(linebot.EventTypeJoin):
		subject := "New Group Joined"
		body := "A new group has added the LINE bot."
		if err := utils.SendEmail(subject, body); err != nil {
			log.Printf("Error sending notification email: %v", err)
		}
		return linebot.NewTextMessage(constants.JoinMessage)
	case string(linebot.EventTypeMessage):
		// グループ以外は処理しない
		if in.SourceType == "" || in.GroupID == "" {
			return linebot.NewTextMessage(constants.PrivateChatMessage)
		}

		// メンションされていない場合はスキップ
		botUserID := os.Getenv("MONEYLINE_BOT_ID")
		if !utils.IsMentioned(in.Mentionees, botUserID) {
			return nil
		}

		fmt.Println(in.Text)

		// メッセージ解析、処理分岐
		switch utils.DetectCommand(in) {
		case utils.CmdPay:
			return Pay(bot, in)
		case utils.CmdSummary:
			return SettleGreedy(bot, in)
		case utils.CmdHistory:
			return History(bot, in)
		case utils.CmdOneClear:
			return OneClear(bot, in)
		case utils.CmdAllClear:
			return AllClear(bot, in)
		case utils.CmdHelp:
			return linebot.NewTextMessage(constants.HelpMessage)
		default:
			return linebot.NewTextMessage(constants.InvalidMessage)
		}
	}
	return nil
}
