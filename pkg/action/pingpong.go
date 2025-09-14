package action

import (
	"time"

	"github.com/mgumz/vopher/pkg/ui"
)

const (
	uiPing   = 500 * time.Millisecond
	uiNPings = 10
)

// PingPong is test-ui
func PingPong(ui ui.UI) {
	ui.Start()

	for range uiNPings {
		ui.AddJob("ping")
		<-time.After(uiPing)
		ui.JobDone("pong")
	}

	ui.Wait()
	ui.Stop()
}
