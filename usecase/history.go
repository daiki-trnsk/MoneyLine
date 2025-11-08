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
func History(bot *linebot.Client, in dto.Incoming) linebot.SendingMessage {
	var txs []models.Transaction
	// 履歴表示は確定済みトランザクションのみを表示する（未確定は除外）
	if err := infra.DB.Where("group_id = ? AND confirmed_by IS NOT NULL AND confirmed_by <> ''", in.GroupID).Order("created_at asc").Find(&txs).Error; err != nil {
		return utils.LogAndReplyError(err, in, "Failed to order transaction")
	}

	txIDs := make([]string, len(txs))
	for i, tx := range txs {
		txIDs[i] = tx.ID.String()
	}

	var allDebtors []models.TransactionDebtor
	if len(txIDs) > 0 {
		if err := infra.DB.Where("transaction_id IN ?", txIDs).Find(&allDebtors).Error; err != nil {
			return utils.LogAndReplyError(err, in, "Failed to get transaction debtors")
		}
	}

	debtorMap := make(map[string][]models.TransactionDebtor)
	for _, debtor := range allDebtors {
		debtorMap[debtor.TransactionID.String()] = append(debtorMap[debtor.TransactionID.String()], debtor)
	}

	msg := "履歴\n\n"

	profileCache := make(map[string]string)

	for _, tx := range txs {
		date := tx.CreatedAt.Format("2006/01/02")

		creditorName := utils.GetCachedProfileName(bot, in.GroupID, tx.CreditorID, profileCache)

		debtors := debtorMap[tx.ID.String()]

		debtorNames := []string{}
		for _, debtor := range debtors {
			debtorName := utils.GetCachedProfileName(bot, in.GroupID, debtor.DebtorID, profileCache)
			debtorNames = append(debtorNames, debtorName)
		}

		msg += fmt.Sprintf("📌【%s】\n", date)
		msg += fmt.Sprintf("%s\n↓\n", creditorName)
		msg += strings.Join(debtorNames, "\n") + "\n"
		msg += fmt.Sprintf("%s：%s円\n\n", tx.Note, utils.FormatAmount(tx.Amount))
	}

	if len(txs) == 0 {
		msg += "取引履歴はありません。"
	}

	return linebot.NewTextMessage(msg)
}
