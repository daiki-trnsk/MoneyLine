package usecase

import (
	"sort"
	"testing"

	"github.com/daiki-trnsk/MoneyLine/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// 転送結果の安定ソート（順序依存のフレーク回避）
func sortTransfers(ts []transfer) {
	sort.Slice(ts, func(i, j int) bool {
		if ts[i].From != ts[j].From {
			return ts[i].From < ts[j].From
		}
		if ts[i].To != ts[j].To {
			return ts[i].To < ts[j].To
		}
		return ts[i].Amt < ts[j].Amt
	})
}

func TestCalculateSettlement(t *testing.T) {
	tests := []struct {
		name      string
		txs       []models.Transaction
		debtors   []models.TransactionDebtor
		expected  []transfer
		expectNil bool
		expectErr bool
	}{
		{
			name:      "空入力",
			txs:       nil,
			debtors:   nil,
			expectNil: true,
		},
		{
			name:      "取引・債務者ゼロ → 清算不要",
			txs:       []models.Transaction{},
			debtors:   []models.TransactionDebtor{},
			expectNil: true,
		},
		{
			name: "単一債務者",
			txs: []models.Transaction{
				{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), CreditorID: "A", Amount: 1000},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), DebtorID: "B"},
			},
			expected: []transfer{
				{From: "B", To: "A", Amt: 1000},
			},
		},
		{
			name: "複数債務者（割り切れる）",
			txs: []models.Transaction{
				{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), CreditorID: "A", Amount: 2000},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), DebtorID: "B"},
				{TransactionID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), DebtorID: "C"},
			},
			expected: []transfer{
				{From: "B", To: "A", Amt: 1000},
				{From: "C", To: "A", Amt: 1000},
			},
		},
		{
			name: "複数債務者（割り切れない）",
			txs: []models.Transaction{
				{ID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), CreditorID: "A", Amount: 1001},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), DebtorID: "B"},
				{TransactionID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), DebtorID: "C"},
			},
			expected: []transfer{
				{From: "B", To: "A", Amt: 501},
				{From: "C", To: "A", Amt: 500},
			},
		},
		{
			name: "重複メンションは無視（1人分のみ）",
			txs: []models.Transaction{
				{ID: uuid.MustParse("44444444-4444-4444-4444-444444444444"), CreditorID: "A", Amount: 1000},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("44444444-4444-4444-4444-444444444444"), DebtorID: "B"},
				{TransactionID: uuid.MustParse("44444444-4444-4444-4444-444444444444"), DebtorID: "B"}, // duplicate
			},
			expected: []transfer{
				{From: "B", To: "A", Amt: 1000},
			},
		},
		{
			name: "債権者が債務者に含まれる（現仕様：含めて均等割）",
			txs: []models.Transaction{
				{ID: uuid.MustParse("55555555-5555-5555-5555-555555555555"), CreditorID: "A", Amount: 3000},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("55555555-5555-5555-5555-555555555555"), DebtorID: "A"},
				{TransactionID: uuid.MustParse("55555555-5555-5555-5555-555555555555"), DebtorID: "B"},
			},
			expected: []transfer{
				{From: "B", To: "A", Amt: 1500},
			},
		},
		// 入力側で弾くのでいったんスキップ
		// {
		// 	name: "0円・負の取引はスキップ → 清算不要",
		// 	txs: []models.Transaction{
		// 		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666666"), CreditorID: "A", Amount: 0},
		// 		{ID: uuid.MustParse("77777777-7777-7777-7777-777777777777"), CreditorID: "A", Amount: -100},
		// 	},
		// 	debtors: []models.TransactionDebtor{
		// 		{TransactionID: uuid.MustParse("66666666-6666-6666-6666-666666666666"), DebtorID: "B"},
		// 		{TransactionID: uuid.MustParse("77777777-7777-7777-7777-777777777777"), DebtorID: "B"},
		// 	},
		// 	expectNil: true,
		// },
		{
			name: "債務者ゼロの取引はスキップ → 清算不要",
			txs: []models.Transaction{
				{ID: uuid.MustParse("88888888-8888-8888-8888-888888888888"), CreditorID: "A", Amount: 1000},
			},
			debtors:   []models.TransactionDebtor{}, // none for that tx
			expectNil: true,
		},
		{
			name: "三人に1001（334,334,333）- 端数2を先頭2人へ",
			txs: []models.Transaction{
				{ID: uuid.MustParse("99999999-9999-9999-9999-999999999999"), CreditorID: "A", Amount: 1001},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("99999999-9999-9999-9999-999999999999"), DebtorID: "B"},
				{TransactionID: uuid.MustParse("99999999-9999-9999-9999-999999999999"), DebtorID: "C"},
				{TransactionID: uuid.MustParse("99999999-9999-9999-9999-999999999999"), DebtorID: "D"},
			},
			expected: []transfer{
				{From: "B", To: "A", Amt: 334},
				{From: "C", To: "A", Amt: 334},
				{From: "D", To: "A", Amt: 333},
			},
		},
		{
			name: "複数取引の相殺（A:+1600, B:-300, C:-1300）",
			txs: []models.Transaction{
				{ID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), CreditorID: "A", Amount: 3000}, // A→(B,C):3000
				{ID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), CreditorID: "B", Amount: 1200}, // B→(A,C):1200
				{ID: uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"), CreditorID: "C", Amount: 800},  // C→(A):800
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), DebtorID: "B"},
				{TransactionID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), DebtorID: "C"},
				{TransactionID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), DebtorID: "A"},
				{TransactionID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), DebtorID: "C"},
				{TransactionID: uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"), DebtorID: "A"},
			},
			// 貪欲：C(1300)→A(1600) 1300、B(300)→A(300) 300 の順が自然
			expected: []transfer{
				{From: "C", To: "A", Amt: 1300},
				{From: "B", To: "A", Amt: 300},
			},
		},
		{
			name: "複数の債権者 vs 1人の債務者",
			txs: []models.Transaction{
				{ID: uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"), CreditorID: "A", Amount: 500},
				{ID: uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"), CreditorID: "C", Amount: 200},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"), DebtorID: "B"},
				{TransactionID: uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"), DebtorID: "B"},
			},
			expected: []transfer{
				{From: "B", To: "A", Amt: 500},
				{From: "B", To: "C", Amt: 200},
			},
		},
		{
			name: "完全相殺 → 清算不要",
			txs: []models.Transaction{
				{ID: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), CreditorID: "A", Amount: 100},
				{ID: uuid.MustParse("10101010-1010-1010-1010-101010101010"), CreditorID: "B", Amount: 100},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), DebtorID: "B"},
				{TransactionID: uuid.MustParse("10101010-1010-1010-1010-101010101010"), DebtorID: "A"},
			},
			expectNil: true,
		},
		{
			name: "大きな金額と複数人（999,999 を3人で割る）",
			txs: []models.Transaction{
				{ID: uuid.MustParse("12121212-1212-1212-1212-121212121212"), CreditorID: "A", Amount: 999_999},
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("12121212-1212-1212-1212-121212121212"), DebtorID: "X1"},
				{TransactionID: uuid.MustParse("12121212-1212-1212-1212-121212121212"), DebtorID: "X2"},
				{TransactionID: uuid.MustParse("12121212-1212-1212-1212-121212121212"), DebtorID: "X3"},
			},
			expected: []transfer{
				{From: "X1", To: "A", Amt: 333_333},
				{From: "X2", To: "A", Amt: 333_333},
				{From: "X3", To: "A", Amt: 333_333},
			},
		},
		{
			name: "中規模複雑（余り・相殺混在）",
			txs: []models.Transaction{
				{ID: uuid.MustParse("13131313-1313-1313-1313-131313131313"), CreditorID: "A", Amount: 1000}, // A→(B,C,D): 333,334,333
				{ID: uuid.MustParse("14141414-1414-1414-1414-141414141414"), CreditorID: "B", Amount: 200},  // B→(A):200
			},
			debtors: []models.TransactionDebtor{
				{TransactionID: uuid.MustParse("13131313-1313-1313-1313-131313131313"), DebtorID: "B"},
				{TransactionID: uuid.MustParse("13131313-1313-1313-1313-131313131313"), DebtorID: "C"},
				{TransactionID: uuid.MustParse("13131313-1313-1313-1313-131313131313"), DebtorID: "D"},
				{TransactionID: uuid.MustParse("14141414-1414-1414-1414-141414141414"), DebtorID: "A"},
			},
			// ネット：A +800, B -134, C -333, D -333 → 貪欲で A 800 を C/D/B の順に 333/333/134 で回収
			// （順序は安定ソートで吸収するので配列順は問わない）
			expected: []transfer{
				{From: "B", To: "A", Amt: 134},
				{From: "C", To: "A", Amt: 333},
				{From: "D", To: "A", Amt: 333},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateSettlement(tt.txs, tt.debtors)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			sortTransfers(result)
			sortTransfers(tt.expected)

			assert.Equal(t, tt.expected, result)
		})
	}
}
