package action

import (
	"github.com/mgumz/vopher/pkg/action/status"
	"github.com/mgumz/vopher/pkg/plugin"
)

func Status(plugins plugin.List, base string) {
	status.Execute(plugins, base)
}
