package usecase

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
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

	// 一意な request id を生成して保留トランザクションを作成（後で postback で確定）
	requestID := uuid.New().String()
	transaction := models.Transaction{
		CreditorID: creditorID,
		GroupID:    in.GroupID,
		Amount:     amount,
		Note:       note,
		RequestID:  requestID,
		DebtorIDs:  strings.Join(debtorIDs, ","),
		// ConfirmedBy は未設定（確定時に埋める）
	}

	if err := infra.DB.Create(&transaction).Error; err != nil {
		return utils.LogAndReplyError(err, in, "Transaction create failed")
	}

	// Quick Reply (postback) で確認を促す
	qr := linebot.NewTextMessage("あなたを含めて割り勘しますか？").WithQuickReplies(
		linebot.NewQuickReplyItems(
			linebot.NewQuickReplyButton(
				"",
				linebot.NewPostbackAction("はい", fmt.Sprintf("act=incl&v=1&tx=%s", requestID), "", "", "", ""),
			),
			linebot.NewQuickReplyButton(
				"",
				linebot.NewPostbackAction("いいえ", fmt.Sprintf("act=incl&v=0&tx=%s", requestID), "", "", "", ""),
			),
		),
	)
	return qr
}

// ConfirmPendingTransaction は postback を受けて確定処理を実行する
func ConfirmPendingTransaction(bot *linebot.Client, in dto.Incoming) linebot.SendingMessage {
	// postback.data 例: act=incl&v=1&tx=<requestId>
	if in.PostbackData == "" {
		return nil
	}
	// 最小限のパース
	parts := strings.Split(in.PostbackData, "&")
	m := map[string]string{}
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	if m["act"] != "incl" || m["tx"] == "" {
		return nil
	}
	includeSelf := m["v"] == "1"
	requestID := m["tx"]

	// 対象トランザクションを取得
	var t models.Transaction
	if err := infra.DB.Where("request_id = ?", requestID).First(&t).Error; err != nil {
		// 対象なし（既に消えた等）は黙殺
		return nil
	}

	didConfirm := false
	err := infra.DB.Transaction(func(tx *gorm.DB) error {
		// ConfirmedBy を書き込む（他と競合したら RowsAffected==0 => 先に確定済み）
		res := tx.Model(&models.Transaction{}).
			Where("request_id = ? AND (confirmed_by = '' OR confirmed_by IS NULL)", requestID).
			Update("confirmed_by", in.SenderID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// 既に誰かが確定済み -> 冪等性を保って何もしない
			return nil
		}
		// 確定フラグが立った
		didConfirm = true

		// 確定処理: DebtorIDs を展開して TransactionDebtor を作成
		debtorIDs := []string{}
		if strings.TrimSpace(t.DebtorIDs) != "" {
			for _, id := range strings.Split(t.DebtorIDs, ",") {
				id = strings.TrimSpace(id)
				if id != "" {
					debtorIDs = append(debtorIDs, id)
				}
			}
		}
		// includeSelf の場合、送信者（creditor）を債務者リストに追加（重複チェック）
		if includeSelf {
			found := false
			for _, d := range debtorIDs {
				if d == t.CreditorID {
					found = true
					break
				}
			}
			if !found {
				debtorIDs = append(debtorIDs, t.CreditorID)
			}
		}

		// TransactionDebtor を作成
		for _, did := range debtorIDs {
			td := models.TransactionDebtor{
				TransactionID: t.ID,
				DebtorID:      did,
			}
			if err := tx.Create(&td).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return utils.LogAndReplyError(err, in, "Failed to confirm transaction")
	}

	// 確定が成功した場合のみグループへ「記録しました」を返す（再送時は nil）
	if didConfirm {
		return linebot.NewTextMessage("記録しました")
	}
	return nil
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
