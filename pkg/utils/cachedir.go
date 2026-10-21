package utils

import (
	"os"
	"path/filepath"
)

// CacheDir returns $XDG_CACHE_HOME/vopher, falling back to ~/.cache/vopher.
// A relative $XDG_CACHE_HOME is ignored, as the XDG spec demands. Returns ""
// if the home directory is unknown.
func CacheDir() string {
	dir := os.Getenv("XDG_CACHE_HOME")
	if !filepath.IsAbs(dir) {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".cache")
	}
	return filepath.Join(dir, "vopher")
}
