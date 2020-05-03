package action

import (
	"sync"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
)

func Check(plugins plugin.List, base string, ui ui.UI) {

	wg := sync.WaitGroup{}
	check := func(p *plugin.Plugin) {
		gh := acquire.Github{}
		text := gh.CheckPlugin(p, base)
		ui.Print(p.Name, text)
		wg.Done()
	}

	for _, p := range plugins {
		switch p.URL.Host {
		case "github.com":
			wg.Add(1)
			check(p)
		}
	}

	wg.Wait()
}
