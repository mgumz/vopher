package plugin

const (

	// Most plugins are fetched from github. The github zip-files
	// put the files into a sub folder like this:
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

	DependsOn  []string
	MinVersion string
}

// NOTE: somewhen, the default "branch" should become "main" to reflect
// current situation in most repos. atm, "" is ok to behave like before
// once we set a default branch, users need to put `branch="master"` into
// their vopher files to point to the correct branch. not yet.
var defaultOpts = Opts{StripDir: DefaultStrip}
