package plugin

import (
	"strings"
	"unicode/utf8"

	"github.com/mgumz/vopher/pkg/archive"
)

func isComment(fields []string) ([]string, bool) {

	return fields, (len(fields) == 0 || strings.HasPrefix(fields[0], "#"))
}

// the first fields is eventually the plugin name, the
// 2nd field is then the URL
func eventualName(fields []string) (string, []string) {
	if len(fields) > 1 && !strings.Contains(fields[0], "://") {
		return fields[0], fields[1:]
	}
	return "", fields
}

// strip away .zip (or other archive-formats)
func cleanName(name string) string {
	if ok, l := archive.IsSupportedArchive(name); ok {
		return name[:len(name)-l]
	}
	return name
}

func stripBom(in string) string {
	if !strings.ContainsRune(in, ByteOrderMark) {
		return in
	}
	return in[utf8.RuneLen(ByteOrderMark):]
}

func parseDependsOn(depends string) []string {

	deps := []string{}
	for _, d := range strings.Split(depends, ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			deps = append(deps, d)
		}
		// TODO: warn about empty dependency
	}
	return deps
}
