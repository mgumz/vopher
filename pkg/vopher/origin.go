package vopher

// Origin tells where the archive of a plugin came from
type Origin int

const (
	OriginDownloaded  Origin = iota // fetched from the server
	OriginNotModified               // kept archive, confirmed by the server (304)
	OriginKept                      // kept archive, server not asked (commit, sha1, offline)
)

func (o Origin) String() string {
	switch o {
	case OriginDownloaded:
		return "downloaded"
	case OriginNotModified:
		return "unchanged (cache)"
	case OriginKept:
		return "kept (cache)"
	}
	return "unknown"
}
