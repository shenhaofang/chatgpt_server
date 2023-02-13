package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const OneDay = 24 * 60 * 60 * time.Second

// TimeStr unix时间戳转换字符串
func TimeStr(u int64) (str string) {
	now := time.Now()
	old := time.Unix(u, 0)

	sub := now.Sub(old)
	if sub < 60*time.Second {
		return "刚刚"
	}
	if sub < 3600*time.Second {
		return fmt.Sprintf("%.f分钟前", sub.Minutes())
	}

	if sub < OneDay {
		return fmt.Sprintf("%.f小时前", sub.Hours())
	}

	if sub < 2*OneDay {
		return "昨天"
	}

	if sub < 3*OneDay {
		return "前天"
	}

	if sub < 7*OneDay {
		return fmt.Sprintf("%.f天前", sub.Hours() / 24)
	}

	if old.Year() == now.Year() {
		return old.Format("01-02")
	}

	return old.Format("2006-01-02")
}

// TimeGetDate 获取日期。
// 格式为："2018-11-11"，年月日都可省略。
func TimeGetDate(d string) (t time.Time) {

	loc, _ := time.LoadLocation("Asia/Shanghai")
	t, _ = time.ParseInLocation("2006-01-02", d, loc)

	ds := strings.Split(d, "-")

	l := len(ds)
	now := time.Now()

	year, month, day := now.Date()

	var err error

	for i := l - 1; i >= 0; i-- {
		if i == l-1 {
			day, err = strconv.Atoi(ds[i])
			if err != nil {
				day = now.Day()
			}
		} else if i == l-2 {
			month1, err := strconv.Atoi(ds[i])
			if err == nil {
				month = time.Month(month1)
			}
		} else if i == l-3 {
			year, err = strconv.Atoi(ds[i])
			if err != nil {
				year = now.Year()
			}
		}
	}

	t = time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)

	return
}
