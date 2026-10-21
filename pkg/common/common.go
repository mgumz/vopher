package common

const (
	Sha1ChecksumLen = 40

	// CommitFileName is written into the plugin folder and holds the
	// commit-id taken from the zip comment. The name stems from the time
	// only GitHub was supported; it is kept so installed plugins stay
	// checkable.
	CommitFileName = "github-commit"
)
