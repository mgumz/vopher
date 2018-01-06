package main

import "sync"

func actCheck(plugins PluginList, base string, ui JobUI) {

	wg := sync.WaitGroup{}

	for _, plugin := range plugins {
		switch plugin.url.Host {
		case "github.com":
			wg.Add(1)
			go func(p *Plugin) {
				gh := Github{}
				text := gh.checkPlugin(p, base)
				ui.Print(p.name, text)
				wg.Done()
			}(plugin)
		}
	}

	wg.Wait()
}
