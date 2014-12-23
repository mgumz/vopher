package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	neturl "net/url"
)

const (
	// most plugins are fetched from github. the github zip-files
	// put the files into a subfolder like this:
	//   vim-plugin/doc/plugin.txt
	//   vim-plugin/README.txt
	//
	DEFAULT_STRIP = 1

	BYTE_ORDER_MARK = '\uFEFF'
)

type Plugin struct {
	name      string
	url       *neturl.URL
	strip_dir int
}

func (pl *Plugin) String() string {
	return fmt.Sprintf("Plugin{%q, %q, strip=%d}",
		pl.name, pl.url.String(), pl.strip_dir)
}

type PluginList map[string]Plugin

func ScanPluginFile(name string) (PluginList, error) {

	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return ScanPluginReader(file)
}

func ScanPluginReader(reader io.ReadCloser) (plugins PluginList, err error) {

	plugins = make(PluginList)
	defer reader.Close()

	var (
		scanner = bufio.NewScanner(reader)
		line    string
		fields  []string
		skip    bool
		url     *neturl.URL
	)

	scanner.Split(bufio.ScanLines)

	for lnumber := 1; scanner.Scan(); lnumber++ {

		line = scanner.Text()

		if lnumber == 1 {
			line = strip_bom(line)
		}

		if fields, skip = is_comment(strings.Fields(line)); skip {
			continue
		}

		name := ""
		if len(fields) > 1 && !strings.Contains(fields[0], "://") {
			name, fields = fields[0], fields[1:]
		}
		if url, err = neturl.Parse(fields[0]); err != nil {
			log.Println("error:", name, ":", lnumber, "not an url", line)
			continue
		}

		name = first_not_empty(name, path.Base(url.Path))
		if _, skip = plugins[name]; skip {
			return nil, fmt.Errorf("existing plugin %q on line %d", name, lnumber)
		}

		// strip away .zip (or other archive-formats)
		if strings.HasSuffix(name, ".zip") {
			name = name[:len(name)-4]
		}

		plugin := Plugin{name: name, url: url, strip_dir: DEFAULT_STRIP}

		// parse optional arguments
		fields = fields[1:]
		for i := range fields {
			if strings.HasPrefix(fields[i], "strip=") {
				strip, err := strconv.ParseUint((fields[i])[6:], 10, 8)
				if err == nil {
					plugin.strip_dir = int(strip)
				} else {
					return nil, fmt.Errorf("strange 'strip' field on line %d", lnumber)
				}
			}
		}

		plugins[name] = plugin
	}
	return
}

func is_comment(fields []string) ([]string, bool) {
	return fields, (len(fields) == 0 || len(fields[0]) == 0 || fields[0][0] == '#')
}

func strip_bom(in string) string {
	if strings.IndexRune(in, BYTE_ORDER_MARK) == -1 {
		return in
	}
	return in[utf8.RuneLen(BYTE_ORDER_MARK):]
}
