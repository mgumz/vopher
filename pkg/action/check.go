package action

import (
	"sync"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
)

func Check(plugins plugin.List, base string, ui ui.UI) {

	m := &sync.Map{}

	wg := sync.WaitGroup{}
	check := func(p *plugin.Plugin) {
		gh := acquire.Github{}
		text := gh.CheckPlugin(p, base)
		m.Store(p.Name, text)
		wg.Done()
	}

	for _, p := range plugins {
		switch p.URL.Host {
		case "github.com":
			wg.Add(1)
			go check(p)
		}
	}

	wg.Wait()

	for _, id := range plugins.SortByLineNumber() {
		if t, ok := m.Load(id); ok {
			text, _ := t.(string)
			ui.Print(id, text)
		}
	}
}
