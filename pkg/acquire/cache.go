package acquire

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/mgumz/vopher/pkg/archive"
)

// ArchiveStem returns the path below cacheDir for the archive fetched from u:
// host and path of u, the query (if any) appended to the last element.
// Example:
//
//	https://github.com/tpope/vim-fugitive/archive/master.zip
//	=> <cacheDir>/github.com/tpope/vim-fugitive/archive/master.zip
func ArchiveStem(cacheDir string, u *url.URL) string {
	host := u.Hostname()
	if port := u.Port(); port != "" {
		host += "_" + port
	}
	p := path.Clean("/" + u.Path)
	if u.RawQuery != "" {
		p += "_" + url.QueryEscape(u.RawQuery)
	}
	return filepath.Join(cacheDir, host, filepath.FromSlash(p))
}

// ArchivePath returns ArchiveStem with ext, see WithExt
func ArchivePath(cacheDir string, u *url.URL, ext string) string {
	return WithExt(ArchiveStem(cacheDir, u), ext)
}

// WithExt returns stem with ext appended, unless stem already ends with ext
func WithExt(stem, ext string) string {
	if !strings.HasSuffix(stem, ext) {
		stem += ext
	}
	return stem
}

// FindArchive returns the kept archive with the given stem, whatever its
// archive suffix is
func FindArchive(stem string) (string, bool) {
	candidates, _ := filepath.Glob(stem + "*")
	for _, name := range candidates {
		if IsArchiveOf(name, stem) {
			return name, true
		}
	}
	return "", false
}

// IsArchiveOf returns true if name is the archive with the given stem, see
// ArchiveStem and ArchivePath
func IsArchiveOf(name, stem string) bool {
	if name == stem {
		return true
	}
	ok, suffixLen := archive.IsSupportedArchive(name)
	return ok && strings.HasPrefix(name, stem) && len(name)-suffixLen == len(stem)
}
