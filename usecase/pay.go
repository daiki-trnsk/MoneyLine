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
	amount, note, includeSelf, errValue := validateMessageFormat(in.Text)
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

	// デフォルトで送信者（債権者）を含める（メッセージ末尾に「自分抜き」があれば除外）
	if includeSelf {
		if !seen[creditorID] {
			debtorIDs = append(debtorIDs, creditorID)
			seen[creditorID] = true
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

	msgs := "記録しました！"
	return linebot.NewTextMessage(msgs)
}

// メッセージのバリデーション
func validateMessageFormat(text string) (int64, string, bool, error) {
	parts := strings.Fields(text)

	// デフォルトは送信者を含める
	includeSelf := true

	// 文頭に @マネリン が含まれているかチェック
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "@マネリン") {
		return 0, "", includeSelf, fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 メモ \n\n使い方を確認するには私のメンションのみ送信してください。")
	}

	// 自分抜きトークン検出
	if parts[len(parts)-1] == "自分抜き" {
		includeSelf = false
		if len(parts) < 5 {
			return 0, "", includeSelf, fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 メモ 自分抜き \n\n使い方を確認するには私のメンションのみ送信してください。")
		}
	} else {
		if len(parts) < 4 {
			return 0, "", includeSelf, fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 メモ \n\n使い方を確認するには私のメンションのみ送信してください。")
		}
	}

	// 金額をパース
	amountIdx := len(parts) - 2
	if !includeSelf {
		amountIdx = len(parts) - 3
	}
	amount, err := utils.ParseAmount(parts[amountIdx])
	if err != nil {
		return 0, "", includeSelf, fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 メモ \n\n使い方を確認するには私のメンションのみ送信してください。")
	}

	// 金額が0またはマイナスの場合のチェック
	if amount <= 0 {
		return 0, "", includeSelf, fmt.Errorf("金額は0より大きい値を入力してください。")
	}

	// メモを取得（amountの次のトークンから末尾まで。ただし '自分抜き' は除外）
	noteStart := amountIdx + 1
	noteEnd := len(parts)
	if !includeSelf {
		noteEnd = len(parts) - 1
	}
	if noteStart >= noteEnd {
		return 0, "", includeSelf, fmt.Errorf("メモが指定されていません。")
	}
	note := strings.Join(parts[noteStart:noteEnd], " ")

	return amount, note, includeSelf, nil
}
