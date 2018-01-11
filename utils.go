package main

import (
	"os"
	stduser "os/user"
	"path/filepath"
	"strings"
)

func firstNotEmpty(parts ...string) (result string) {
	for i := range parts {
		if len(parts[i]) > 0 {
			return parts[i]
		}
	}
	return
}

func indexByteN(path string, needle byte, n int) int {
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

func prefixInStringSlice(lst []string, s string) int {
	for i := range lst {
		if strings.HasPrefix(lst[i], s) {
			return i
		}
	}
	return -1
}

func expandPath(p string) (string, error) {
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
func expandVarEnvironment(v, vopherDir string) string {
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
func expandPathEnvironment(path, vopherDir string) string {
	return os.Expand(path, func(p string) string {
		return expandVarEnvironment(p, vopherDir)
	})
}
