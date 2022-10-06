package utils

import (
	"net/url"
	"os"
	stduser "os/user"
	"path/filepath"
	"strings"
)

// FirstNotEmpty checks all given `parts` and returns the first one which is
// not empty.
func FirstNotEmpty(parts ...string) (result string) {
	for i := range parts {
		if len(parts[i]) > 0 {
			return parts[i]
		}
	}
	return
}

// IndexByteN // FIXME: fix docu, what does this function is intended to do?
func IndexByteN(path string, needle byte, n int) int {
	idx, s := 0, 0
	for ; s < n; s++ {
		i := strings.IndexByte(path[idx:], needle)
		if i < 0 { // not found
			break
		}

		idx = idx + i + 1
	}

	if s < n {
		idx = 0
	}

	return idx - 1
}

// PrefixInStringSlice checks if prefix `s` is in any of the given strings in
// `lst`
func PrefixInStringSlice(lst []string, s string) int {
	for i := range lst {
		if strings.HasPrefix(lst[i], s) {
			return i
		}
	}
	return -1
}

// StringHasSuffix checks if s is in any of the provided
// suffixes
func StringHasSuffix(s string, suffix []string) bool {
	for i := range suffix {
		if strings.HasSuffix(s, suffix[i]) {
			return true
		}
	}
	return false
}

// ExpandPath expands the path p if it starts with a ~ to the users home folder
func ExpandPath(p string) (string, error) {
	if p == "" {
		return p, nil
	}

	if p[0] == '~' {
		user, err := stduser.Current()
		if err != nil {
			return p, err
		}
		p = filepath.Join(user.HomeDir, p[1:])
	}

	return p, nil
}

// expands 'v' by replacing occurrences of  by their
// os.Environ() equivalent, except for $VOPHER_DIR which is
// replaced by 'vopher_dir'
//
// if no match is found, $VAR is returned.
//
// NOTE: this behavior is different from os.ExpandEnv()
func ExpandVarEnvironment(v, vopherDir string) string {
	switch v {
	case "VOPHER_DIR":
		return vopherDir
	default:
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, v) && (env[len(v)] == '=') && (len(env)-len(v) > 1) {
				return env[len(v)+1:]
			}
		}
	}

	// 404-environment -> "unaltered"
	return "$" + v
}

// wrapper around os.Expand() and expand_var_environment
func ExpandPathEnvironment(path, vopherDir string) string {
	return os.Expand(path, func(p string) string {
		return ExpandVarEnvironment(p, vopherDir)
	})
}

// parsePluginURL will parse a given URL into a *url.URL
// Since people are lazy and lazy people will just hand over
// "github.com/tpope/vim-fugitive" - we want to support lazy typing by
// prepending a protocol
func ParsePluginURL(u string) (*url.URL, error) {

	if strings.HasPrefix(u, "github.com/") {
		u = "https://" + u
	}

	return url.Parse(u)
}
