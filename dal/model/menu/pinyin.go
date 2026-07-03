// Package menu 拼音工具函数
package menu

import (
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

var (
	// pinyinArgs 拼音转换参数：不带声调，小写
	pinyinArgs = pinyin.NewArgs()
)

func init() {
	pinyinArgs.Style = pinyin.Normal // 不带声调，如：gongbao
	pinyinArgs.Heteronym = false     // 不使用多音字
}

// ToPinyin 将中文转换为拼音（不带声调，小写）
// 例如："宫保鸡丁" -> "gongbaojiding"
func ToPinyin(text string) string {
	if text == "" {
		return ""
	}
	// 只转换中文字符，其他字符保留
	var result strings.Builder
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			// 中文字符转拼音
			py := pinyin.SinglePinyin(r, pinyinArgs)
			if len(py) > 0 {
				result.WriteString(py[0])
			}
		} else {
			// 非中文字符保留（如英文、数字）
			result.WriteRune(r)
		}
	}
	return strings.ToLower(result.String())
}

// ToPinyinInitials 提取中文的拼音首字母（非中文字符原样保留，统一小写）
// 例如："宫保鸡丁" -> "gbjd"，"Q弹鱼丸" -> "qdyw"
func ToPinyinInitials(text string) string {
	if text == "" {
		return ""
	}
	var result strings.Builder
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			py := pinyin.SinglePinyin(r, pinyinArgs)
			if len(py) > 0 && py[0] != "" {
				result.WriteByte(py[0][0])
			}
		} else {
			result.WriteRune(r)
		}
	}
	return strings.ToLower(result.String())
}

// HasChinese 判断字符串是否包含中文字符
func HasChinese(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// MatchPinyin 检查菜单名称是否匹配搜索关键词（支持中文和拼音）
// name: 菜单名称
// keyword: 搜索关键词（可能是中文或拼音）
func MatchPinyin(name, keyword string) bool {
	if keyword == "" {
		return true
	}

	keyword = strings.ToLower(strings.TrimSpace(keyword))
	nameLower := strings.ToLower(strings.TrimSpace(name))

	// 1. 直接包含匹配（中文或英文）
	if strings.Contains(nameLower, keyword) {
		return true
	}

	// 2. 如果关键词包含中文，将菜单名称转为拼音后匹配
	if HasChinese(keyword) {
		namePinyin := ToPinyin(name)
		keywordPinyin := ToPinyin(keyword)
		if strings.Contains(namePinyin, keywordPinyin) {
			return true
		}
	} else {
		// 3. 如果关键词是纯拼音，将菜单名称转为全拼后匹配
		namePinyin := ToPinyin(name)
		if strings.Contains(namePinyin, keyword) {
			return true
		}
		// 4. 拼音首字母匹配，如 "宫保鸡丁"->"gbjd"，关键词 "jd"/"gbjd" 均可命中
		nameInitials := ToPinyinInitials(name)
		if strings.Contains(nameInitials, keyword) {
			return true
		}
	}

	return false
}
