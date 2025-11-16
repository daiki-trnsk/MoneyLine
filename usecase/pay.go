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
	// メッセージのバリデーション（メモは不要。余分な末尾トークンは無視して金額を抽出）
	amount, includeSelf, errValue := validateMessageFormat(in.Text)
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
// メモは不要。末尾に余分なトークンがあってもエラーにせず、末尾から数値トークンを探して金額を抽出する。
// 戻り値: amount, includeSelf, error
func validateMessageFormat(text string) (int64, bool, error) {
	parts := strings.Fields(text)

	// デフォルトは送信者を含める
	includeSelf := true

	// 文頭に @マネリン が含まれているかチェック
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "@マネリン") {
		return 0, includeSelf, fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 \n\n使い方を確認するには私のメンションのみ送信してください。")
	}

	// 自分抜きトークン検出（末尾）
	endIdx := len(parts) - 1
	if parts[len(parts)-1] == "自分抜き" {
		includeSelf = false
		if len(parts) < 3 {
			return 0, includeSelf, fmt.Errorf("メッセージ形式が正しくありません。\n\n形式: @マネリン @借りた人(複数可) 金額 自分抜き")
		}
		endIdx = len(parts) - 2
	}

	// 末尾側から最初にパース可能な金額トークンを探す（メモは無視）
	var amount int64
	found := false
	for i := endIdx; i >= 1; i-- {
		a, err := utils.ParseAmount(parts[i])
		if err == nil {
			amount = a
			// 金額トークンは必ずメンションより後にあるべき（parts[0]=@マネリン, parts[1]=@...）
			if i < 2 {
				return 0, includeSelf, fmt.Errorf("金額が指定されていません。形式: @マネリン @借りた人 金額")
			}
			found = true
			break
		}
	}
	if !found {
		return 0, includeSelf, fmt.Errorf("金額が指定されていません。形式: @マネリン @借りた人 金額")
	}

	if amount <= 0 {
		return 0, includeSelf, fmt.Errorf("金額は0より大きい値を入力してください。")
	}

	return amount, includeSelf, nil
}
