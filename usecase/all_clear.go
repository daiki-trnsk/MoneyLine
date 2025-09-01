package usecase

import (
	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/infra"
	"github.com/daiki-trnsk/MoneyLine/models"
)

// グループ内取引履歴の全削除
func AllClear(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	var txs []models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Find(&txs).Error; err != nil {
		return linebot.NewTextMessage("全履歴の取得に失敗しました。"), nil
	}

	if len(txs) == 0 {
		return linebot.NewTextMessage("取引履歴はありません。"), nil
	}

	// 各トランザクションに紐づくTransactionDebtorを削除
	for _, tx := range txs {
		if err := infra.DB.Where("transaction_id = ?", tx.ID).Delete(&models.TransactionDebtor{}).Error; err != nil {
			return linebot.NewTextMessage("関連する債務者の削除に失敗しました。"), nil
		}
	}

	// トランザクション自体を削除
	if err := infra.DB.Where("group_id = ?", in.GroupID).Delete(&models.Transaction{}).Error; err != nil {
		return linebot.NewTextMessage("全履歴の削除に失敗しました。"), nil
	}

	return linebot.NewTextMessage("全履歴を削除しました。"), nil
}
