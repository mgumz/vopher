package vim_packs

import (
	"fmt"
	"path"

	"github.com/mgumz/vopher/pkg/plugin"
)

type packNode struct {
	Name        string
	NDependents int
}

func (pn *packNode) IsDependency() bool {
	return pn.NDependents > 0
}

type packNodes []packNode

// https://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm
//
// related:
// * https://pkg.go.dev/github.com/looplab/tarjan - its a single file in the
//   end, no need to introduce a dependency; was inspired by
//   mgo/txn/tarjan.go, see next line
// * https://github.com/juju/mgo/blob/master/txn/tarjan.go - the
//   implementation of Gustavo Niemeyer

type tarjanState struct {
	index      int
	lowLink    int
	onStack    bool
	discovered bool
}

type tarjanSolver struct {
	graph      map[string][]string // adjacency list: pack -> dep
	state      map[string]*tarjanState
	stack      []string
	index      int
	result     packNodes
	sccs       [][]string // strongly connected components
	currentSCC []string
	dependents map[string]int
}

func newTarjanSolver(packs plugin.List) *tarjanSolver {
	graph := make(map[string][]string)
	state := make(map[string]*tarjanState)
	dependents := make(map[string]int)

	for _, pack := range packs {
		pn := path.Base(pack.Name)
		if _, exists := graph[pn]; exists {
			graph[pn] = []string{}
		}
		state[pn] = &tarjanState{
			index:      -1,
			lowLink:    -1,
			onStack:    false,
			discovered: false,
		}

		for _, dep := range pack.Opts.DependsOn {
			dep = path.Base(dep)
			graph[pn] = append(graph[pn], dep)
			if _, exists := dependents[dep]; !exists {
				dependents[dep] = 1
			} else {
				dependents[dep] += 1
			}
			if _, exists := state[dep]; !exists {
				state[dep] = &tarjanState{
					index:      -1,
					lowLink:    -1,
					onStack:    false,
					discovered: false,
				}
			}
		}
	}

	return &tarjanSolver{
		graph:      graph,
		state:      state,
		stack:      []string{},
		index:      0,
		result:     packNodes{},
		sccs:       [][]string{},
		dependents: dependents,
	}
}

func (ts *tarjanSolver) strongconnect(v string) {
	s := ts.state[v]
	s.index = ts.index
	s.lowLink = ts.index
	ts.index++
	ts.stack = append(ts.stack, v)
	s.onStack = true

	for _, dep := range ts.graph[v] {
		depState := ts.state[dep]
		if depState.index == -1 {
			// Successor not yet visited
			ts.strongconnect(dep)
			s.lowLink = min(s.lowLink, depState.lowLink)
		} else if depState.onStack {
			// Successor is on stack - back edge, cycle detected
			s.lowLink = min(s.lowLink, depState.index)
		}
	}

	if s.lowLink == s.index {
		ts.currentSCC = []string{}
		for {
			l := len(ts.stack) - 1
			w := ts.stack[l]
			ts.stack = ts.stack[:l]
			ts.state[w].onStack = false
			ts.currentSCC = append(ts.currentSCC, w)
			ts.result = append(ts.result, packNode{
				Name:        w,
				NDependents: ts.dependents[w],
			})

			if w == v {
				break
			}
		}
		ts.sccs = append(ts.sccs, ts.currentSCC)
	}
}

func (ts *tarjanSolver) solve() (packNodes, [][]string, error) {
	// Run DFS on all undiscovered nodes
	for node := range ts.state {
		if ts.state[node].index == -1 {
			ts.strongconnect(node)
		}
	}

	// Check for cycles (SCCs with more than one node)
	var cycles [][]string
	for _, scc := range ts.sccs {
		if len(scc) > 1 {
			cycles = append(cycles, scc)
		}
	}

	if len(cycles) > 0 {
		return nil, cycles, fmt.Errorf("circular dependencies detected")
	}

	return ts.result, nil, nil
}
