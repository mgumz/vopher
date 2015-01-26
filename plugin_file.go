package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
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
	BYTE_ORDER_MARK = '\uFEFF'
)

var PLUGIN_OPTS = []string{
	"strip=",
	"postupdate=",
	"postupdate." + runtime.GOOS + "=",
	"sha1=",
}

type PluginList map[string]*Plugin

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
		if ok, len_suffix := IsSupportedArchive(name); ok {
			name = name[:len(name)-len_suffix]
		}

		plugin := Plugin{name: name, url: url}
		plugin.opts.strip_dir = DEFAULT_STRIP
		if err = plugin.OptionsFromFields(fields[1:]); err != nil {
			return nil, fmt.Errorf("parsing optional fields: %q, plugin %q on line %d", err, name, lnumber)
		}

		plugins[name] = &plugin
	}
	return
}

func (p *Plugin) OptionsFromFields(fields []string) error {

	for _, field := range fields {
		if strings.HasPrefix(field, PLUGIN_OPTS[OPT_STRIP_DIR]) {
			strip, err := strconv.ParseUint(field[len(PLUGIN_OPTS[OPT_STRIP_DIR]):], 10, 8)
			if err == nil {
				p.opts.strip_dir = int(strip)
			} else {
				return fmt.Errorf("strange 'strip' field")
			}
		} else if strings.HasPrefix(field, PLUGIN_OPTS[OPT_POSTUPDATE]) && p.opts.postupdate == "" {
			p.opts.postupdate = field[len(PLUGIN_OPTS[OPT_POSTUPDATE]):]
		} else if strings.HasPrefix(field, PLUGIN_OPTS[OPT_POSTUPDATE_OS]) {
			p.opts.postupdate = field[len(PLUGIN_OPTS[OPT_POSTUPDATE_OS]):]
		} else if strings.HasPrefix(field, PLUGIN_OPTS[OPT_SHA1]) {
			p.opts.sha1 = strings.ToLower(field[len(PLUGIN_OPTS[OPT_SHA1]):])
		}
	}

	if p.opts.postupdate != "" {
		decoded, err := neturl.QueryUnescape(p.opts.postupdate)
		if err != nil {
			return err
		}
		p.opts.postupdate = decoded
	}

	if p.opts.sha1 != "" && len(p.opts.sha1) != hex.EncodedLen(sha1.Size) {
		return fmt.Errorf("'sha1' field does not match size of a sha1")
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
