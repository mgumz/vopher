// *vopher* - acquire and manage vim-plugins by the power of the gopher
//
// idea: instead of having python/ruby/curl/wget/fetch/git installed
// for a vim-plugin-manager to fetch the plugins i just want one binary
// which does it all.
//
// plugins: http://vimawesome.com/
package main

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
	"clean",
	"c",
	"check",
	"prune",
	"sample",
	"st",
	"status",
	"search",
}

func usage() {

	fmt.Fprintln(os.Stderr, `vopher - acquire vim-plugins the gopher-way

usage: vopher [flags] <action>

actions
  update - acquire the given plugins from the -f <list>
  search - use http://vimawesome.com/ to list some plugins
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
		dry    bool
		all    bool
		file   string
		dir    string
		ui     string
	}{action: "update", dir: ".", ui: "oneline"}

	flag.BoolVar(&cli.force, "force", cli.force, "force certain actions")
	flag.BoolVar(&cli.dry, "dry", cli.dry, "dry-run, show what would happen")
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
	} else if cli.action == "search" && len(flag.Args()) < 2 {
		log.Fatal("error: missing arguments for 'search'")
	}

	var ui JobUi
	switch cli.ui {
	case "oneline":
		ui = &UiOneLine{
			ProgressTicker: NewProgressTicker(0),
			prefix:         "vopher",
			duration:       25 * time.Millisecond,
		}
	case "simple":
		ui = &UiSimple{jobs: make(map[string]_ri)}
	}

	if cli.dir == "" {
		log.Fatal("error: empty -dir")
	}

	var (
		path string
		err  error
	)
	if path, err = expand_path(cli.dir); err != nil {
		log.Fatal("error: expanding %q failed while obtaining current user?? %s", cli.dir, err)
	}
	cli.dir = path

	switch cli.action {
	case "update", "u", "up":
		plugins := must_read_plugins(cli.file)
		opts := actUpdateOptions{dir: cli.dir, force: cli.force, dry_run: cli.dry}
		act_update(plugins, ui, &opts)
	case "check", "c":
		plugins := must_read_plugins(cli.file)
		act_check(plugins, cli.dir, ui)
	case "clean":
		plugins := must_read_plugins(cli.file)
		act_clean(plugins, cli.dir, cli.force)
	case "prune":
		plugins := must_read_plugins(cli.file)
		act_prune(plugins, cli.dir, cli.force, cli.all)
	case "status", "st":
		plugins := may_read_plugins(cli.file)
		act_status(plugins, cli.dir)
	case "search":
		act_search(flag.Args()[1:]...)
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
