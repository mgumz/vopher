package main

import (
	"log"
	stduser "os/user"
	"path/filepath"
	"strings"
)

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

func expand_path(p string) (string, error) {
	if p == "" {
		return p, nil
	}

	if p[0] == '~' {
		user, err := stduser.Current()
		if err != nil {
			return p, err
			log.Fatal("error: optaining current user?? %s", err)
		}
		p = filepath.Join(user.HomeDir, p[1:])
	}

	return p, nil
}
