package action

import (
	"fmt"
	"sort"

	"github.com/mgumz/vopher/pkg/archive"
)

func ListArchives() {
	l := archive.SupportedArchives()
	sort.Strings(l)
	for _, suf := range l {
		fmt.Println(suf)
	}
	return
}
