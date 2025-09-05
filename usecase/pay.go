package usecase

import (
	"fmt"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"gorm.io/gorm"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/infra"
	"github.com/daiki-trnsk/MoneyLine/models"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

// 貸し借りを記録
func Pay(bot *linebot.Client, in dto.Incoming) linebot.SendingMessage {
	// メッセージのバリデーション
	amount, note, errValue := validateMessageFormat(in.Text)
	if errValue != nil {
		return linebot.NewTextMessage(errValue.Error())
	}

	// 送信者（債権者）
	creditorID := in.SenderID
	if creditorID == "" || len(in.Mentionees) < 2 {
		return linebot.NewTextMessage("記録に必要な情報が不足しています。")
	}

	botUserID := os.Getenv("MONEYLINE_BOT_ID")

	// 債務者を取得
	debtorIDs := []string{}
	seen := make(map[string]bool)
	for i := 1; i < len(in.Mentionees); i++ {
		userID := in.Mentionees[i].UserID
		if userID == botUserID {
			return linebot.NewTextMessage("文頭にのみマネリンをメンションしてください")
		}
		if !seen[userID] {
			debtorIDs = append(debtorIDs, userID)
			seen[userID] = true
		}
	}

	// トランザクション処理
	if err := infra.DB.Transaction(func(tx *gorm.DB) error {
		transaction := models.Transaction{
			CreditorID: creditorID,
			GroupID:    in.GroupID,
			Amount:     amount,
			Note:       note,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		for _, debtorID := range debtorIDs {
			txDebtor := models.TransactionDebtor{
				TransactionID: transaction.ID,
				DebtorID:      debtorID,
			}
			if err := tx.Create(&txDebtor).Error; err != nil {
				return fmt.Errorf("failed to create transaction debtor: %w", err)
			}
		}
		return nil
	}); err != nil {
		return utils.LogAndReplyError(err, in, "Transaction failed")
	}

	msgs := "記録しました！\n\n"

	profileCache := make(map[string]string)

	creditorName := utils.GetCachedProfileName(bot, in.GroupID, creditorID, profileCache)
	msgs += creditorName + "\n↓\n"

	// 債務者のプロフィールをキャッシュ経由で取得
	for _, debtorID := range debtorIDs {
		debtorName := utils.GetCachedProfileName(bot, in.GroupID, debtorID, profileCache)
		msgs += debtorName + "\n"
	}
	msgs += "\n" + note + "：" + utils.FormatAmount(amount) + "円"

	return linebot.NewTextMessage(msgs)
}

// メッセージのバリデーション
func validateMessageFormat(text string) (int64, string, error) {
	parts := strings.Fields(text)

	// 文頭に @マネリン が含まれているかチェック
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "@マネリン") {
		return 0, "", fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 メモ \n\n使い方を確認するには私のメンションのみ送信してください。")
	}

	// 必須要素の数をチェック
	if len(parts) < 4 {
		return 0, "", fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 メモ \n\n使い方を確認するには私のメンションのみ送信してください。")
	}

	// 金額をパース
	amount, err := utils.ParseAmount(parts[len(parts)-2])
	if err != nil {
		return 0, "", fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 メモ \n\n使い方を確認するには私のメンションのみ送信してください。")
	}

	// 金額が0またはマイナスの場合のチェック
	if amount <= 0 {
		return 0, "", fmt.Errorf("金額は0より大きい値を入力してください。")
	}

	// メモを取得
	note := strings.Join(parts[len(parts)-1:], " ")

	return amount, note, nil
}
