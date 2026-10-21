package action

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/mgumz/vopher/pkg/acquire"
	"github.com/mgumz/vopher/pkg/plugin"
	"github.com/mgumz/vopher/pkg/ui"
)

// Check checks for updates in the upstream repository of each plugin in the
// plugins list and uses the ui to output the result
func Check(plugins plugin.List, base, cacheDir string, ui ui.UI) {

	m := &sync.Map{}

	hosters := map[*plugin.Plugin]acquire.Hoster{}
	for _, p := range plugins {
		hoster, err := acquire.HosterFor(p)
		if err != nil {
			log.Printf("error: %q: %s", p.URL, err)
			continue
		}
		if hoster != nil {
			hosters[p] = hoster
		}
	}
	total := len(hosters)
	done := atomic.Int32{}

	wg := sync.WaitGroup{}
	check := func(p *plugin.Plugin, hoster acquire.Hoster) {
		text := acquire.CheckPlugin(hoster, p, base, cacheDir)
		m.Store(p.Name, text)
		fmt.Fprintf(os.Stderr, "check: %d/%d %s\n", done.Add(1), total, p.Name)
		wg.Done()
	}

	for p, hoster := range hosters {
		wg.Add(1)
		go check(p, hoster)
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
