package usecase

import (
	"errors"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"gorm.io/gorm"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/infra"
	"github.com/daiki-trnsk/MoneyLine/models"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

// グループ内取引履歴の最新一件削除
func OneClear(bot *linebot.Client, in dto.Incoming) linebot.SendingMessage {
	var tx models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Order("created_at desc").Last(&tx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return linebot.NewTextMessage("取引履歴はありません。")
		}
		return utils.LogAndReplyError(err, in, "Failed to order transaction")
	}

	// トランザクションに紐づくTransactionDebtorを削除
	if err := infra.DB.Where("transaction_id = ?", tx.ID).Delete(&models.TransactionDebtor{}).Error; err != nil {
		return utils.LogAndReplyError(err, in, "Failed to delete related transaction debtor")
	}

	// トランザクション自体を削除
	if err := infra.DB.Delete(&tx).Error; err != nil {
		return utils.LogAndReplyError(err, in, "Failed to delete transaction")
	}

	return linebot.NewTextMessage("最新の立替一件を削除しました。")
}