package main

import "strings"

func first_not_empty(parts ...string) (result string) {
	for i := range parts {
		if len(parts[i]) > 0 {
			return parts[i]
		}
	}
	return
}

func index_byte_n(path string, needle byte, n int) int {

	idx, s := 0, 0
	for ; s < n; s++ {

		i := strings.IndexByte(path[idx:], needle)
		if i < 0 { // not found
			break
		}

		idx = idx + i + 1
	}

	if s < n {
		idx = 0
	}

	return idx - 1
}

func prefix_in_stringslice(lst []string, s string) int {
	for i := range lst {
		if strings.HasPrefix(lst[i], s) {
			return i
		}
	}
	return -1
}
