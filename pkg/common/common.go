package common

const (
	Sha1ChecksumLen = 40

	// CommitFileName is written into the plugin folder and holds the
	// commit-id taken from the zip comment
	CommitFileName = "git-commit"

	// LegacyCommitFileName is the name of CommitFileName from the time only
	// GitHub was supported; read so plugins installed back then stay
	// checkable
	LegacyCommitFileName = "github-commit"
)

// KnownHosts maps the host of a git hoster to the name of its hoster, as
// used in the plugin option `hoster=`
var KnownHosts = map[string]string{
	"github.com":   "github",
	"codeberg.org": "forgejo",
	"gitea.com":    "gitea",
	"gitlab.com":   "gitlab",
}
