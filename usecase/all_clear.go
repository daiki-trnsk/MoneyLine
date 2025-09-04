package usecase

import (
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/infra"
	"github.com/daiki-trnsk/MoneyLine/models"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

// グループ内取引履歴の全削除
func AllClear(bot *linebot.Client, in dto.Incoming) linebot.SendingMessage {
	var txs []models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Find(&txs).Error; err != nil {
		return utils.LogAndReplyError(err, in, "Failed to get transaction")
	}

	if len(txs) == 0 {
		return linebot.NewTextMessage("取引履歴はありません。")
	}

	// TransactionDebtorもカスケード削除される
	if err := infra.DB.Where("group_id = ?", in.GroupID).Delete(&models.Transaction{}).Error; err != nil {
		return utils.LogAndReplyError(err, in, "Failed to delete all transactions")
	}

	return linebot.NewTextMessage("全履歴を削除しました。")
}
