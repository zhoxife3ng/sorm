package internal

import (
	"reflect"
	"strings"
	"unsafe"
)

func TitleCasedName(name string) string {
	newStr := make([]rune, 0)
	upNextChar := true
	name = strings.ToLower(name)
	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			if 'a' <= chr && chr <= 'z' {
				chr -= 'a' - 'A'
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
	firstChr := true
	for _, chr := range name {
		if 'A' <= chr && chr <= 'Z' {
			if !firstChr {
				newStr = append(newStr, '_')
			}
			newStr = append(newStr, chr+'a'-'A')
		} else {
			newStr = append(newStr, chr)
		}
		firstChr = false
	}
	return string(newStr)
}

func StringToBytes(s string) (b []byte) {
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
