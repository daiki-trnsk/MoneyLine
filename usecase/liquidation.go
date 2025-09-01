package usecase

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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

type node struct {
	id  string
	amt int64
}

// 貪欲清算
func SettleGreedy(bot *linebot.Client, in dto.Incoming) linebot.SendingMessage {
	var txs []models.Transaction
	if err := infra.DB.Where("group_id = ?", in.GroupID).Find(&txs).Error; err != nil {
		return utils.LogAndReplyError(err, in, "Failed to get transaction")
	}

	net := make(map[string]int64)

	// 各取引に紐づく債務者を取得
	for _, tx := range txs {
		var debtors []models.TransactionDebtor
		if err := infra.DB.Where("transaction_id = ?", tx.ID).Find(&debtors).Error; err != nil {
			return utils.LogAndReplyError(err, in, "Failed to get transaction debtor")
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

	var creditors, debtors []node
	for id, v := range net {
		if v > 0 {
			creditors = append(creditors, node{id, v})
		} else if v < 0 {
			debtors = append(debtors, node{id, -v})
		}
	}
	if len(creditors) == 0 || len(debtors) == 0 {
		return linebot.NewTextMessage("清算は不要です。")
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
	return linebot.NewTextMessage(b.String())
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
