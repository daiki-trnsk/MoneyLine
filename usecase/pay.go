package usecase

import (
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"

	"github.com/daiki-trnsk/MoneyLine/dto"
	"github.com/daiki-trnsk/MoneyLine/infra"
	"github.com/daiki-trnsk/MoneyLine/models"
	"github.com/daiki-trnsk/MoneyLine/utils"
)

// 貸し借りを記録
func Pay(bot *linebot.Client, in dto.Incoming) (*linebot.TextMessage, error) {
	// 送信者（債権者）
	creditorID := in.SenderID

	// 「@マネリン @債務者 金額 メモ」を想定
	parts := strings.Fields(in.Text)
	amount := 0.0
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

		balance := 0.0
		for _, t := range txs {
			if t.CreditorID == creditorID && t.DebtorID == debtorID {
				balance += t.Amount
			} else if t.CreditorID == debtorID && t.DebtorID == creditorID {
				balance -= t.Amount
			}
		}

		var upper, lower string
		var bal float64
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

	balances := make(map[pair]float64)
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
		var bal float64
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
