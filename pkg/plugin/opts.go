package plugin

type Opts struct {
	StripDir   int    // strip n dir-parts from archive-entries
	PostUpdate string // execute after 'update'-action
	SHA1       string
}
