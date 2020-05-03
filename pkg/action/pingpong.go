package action

import (
	"time"

	"github.com/mgumz/vopher/pkg/ui"
)

// PingPong is test-ui
func PingPong(ui ui.UI) {
	ui.Start()

	for i := 0; i < 10; i++ {
		ui.AddJob("ping")
		<-time.After(500 * time.Millisecond)
		ui.JobDone("pong")
	}

	ui.Wait()
	ui.Stop()
}
