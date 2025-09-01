package usecase

import (
	"fmt"
	"strings"

	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/infra"
	"github.com/daiki-trnsk/MoneyLine/models"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

// 履歴（グループごとの取引履歴）
func History(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	var txs []models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Order("created_at desc").Find(&txs).Error; err != nil {
		return linebot.NewTextMessage("履歴取得に失敗しました。"), nil
	}

	msg := "履歴\n\n"

	for _, tx := range txs {
		date := tx.CreatedAt.Format("2006/01/02")

		creditorProfile, err := bot.GetGroupMemberProfile(in.GroupID, tx.CreditorID).Do()
		if err != nil {
			return linebot.NewTextMessage("債権者情報の取得に失敗しました。"), nil
		}

		var debtors []models.TransactionDebtor
		if err := infra.DB.Where("transaction_id = ?", tx.ID).Find(&debtors).Error; err != nil {
			return linebot.NewTextMessage("債務者情報の取得に失敗しました。"), nil
		}

		debtorNames := []string{}
		for _, debtor := range debtors {
			debtorProfile, err := bot.GetGroupMemberProfile(in.GroupID, debtor.DebtorID).Do()
			if err != nil {
				return linebot.NewTextMessage("債務者情報の取得に失敗しました。"), nil
			}
			debtorNames = append(debtorNames, "@"+debtorProfile.DisplayName)
		}

		msg += fmt.Sprintf("📌【%s】\n", date)
		msg += fmt.Sprintf("@%s\n↓\n", creditorProfile.DisplayName)
		msg += strings.Join(debtorNames, "\n") + "\n"
		msg += fmt.Sprintf("%s：%s円\n\n", tx.Note, utils.FormatAmount(tx.Amount))
	}

	if len(txs) == 0 {
		msg += "取引履歴はありません。"
	}

	return linebot.NewTextMessage(msg), nil
}