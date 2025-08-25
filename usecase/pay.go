package usecase

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"

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
	// 送信者（債権者）
	creditorID := in.SenderID

	// 「@マネリン @債務者 金額 メモ」を想定
	parts := strings.Fields(in.Text)
	amount := int64(0)
	note := ""
	for _, p := range parts {
		if strings.HasPrefix(p, "@") {
			continue
		}
		if amount == 0 {
			if a, err := utils.ParseAmount(p); err == nil {
				amount = a
				continue
			}
		}
		if amount > 0 {
			if note != "" {
				note += " "
			}
			note += p
		}
	}

	if creditorID == "" || len(in.Mentionees) < 2 || amount == 0 {
		return linebot.NewTextMessage("記録に必要な情報が不足しています。"), nil
	}

	msgs := "記録しました！\n" +
		note + " : " + utils.FormatAmount(amount) + "円\n\n" +
		"💰差引残高：\n"

	var balanceLines []string

	for i := 1; i < len(in.Mentionees); i++ {
		debtorID := in.Mentionees[i].UserID
		tx := models.Transaction{
			CreditorID: creditorID,
			DebtorID:   debtorID,
			GroupID:    in.GroupID,
			Amount:     amount,
			Note:       note,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := infra.DB.Create(&tx).Error; err != nil {
			msgs = ""
			msgs += "@" + debtorID + " 記録に失敗しました。\n"
			return linebot.NewTextMessage(msgs), nil
		}

		var txs []models.Transaction
		if err := infra.DB.Where(
			"group_id = ? AND ((creditor_id = ? AND debtor_id = ?) OR (creditor_id = ? AND debtor_id = ?))",
			in.GroupID, creditorID, debtorID, debtorID, creditorID,
		).Find(&txs).Error; err != nil {
			msgs = ""
			msgs += "@" + debtorID + " 記録しましたが、残高取得に失敗しました。\n"
			return linebot.NewTextMessage(msgs), nil
		}

		balance := int64(0)
		for _, t := range txs {
			if t.CreditorID == creditorID && t.DebtorID == debtorID {
				balance += t.Amount
			} else if t.CreditorID == debtorID && t.DebtorID == creditorID {
				balance -= t.Amount
			}
		}

		var upper, lower string
		var bal int64
		if balance >= 0 {
			upper = creditorID
			lower = debtorID
			bal = balance
		} else {
			upper = debtorID
			lower = creditorID
			bal = -balance
		}
		upperProfile, _ := bot.GetGroupMemberProfile(in.GroupID, upper).Do()
		lowerProfile, _ := bot.GetGroupMemberProfile(in.GroupID, lower).Do()
		balanceLines = append(balanceLines, upperProfile.DisplayName+" → "+lowerProfile.DisplayName+"\n"+utils.FormatAmount(bal)+"円")
	}

	if len(balanceLines) > 0 {
		msgs += strings.Join(balanceLines, "\n")
	}

	return linebot.NewTextMessage(msgs), nil
}

// 一覧（グループごとの債権債務集計）
func Summary(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	defer func() {
		if r := recover(); r != nil {
			bot.PushMessage(in.GroupID, linebot.NewTextMessage("一覧取得中に予期せぬエラーが発生しました")).Do()
		}
	}()

	var txs []models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Find(&txs).Error; err != nil {
		return linebot.NewTextMessage("一覧取得に失敗しました。"), nil
	}

	type pair struct {
		User1 string
		User2 string
	}

	balances := make(map[pair]int64)
	for _, tx := range txs {
		u1, u2 := tx.CreditorID, tx.DebtorID
		if u1 > u2 {
			u1, u2 = u2, u1
			balances[pair{u1, u2}] -= tx.Amount
		} else {
			balances[pair{u1, u2}] += tx.Amount
		}
	}

	msg := "💰未払い一覧\n\n"
	count := 0
	var lines []string
	for p, amount := range balances {
		if amount == 0 {
			continue
		}
		// amount > 0: User1がUser2に貸している
		// amount < 0: User2がUser1に貸している
		var upper, lower string
		var bal int64
		if amount > 0 {
			upper = p.User1
			lower = p.User2
			bal = amount
		} else {
			upper = p.User2
			lower = p.User1
			bal = -amount
		}
		upperProfile, _ := bot.GetGroupMemberProfile(in.GroupID, upper).Do()
		lowerProfile, _ := bot.GetGroupMemberProfile(in.GroupID, lower).Do()
		lines = append(lines, upperProfile.DisplayName+" → "+lowerProfile.DisplayName+"\n"+utils.FormatAmount(bal)+"円")
		count++
	}
	if count == 0 {
		msg += "現在、未払い情報はありません。"
	} else {
		msg += strings.Join(lines, "\n")
	}
	return linebot.NewTextMessage(msg), nil
}

// 貪欲清算
func SettleGreedy(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	// 1) 取引履歴 → ネット残高
	var txs []models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Find(&txs).Error; err != nil {
		return linebot.NewTextMessage("清算案の作成に失敗しました。"), nil
	}
	net := make(map[string]int64) // +受取 / -支払
	for _, tx := range txs {
		net[tx.CreditorID] += tx.Amount
		net[tx.DebtorID] -= tx.Amount
	}

	// 2) 債権者(+) / 債務者(-) に分割
	type node struct {
		id  string
		amt int64
	}
	var creditors, debtors []node
	for id, v := range net {
		if v > 0 {
			creditors = append(creditors, node{id, v})
		} else if v < 0 {
			debtors = append(debtors, node{id, -v}) // 正にして保持（払うべき額）
		}
	}
	if len(creditors) == 0 || len(debtors) == 0 {
		return linebot.NewTextMessage("清算は不要です。"), nil
	}

	// 3) 金額大きい順で貪欲に消し込み
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

	// 4) 出力
	var b strings.Builder
	b.WriteString("🧮 貪欲法清算テスト\n")
	b.WriteString(fmt.Sprintf("取引回数: %d 回\n\n", len(res)))
	for _, t := range res {
		from, _ := bot.GetGroupMemberProfile(in.GroupID, t.From).Do()
		to, _ := bot.GetGroupMemberProfile(in.GroupID, t.To).Do()
		b.WriteString(fmt.Sprintf("%s → %s: %s円\n",
			safeName(from), safeName(to), formatYen(t.Amt)))
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
	return p.DisplayName
}

func formatYen(v int64) string {
	// カンマ区切りにしたいならここを差し替え（utils側に合わせてOK）
	return strconv.FormatInt(v, 10)
}

// 履歴（グループごとの取引履歴）
func History(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	// var txs []models.Transaction
	// if err := infra.DB.Where("group_id = ?", in.GroupID).Order("created_at desc").Limit(10).Find(&txs).Error; err != nil {
	// 	return linebot.NewTextMessage("履歴取得に失敗しました。"), nil
	// }
	// msg := "📜 履歴（最新10件）\n"
	// for _, tx := range txs {
	// 	msg += tx.CreatedAt.Format("2006-01-02 15:04") + " @" + tx.CreditorID + "→@" + tx.DebtorID + " " + utils.FormatAmount(tx.Amount) + " " + tx.Note + "\n"
	// }
	msg := "History called"
	return linebot.NewTextMessage(msg), nil
}
