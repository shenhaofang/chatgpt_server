package util

import "fmt"

func SliceUniqueInt64(s []int64) (u []int64) {

	m := make(map[int64]bool)

	for _, v := range s {
		if !m[v] {
			u = append(u, v)
		}

		m[v] = true
	}

	return
}

// SliceFmtStr slice 转换为
func SliceInt64FmtStr(i []int64) (s []string) {

	for _, v := range i {
		s = append(s, fmt.Sprintf("%d", v))
	}

	return
}
