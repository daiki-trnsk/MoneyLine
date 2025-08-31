package usecase

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"gorm.io/gorm"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/infra"
	"github.com/daiki-trnsk/MoneyLine/models"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

type transfer struct {
	From string
	To   string
	Amt  int64
}

// 貸し借りを記録
func Pay(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	// メッセージのバリデーション
	amount, note, err := validateMessageFormat(in.Text)
	if err != nil {
		return linebot.NewTextMessage(err.Error()), nil
	}

	// 送信者（債権者）
	creditorID := in.SenderID
	if creditorID == "" || len(in.Mentionees) < 2 {
		return linebot.NewTextMessage("記録に必要な情報が不足しています。"), nil
	}

	botUserID := os.Getenv("MONEYLINE_BOT_ID")

	// 債務者を取得
	debtorIDs := []string{}
	seen := make(map[string]bool)
	for i := 1; i < len(in.Mentionees); i++ {
		userID := in.Mentionees[i].UserID
		if userID == botUserID {
			return linebot.NewTextMessage("文頭にのみマネリンをメンションしてください"), nil
		}
		if !seen[userID] {
			debtorIDs = append(debtorIDs, userID)
			seen[userID] = true
		}
	}

	// 取引を作成
	tx := models.Transaction{
		CreditorID: creditorID,
		GroupID:    in.GroupID,
		Amount:     amount,
		Note:       note,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := infra.DB.Create(&tx).Error; err != nil {
		return linebot.NewTextMessage("取引の記録に失敗しました。"), nil
	}

	// 債務者を登録
	for _, debtorID := range debtorIDs {
		txDebtor := models.TransactionDebtor{
			TransactionID: tx.ID,
			DebtorID:      debtorID,
		}
		if err := infra.DB.Create(&txDebtor).Error; err != nil {
			return linebot.NewTextMessage("債務者の記録に失敗しました。"), nil
		}
	}

	// メッセージ作成
	msgs := "記録しました！\n\n"

	creditorProfile, err := bot.GetGroupMemberProfile(in.GroupID, creditorID).Do()
	if err != nil {
		return linebot.NewTextMessage("債権者情報の取得に失敗しました。"), nil
	}
	msgs += "@" + creditorProfile.DisplayName + "\n↓\n"
	for _, debtorID := range debtorIDs {
		debtorProfile, err := bot.GetGroupMemberProfile(in.GroupID, debtorID).Do()
		if err != nil {
			return linebot.NewTextMessage("債務者情報の取得に失敗しました。"), nil
		}
		msgs += "@" + debtorProfile.DisplayName + "\n"
	}
	msgs += "\n" + note + "：" + utils.FormatAmount(amount) + "円"

	return linebot.NewTextMessage(msgs), nil
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

	// メモを取得
	note := strings.Join(parts[len(parts)-1:], " ")

	return amount, note, nil
}

// 貪欲清算
func SettleGreedy(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	var txs []models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Find(&txs).Error; err != nil {
		return linebot.NewTextMessage("清算案の作成に失敗しました。"), nil
	}

	net := make(map[string]int64)

	// 各取引に紐づく債務者を取得
	for _, tx := range txs {
		var debtors []models.TransactionDebtor
		if err := infra.DB.Where("transaction_id = ?", tx.ID).Find(&debtors).Error; err != nil {
			return linebot.NewTextMessage("債務者情報の取得に失敗しました。"), nil
		}

		// 債務者ごとに金額を均等に分割
		debtorCount := int64(len(debtors))
		if debtorCount == 0 {
			continue
		}
		share := tx.Amount / debtorCount

		// 債権者に加算、債務者に減算
		net[tx.CreditorID] += tx.Amount
		for _, debtor := range debtors {
			net[debtor.DebtorID] -= share
		}
	}

	type node struct {
		id  string
		amt int64
	}
	var creditors, debtors []node
	for id, v := range net {
		if v > 0 {
			creditors = append(creditors, node{id, v})
		} else if v < 0 {
			debtors = append(debtors, node{id, -v})
		}
	}
	if len(creditors) == 0 || len(debtors) == 0 {
		return linebot.NewTextMessage("清算は不要です。"), nil
	}

	sort.Slice(creditors, func(i, j int) bool { return creditors[i].amt > creditors[j].amt })
	sort.Slice(debtors, func(i, j int) bool { return debtors[i].amt > debtors[j].amt })

	var res []transfer
	i, j := 0, 0
	for i < len(creditors) && j < len(debtors) {
		pay := min64(creditors[i].amt, debtors[j].amt)
		res = append(res, transfer{
			From: debtors[j].id,
			To:   creditors[i].id,
			Amt:  pay,
		})
		creditors[i].amt -= pay
		debtors[j].amt -= pay
		if creditors[i].amt == 0 {
			i++
		}
		if debtors[j].amt == 0 {
			j++
		}
	}

	var b strings.Builder
	b.WriteString("清算方法\n\n")
	for i, t := range res {
		from, _ := bot.GetGroupMemberProfile(in.GroupID, t.From).Do()
		to, _ := bot.GetGroupMemberProfile(in.GroupID, t.To).Do()
		b.WriteString(fmt.Sprintf("%s → %s \n %s円", safeName(from), safeName(to), formatYen(t.Amt)))
		if i < len(res)-1 {
			b.WriteString("\n\n")
		}
	}
	return linebot.NewTextMessage(b.String()), nil
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func safeName(p *linebot.UserProfileResponse) string {
	if p == nil {
		return "（不明）"
	}
	return "@" + p.DisplayName
}

func formatYen(v int64) string {
	return strconv.FormatInt(v, 10)
}

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

// グループ内取引履歴の最新一件削除
func OneClear(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	var tx models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Order("created_at desc").Last(&tx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return linebot.NewTextMessage("取引履歴はありません。"), nil
		}
		return linebot.NewTextMessage("最新の立替一件の取得に失敗しました。"), nil
	}

	// トランザクションに紐づくTransactionDebtorを削除
	if err := infra.DB.Where("transaction_id = ?", tx.ID).Delete(&models.TransactionDebtor{}).Error; err != nil {
		return linebot.NewTextMessage("関連する債務者の削除に失敗しました。"), nil
	}

	// トランザクション自体を削除
	if err := infra.DB.Delete(&tx).Error; err != nil {
		return linebot.NewTextMessage("最新の立替一件の削除に失敗しました。"), nil
	}

	return linebot.NewTextMessage("最新の立替一件を削除しました。"), nil
}

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
