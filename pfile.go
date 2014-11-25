package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	neturl "net/url"
)

type Plugin struct {
	name string
	url  *neturl.URL
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

		plugins[name] = Plugin{name: name, url: url}
	}
	return
}

func is_comment(fields []string) ([]string, bool) {
	return fields, (len(fields) == 0 || len(fields[0]) == 0 || fields[0][0] == '#')
}
