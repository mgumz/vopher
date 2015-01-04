package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
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

	OPT_STRIP_DIR     = 0
	OPT_POSTUPDATE    = 1
	OPT_POSTUPDATE_OS = 2
)

var PLUGIN_OPTS = []string{
	"strip=",
	"postupdate=",
	"postupdate." + runtime.GOOS,
}

type Plugin struct {
	name       string
	url        *neturl.URL
	strip_dir  int    // strip n dir-parts from archive-entries
	postupdate string // execute after 'update'-action
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
		if err = plugin.OptionsFromFields(fields[1:]); err != nil {
			return nil, fmt.Errorf("parsing optional fields: %q, plugin %q on line %d", err, name, lnumber)
		}

		plugins[name] = plugin
	}
	return
}

func (p *Plugin) OptionsFromFields(fields []string) error {

	for i := range fields {
		if strings.HasPrefix(fields[i], PLUGIN_OPTS[OPT_STRIP_DIR]) {
			strip, err := strconv.ParseUint((fields[i])[len(PLUGIN_OPTS[OPT_STRIP_DIR]):], 10, 8)
			if err == nil {
				p.strip_dir = int(strip)
			} else {
				return fmt.Errorf("strange 'strip' field")
			}
		} else if strings.HasPrefix(fields[i], PLUGIN_OPTS[OPT_POSTUPDATE]) && p.postupdate == "" {
			p.postupdate = fields[i][len(PLUGIN_OPTS[OPT_POSTUPDATE]):]
		} else if strings.HasPrefix(fields[i], PLUGIN_OPTS[OPT_POSTUPDATE_OS]) {
			p.postupdate = fields[i][len(PLUGIN_OPTS[OPT_POSTUPDATE_OS]):]
		}
	}

	if p.postupdate != "" {
		decoded, err := neturl.QueryUnescape(p.postupdate)
		if err != nil {
			return err
		}
		p.postupdate = decoded
	}

	return nil
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
