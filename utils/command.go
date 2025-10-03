package utils

import (
	"strings"

	"github.com/daiki-trnsk/MoneyLine/dto"
)

const (
	CmdPay      = "pay"
	CmdSummary  = "summary"
	CmdHistory  = "history"
	CmdOneClear = "one_clear"
	CmdAllClear = "all_clear"
	CmdHelp     = "help"
)

// 全角→半角、trimする関数
func norm(text string) string {
	// 全角英数字を半角に変換
	var result []rune
	for _, r := range text {
		// 全角英数字
		if r >= '！' && r <= '～' {
			r = rune(r - '！' + '!')
		}
		// 全角スペース
		if r == '　' {
			r = ' '
		}
		result = append(result, r)
	}
	return strings.TrimSpace(string(result))
}

// 数字が含まれているかチェックする関数（全角・半角対応）
func ContainsNumber(text string) bool {
	for _, r := range text {
		// 半角数字
		if r >= '0' && r <= '9' {
			return true
		}
		// 全角数字
		if r >= '０' && r <= '９' {
			return true
		}
	}
	return false
}

// コマンドを判別
func DetectCommand(in dto.Incoming) string {
	// マネリン以外のメンション+数字でPay処理
	if len(in.Mentionees) > 1 && ContainsNumber(in.Text) {
		return CmdPay
	}

	// Pay以外のコマンド判別
	t := norm(in.Text)
	switch {
	case strings.Contains(t, "一覧"):
		return CmdSummary
	case strings.Contains(t, "清算"), strings.Contains(t, "精算"), strings.Contains(t, "せいさん"):
		return CmdSummary
	case strings.Contains(t, "履歴"), strings.Contains(t, "りれき"):
		return CmdHistory
	case strings.Contains(t, "一件削除"):
		return CmdOneClear
	case strings.Contains(t, "全削除"):
		return CmdAllClear
	case strings.Contains(t, "使い方"), strings.Contains(t, "ヘルプ"):
		return CmdHelp
	// アカウント名に合わせる必要あり、あとで環境変数化
	case t == "@マネリン":
		return CmdHelp
	}
	return ""
}
