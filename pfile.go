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

	neturl "net/url"
)

// most plugins are fetched from github. the github zip-files
// put the files into a subfolder like this:
//   vim-plugin/doc/plugin.txt
//   vim-plugin/README.txt
//
const DEFAULT_STRIP = 1

type Plugin struct {
	name      string
	url       *neturl.URL
	strip_dir int
}

func ScanPluginFile(name string) (map[string]Plugin, error) {

	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	return ScanPluginReader(file)
}

func ScanPluginReader(reader io.ReadCloser) (plugins map[string]Plugin, err error) {

	plugins = make(map[string]Plugin)
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
		if fields, skip = is_comment(strings.Fields(line)); skip {
			//log.Printf("skip %v", fields)
			continue
		}

		// TODO: uncaught situation:
		// http://example.com/bar.zip strip=1
		name := ""
		if len(fields) > 1 {
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
