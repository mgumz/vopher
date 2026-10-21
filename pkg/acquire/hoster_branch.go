package acquire

import (
	"io"
	"net/http"
	"slices"

	vhttp "github.com/mgumz/vopher/pkg/http"
)

// maxBranchesBytes limits the answers read by fetchBody
const maxBranchesBytes = 1024 * 1024

// branchCandidates returns the branches BranchOf prefers, in this order:
// `branch=`, the default branch the hoster names, DefaultBranches. Empty
// and repeated names are dropped.
func branchCandidates(branch, defaultBranch string) []string {
	candidates := []string{}
	for _, b := range append([]string{branch, defaultBranch}, DefaultBranches...) {
		if b != "" && !slices.Contains(candidates, b) {
			candidates = append(candidates, b)
		}
	}
	return candidates
}

// pickBranch returns the first of candidates in branches, else the first
// of branches
func pickBranch(branches, candidates []string) string {
	for _, c := range candidates {
		if slices.Contains(branches, c) {
			return c
		}
	}
	if len(branches) > 0 {
		return branches[0]
	}
	return ""
}

// fetchBody fetches u and returns the status and at most maxBranchesBytes
// of the body
func fetchBody(u string) (body []byte, status int, err error) {
	resp, done, err := vhttp.Request(http.MethodGet, u)
	if err != nil {
		return nil, 0, err
	}
	defer done()
	body, err = io.ReadAll(io.LimitReader(resp.Body, maxBranchesBytes))
	return body, resp.StatusCode, err
}
