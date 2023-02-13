package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"unicode"
)

// StrEmpty 判断字符串是否为空, 同时更新字段
func StrEmpty(str *string) (ok bool) {
	*str = strings.Trim(*str, " ")
	return *str == ""
}

// MD5
func MD5(str string) (h string) {
	c := md5.New()
	c.Write([]byte(str))

	return hex.EncodeToString(c.Sum(nil))
}

// StrIsNEH 是否为数字、英语、汉语
func StrIsNEH(s string) (ok bool) {

	for _, v := range s {
		if !unicode.IsNumber(v) && !(v <= unicode.MaxASCII && unicode.IsLetter(v)) && !unicode.Is(unicode.Scripts["Han"], v) {
			return
		}
	}

	ok = true
	return
}

// FmtNEH 格式化为数字、英语、汉语
// except 保留字符
func StrFmtNEH(s string, except ...string) (f string) {

	var rs []rune
	for _, v := range s {
		if unicode.IsNumber(v) || (v <= unicode.MaxASCII && unicode.IsLetter(v)) || unicode.Is(unicode.Scripts["Han"], v) {
			rs = append(rs, v)
			continue
		}

		for _, vv := range except {
			if string(v) == vv {
				rs = append(rs, v)
				break
			}
		}
	}

	f = string(rs)
	return
}

// StrFmtNEHMul 格式化为数字、英语、汉语 - 批量
func StrFmtNEHMul(s []string) (f []string) {

	var wg sync.WaitGroup
	wg.Add(len(s))

	c := make(chan string, len(s))
	for _, v := range s {
		go func(v string) {
			defer wg.Done()

			item := StrFmtNEH(v)
			if item != "" {
				c <- item
			}
		}(v)
	}

	wg.Wait()

	close(c)
	for v := range c {
		f = append(f, v)
	}

	return
}

// StrLen 字符串长度；
// 两个拉丁字符长度=1
func StrLen(s string) (c int) {

	c = len([]rune(s))
	var lc int

	for _, v := range s {
		if v <= unicode.MaxLatin1 {
			lc++
		}
	}

	c -= lc / 2

	return
}

// 格式化数量()
func StrFmtCount(count int) string {
	// 小于1万，显示原值
	if count < 10000 {
		return fmt.Sprintf("%d", count)
	}

	// 小于100万，显示xx.y万 如53.4万
	if count < 1000000 {
		if count%10000 > 1000 {
			return fmt.Sprintf("%.1f万", float32(count)/10000)
		} else {
			return fmt.Sprintf("%d万", count/10000)
		}
	}

	// 大于9999,9999，显示999万
	if count >= 99999999 {
		return "9999万"
	}

	// 默认显示 32万
	return fmt.Sprintf("%d万", count/10000)
}
