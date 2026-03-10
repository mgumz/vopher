package action

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
)

// Check checks for updates in the upstream repository of each plugin in the
// plugins list and uses the ui to output the result
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
		if p.URL.Host == "github.com" {
			wg.Add(1)
			go check(p)
		}
	}

	idFmt := calcIdFmt(plugins)
	wg.Wait()

	sorted := plugins.SortByLineNumber()
	if len(sorted) == 0 {
		return
	}

	for _, id := range sorted {
		if t, ok := m.Load(id); ok {
			text, _ := t.(string)
			// right aligned Id
			id := fmt.Sprintf(idFmt, id)
			ui.Print(id, text)
		}
	}
	ui.Print("", "<!-- vim: foldmethod=marker foldmarker=✚,▲ foldenable -->\n")
}

func calcIdFmt(plugins plugin.List) string {

	w := 0
	for _, p := range plugins {
		if len(p.Name) > w {
			w = len(p.Name)
		}
	}
	return "%" + strconv.Itoa(w) + "s"

}
