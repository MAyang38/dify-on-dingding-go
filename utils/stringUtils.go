package utils

import "strings"

// 判断字符串a 是否在list中
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// 判断字符串s是否包含任意关键词keywords
func ContainsKeywords(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}
