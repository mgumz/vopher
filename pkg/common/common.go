package common

const (
	Sha1ChecksumLen = 40

	// CommitFileName is written into the plugin folder and holds the
	// commit-id taken from the zip comment. The name stems from the time
	// only GitHub was supported; it is kept so installed plugins stay
	// checkable.
	CommitFileName = "github-commit"
)

// KnownHosts maps the host of a git hoster to the name of its hoster, as
// used in the plugin option `hoster=`
var KnownHosts = map[string]string{
	"github.com":   "github",
	"codeberg.org": "forgejo",
	"gitea.com":    "gitea",
	"gitlab.com":   "gitlab",
}
