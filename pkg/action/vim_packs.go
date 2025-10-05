package action

import (
	"github.com/mgumz/vopher/pkg/action/vim_packs"
	"github.com/mgumz/vopher/pkg/plugin"
)

func VimPacks(plugins plugin.List) {
	vim_packs.Execute(plugins)
}
