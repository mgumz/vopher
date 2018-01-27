package main

import (
	"fmt"
	"sort"
)

func actListArchives() {
	sort.Strings(supportedArchives)
	for _, suf := range supportedArchives {
		fmt.Println(suf)
	}
	return
}
