package main

// idea: instead of having python/ruby/curl/wget/fetch/git installed
// for a vim-plugin-manager to fetch the plugins i just want one binary
// which does it all.
//
// plugins: http://vimawesome.com/
//
// ui-options:
//
// * https://godoc.org/github.com/jroimartin/gocui
//
//  global-progress [..............]
//  plugin1         [....]
//  plugin2         [............]
//  plugin3         [..............]
//
// cons: vertical space
//
// ui-option2:
//   <-> global progress
//  [....|.....|.....|....|....|....]
//   ^
//   | plugin-progress via _-=#░█▓▒░█
//   v
//
// cons: horizontal space
//        plugin-name fehlt

import (
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var allowed_actions = []string{
	"u",
	"up",
	"update",
	"c",
	"clean",
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
	flag.Parse()

	if len(flag.Args()) > 0 {
		cli.action = flag.Args()[0]
	}

	if prefix_in_stringslice(allowed_actions, cli.action) == -1 {
		log.Fatal("error: unknown action")
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
		update(plugins, cli.dir, cli.force, ui)
	case "clean", "c", "cl":
		plugins := must_read_plugins(cli.file)
		clean(plugins, cli.dir, cli.force)
	}
}

func update(plugins PluginList, dir string, force bool, ui JobUi) {

	ui.Start()

	for _, plugin := range plugins {

		plugin_folder := filepath.Join(dir, plugin.name)

		_, err := os.Stat(plugin_folder)
		if err == nil { // plugin_folder exists
			if !force {
				continue
			}
		}

		if !strings.HasSuffix(plugin.url.Path, ".zip") {
			switch plugin.url.Host {
			case "github.com":
				remote_zip := first_not_empty(plugin.url.Fragment, "master") + ".zip"
				plugin.url.Path = path.Join(plugin.url.Path, "archive", remote_zip)
			default:
				ext, err := httpdetect_ftype(plugin.url.String())
				if err != nil {
					log.Printf("error: %q: %s", plugin.url, err)
					continue
				}
				if ext != ".zip" {
					log.Printf("error: %q: not a zip", plugin.url)
					continue
				}
			}
		}

		ui.AddJob(plugin_folder)
		go acquire(plugin_folder, plugin.url.String(), plugin.strip_dir, ui)
	}
	ui.Wait()
	ui.Stop()
}

func clean(plugins PluginList, dir string, force bool) {

	if !force {
		log.Println("'clean' needs -force flag")
		return
	}

	var prefix, suffix string

	for _, plugin := range plugins {
		plugin_folder := filepath.Join(dir, plugin.name)
		prefix = ""
		suffix = "ok"
		_, err := os.Stat(plugin_folder)
		if err == nil { // plugin_folder exists
			err = os.RemoveAll(plugin_folder)
			if err != nil {
				prefix = "error:"
				suffix = err.Error()
			}
		} else {
			prefix = "info:"
			suffix = "does not exist"
		}
		log.Println("'clean'", prefix, plugin_folder, suffix)
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
