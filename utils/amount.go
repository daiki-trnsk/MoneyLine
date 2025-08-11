package utils

import (
	"strconv"
	"strings"
	"unicode"
)

// 全角数字→半角数字変換
func toHalfWidth(s string) string {
	var result strings.Builder
	for _, r := range s {
		// 全角数字
		if r >= '０' && r <= '９' {
			result.WriteRune(r - '０' + '0')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ParseAmount: 文字列から金額(float64)を抽出
func ParseAmount(s string) (float64, error) {
	s = toHalfWidth(s)
	// 末尾の「円」や空白を除去
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "円")
	// 半角数字のみ抽出
	var digits strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) || r == '.' {
			digits.WriteRune(r)
		}
	}
	value, err := strconv.ParseFloat(digits.String(), 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// FormatAmount: 金額をカンマ区切りで文字列化
func FormatAmount(a float64) string {
	return strconv.FormatFloat(a, 'f', 0, 64)
}
