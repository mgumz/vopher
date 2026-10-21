package plugin

const (

	// The zip-files of the git hosters put the files into a sub folder
	// like this:
	//   vim-plugin/doc/plugin.txt
	//   vim-plugin/README.txt
	//
	DefaultStrip = 1
)

type Opts struct {
	StripDir   int    // strip n dir-parts from archive-entries
	PostUpdate string // execute after 'update'-action
	SHA1       string
	Branch     string // usually "master" or "main"
	Hoster     string // git hoster, eg. "github"; default by host of the URL

	DependsOn  []string
	MinVersion string
}

// There is no default "branch": without `branch=` and `#ref`, `update` asks
// the hoster for acquire.DefaultBranches.
var defaultOpts = Opts{StripDir: DefaultStrip}
