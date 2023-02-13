package util

import (
	"strconv"
)

// Mask2Id mask_id 转化 article_id
func Mask2Id(maskId string) (id int64) {

	// 向前兼容
	id, err := strconv.ParseInt(maskId, 10, 64)
	if err == nil && id < 1161 {
		return id
	}

	id, err = strconv.ParseInt(maskId, 36, 64)
	if err != nil {
		return id
	}

	return id / 1000
}

// SliceStrUnique 是否无重复
func SliceStrUnique(s []string) bool {

	m := make(map[interface{}]int)

	for _, v := range s {
		m[v]++
		if m[v] > 1 {
			return false
		}
	}

	return true
}
