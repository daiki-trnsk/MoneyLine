package usecase

import (
	"fmt"
	"log"
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

	// txs 空ガード
	if len(txs) == 0 {
		return linebot.NewTextMessage("清算は不要です。")
	}

	// TransactionDebtor を一括取得
	var allDebtors []models.TransactionDebtor
	if err := infra.DB.
		Where("transaction_id IN ?", extractTransactionIDs(txs)).
		Order("transaction_id, created_at, id").
		Find(&allDebtors).Error; err != nil {
		return utils.LogAndReplyError(err, in, "Failed to get transaction debtors")
	}

	// transaction_id をキーにしたマップを作成
	debtorMap := make(map[string][]models.TransactionDebtor)
	for _, debtor := range allDebtors {
		debtorMap[debtor.TransactionID.String()] = append(debtorMap[debtor.TransactionID.String()], debtor)
	}

	net := make(map[string]int64)

	// 各取引に紐づく債務者を取得
	for _, tx := range txs {
		// 取引内の Debtor 重複除去（配分前）
		debtors := debtorMap[tx.ID.String()]
		uniq := make([]models.TransactionDebtor, 0, len(debtors))
		seen := make(map[string]struct{})
		for _, d := range debtors {
			if _, ok := seen[d.DebtorID]; ok {
				continue
			}
			seen[d.DebtorID] = struct{}{}
			uniq = append(uniq, d)
		}
		debtors = uniq

		debtorCount := int64(len(debtors))
		if debtorCount == 0 {
			continue
		}

		share := tx.Amount / debtorCount
		rem := tx.Amount % debtorCount

		net[tx.CreditorID] += tx.Amount

		// 余りを均等に配分
		for idx, debtor := range debtors {
			delta := share
			if int64(idx) < rem {
				delta++
			}
			net[debtor.DebtorID] -= delta
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

	// 検算
	var sumPos, sumNeg int64
	for _, c := range creditors {
		sumPos += c.amt
	}
	for _, d := range debtors {
		sumNeg += d.amt
	}
	if sumPos != sumNeg {
		return utils.LogAndReplyError(fmt.Errorf("internal imbalance: creditors=%d, debtors=%d", sumPos, sumNeg), in, "Internal imbalance detected")
	}

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

	// プロフィールキャッシュ
	profileCache := make(map[string]string)
	var b strings.Builder
	b.WriteString("清算方法\n\n")
	for i, t := range res {
		fromName := getCachedProfileName(bot, in.GroupID, t.From, profileCache)
		toName := getCachedProfileName(bot, in.GroupID, t.To, profileCache)
		b.WriteString(fmt.Sprintf("%s → %s \n %s円", fromName, toName, formatYen(t.Amt)))
		if i < len(res)-1 {
			b.WriteString("\n\n")
		}
	}
	return linebot.NewTextMessage(b.String())
}

func extractTransactionIDs(txs []models.Transaction) []string {
	ids := make([]string, len(txs))
	for i, tx := range txs {
		ids[i] = tx.ID.String()
	}
	return ids
}

func getCachedProfileName(bot *linebot.Client, groupID, userID string, cache map[string]string) string {
	if name, exists := cache[userID]; exists {
		return name
	}
	profile, err := bot.GetGroupMemberProfile(groupID, userID).Do()
	if err != nil {
		log.Println(err, "Failed to get profile for user: "+userID)
		return "@" + userID
	}
	name := safeName(profile)
	cache[userID] = name
	return name
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
	s := strconv.FormatInt(v, 10)
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	pre := n % 3
	if pre == 0 {
		pre = 3
	}
	b.WriteString(s[:pre])
	for i := pre; i < n; i += 3 {
		b.WriteString("," + s[i:i+3])
	}
	return b.String()
}
