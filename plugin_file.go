package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	neturl "net/url"
	"os"
	"path"
	"strings"
	"unicode/utf8"
)

const (
	BYTE_ORDER_MARK = '\uFEFF'
)

// Parse parses a vopher-plugin file
func (plugins PluginList) Parse(reader io.ReadCloser) error {

	defer reader.Close()

	var (
		scanner = bufio.NewScanner(reader)
		line    string
		fields  []string
		skip    bool
	)

	scanner.Split(bufio.ScanLines)

	for lnumber := 1; scanner.Scan(); lnumber++ {

		line = scanner.Text()

		if lnumber == 1 {
			line = stripBom(line)
		}

		if fields, skip = isComment(strings.Fields(line)); skip {
			continue
		}

		name := ""
		if len(fields) > 1 && !strings.Contains(fields[0], "://") {
			name, fields = fields[0], fields[1:]
		}
		url, err := neturl.Parse(fields[0])
		if err != nil {
			log.Println("error:", name, ":", lnumber, "not an url", line)
			continue
		}

		name = firstNotEmpty(name, path.Base(url.Path))
		if _, skip = plugins[name]; skip {
			return fmt.Errorf("existing plugin %q on line %d", name, lnumber)
		}

		// strip away .zip (or other archive-formats)
		if ok, lenSuffix := isSupportedArchive(name); ok {
			name = name[:len(name)-lenSuffix]
		}

		plugin := Plugin{name: name, url: url}
		plugin.opts.stripDir = DEFAULT_STRIP
		if err = plugin.optionsFromFields(fields[1:]); err != nil {
			return fmt.Errorf("parsing optional fields: %q, plugin %q on line %d", err, name, lnumber)
		}

		plugins[name] = &plugin
	}
	return nil
}

func (plugins PluginList) ParseFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	return plugins.Parse(file)
}

func isComment(fields []string) ([]string, bool) {
	return fields, (len(fields) == 0 || len(fields[0]) == 0 || fields[0][0] == '#')
}

func stripBom(in string) string {
	if strings.IndexRune(in, BYTE_ORDER_MARK) == -1 {
		return in
	}
	return in[utf8.RuneLen(BYTE_ORDER_MARK):]
}
