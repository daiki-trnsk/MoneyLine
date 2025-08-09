package utils

import (
	"strings"
)

const (
	CmdSummary = "summary"
	CmdHistory = "history"
	CmdHelp    = "help"
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

func DetectCommand(text string) string {
	t := norm(text)
	switch {
	case strings.Contains(t, "一覧"):
		return CmdSummary
	case strings.Contains(t, "履歴"):
		return CmdHistory
	case strings.Contains(t, "使い方"), strings.Contains(t, "ヘルプ"):
		return CmdHelp
	case t == "":
		return CmdHelp
	}
	return ""
}
