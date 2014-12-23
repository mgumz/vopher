package main

// idea: instead of having python/ruby/curl/wget/fetch/git installed
// for a vim-plugin-manager to fetch the plugins i just want one binary
// which does it all.
//
// plugins: http://vimawesome.com/

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var allowed_actions = []string{
	"u",
	"up",
	"update",
	"c",
	"clean",
	"sample",
}

func usage() {

	fmt.Fprintln(os.Stderr, `vopher - acquire vim-plugins the gopher-way

usage: vopher [flags] <action>

actions
  update - acquire the given plugins from the -f list
  clean  - remove given plugins frmo the -f list
  sample - print sample vopher.list to stdout

flags`)
	flag.PrintDefaults()
}

func main() {

	log.SetPrefix("vopher.")
	cli := struct {
		action string
		force  bool
		file   string
		dir    string
		ui     string
	}{action: "update", dir: ".", ui: "progressline"}

	flag.BoolVar(&cli.force, "force", cli.force, "if already existant: refetch plugins")
	flag.StringVar(&cli.file, "f", cli.file, "path to list of plugins")
	flag.StringVar(&cli.dir, "dir", cli.dir, "directory to extract the plugins to")
	flag.StringVar(&cli.ui, "ui", cli.ui, "ui mode")

	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) > 0 {
		cli.action = flag.Args()[0]
	}

	if prefix_in_stringslice(allowed_actions, cli.action) == -1 {
		log.Fatal("error: unknown action")
	}

	if cli.action == "sample" {
		act_sample()
		return
	}

	var ui JobUi
	switch cli.ui {
	case "progressline":
		ui = &UiOneLine{
			ProgressTicker: NewProgressTicker(0),
			prefix:         "vopher",
			duration:       25 * time.Millisecond,
		}
	case "simple":
		ui = &UiSimple{jobs: make(map[string]_ri)}
	}

	switch cli.action {
	case "update", "u", "up":
		plugins := must_read_plugins(cli.file)
		act_update(plugins, cli.dir, cli.force, ui)
	case "clean", "c", "cl":
		plugins := must_read_plugins(cli.file)
		act_clean(plugins, cli.dir, cli.force)
	}
}

func must_read_plugins(path string) PluginList {
	plugins, err := ScanPluginFile(path)
	if err != nil {
		log.Fatal(err)
	}

	if len(plugins) == 0 {
		log.Fatalf("empty plugin-file %q", path)
	}
	return plugins
}
