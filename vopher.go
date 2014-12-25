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
	"check",
	"prune",
	"sample",
	"status",
	"st",
}

func usage() {

	fmt.Fprintln(os.Stderr, `vopher - acquire vim-plugins the gopher-way

usage: vopher [flags] <action>

actions
  update - acquire the given plugins from the -f <list>
  check  - check plugins from -f <list> against a more
           recent version
  clean  - remove given plugins from the -f <list>
  prune  - remove all entries from -dir <folder>
           which are not referenced in -f <list>.
           use -force=true to actually delete entries.
           use -all=true to also delete <plugin>.zip
           entries.
  status - lists plugins in -dir <folder> and marks missing or
           referenced and unreferenced plugins accordingly
  sample - print sample vopher.list to stdout

flags`)
	flag.PrintDefaults()
}

func main() {

	log.SetPrefix("vopher.")
	cli := struct {
		action string
		force  bool
		all    bool
		file   string
		dir    string
		ui     string
	}{action: "update", dir: ".", ui: "progressline"}

	flag.BoolVar(&cli.force, "force", cli.force, "force certain actions")
	flag.BoolVar(&cli.all, "all", cli.force, "don't keep <plugin>.zip around")
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
	case "check":
		plugins := must_read_plugins(cli.file)
		act_check(plugins, cli.dir, ui)
	case "clean", "c", "cl":
		plugins := must_read_plugins(cli.file)
		act_clean(plugins, cli.dir, cli.force)
	case "prune":
		plugins := must_read_plugins(cli.file)
		act_prune(plugins, cli.dir, cli.force, cli.all)
	case "status", "st":
		plugins := may_read_plugins(cli.file)
		act_status(plugins, cli.dir)
	}
}

func may_read_plugins(path string) PluginList {
	plugins, err := ScanPluginFile(path)
	if err != nil {
		plugins = make(PluginList)
	}
	return plugins
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
