package action

import (
	"fmt"
	"sort"

	"github.com/mgumz/vopher/pkg/archive"
)

// ListArchives prints all supported archive types to stdout
func ListArchives() {
	l := archive.SupportedArchives()
	sort.Strings(l)
	for _, suf := range l {
		fmt.Println(suf)
	}
}
