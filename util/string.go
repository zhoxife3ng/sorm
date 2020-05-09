package util

import "strings"

func TitleCasedName(name string) string {
	newStr := make([]rune, 0)
	upNextChar := true
	name = strings.ToLower(name)
	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			if 'a' <= chr && chr <= 'z' {
				chr -= ('a' - 'A')
			}
		case chr == '_':
			upNextChar = true
			continue
		}
		newStr = append(newStr, chr)
	}
	return string(newStr)
}

func TitleSnakeName(name string) string {
	newStr := make([]rune, 0)
	for _, chr := range name {
		if 'A' <= chr && chr <= 'Z' {
			newStr = append(newStr, '_', chr+'a'-'A')
		} else {
			newStr = append(newStr, chr)
		}
	}
	return string(newStr[1:])
}
