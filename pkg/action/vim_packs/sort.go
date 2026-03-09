package vim_packs

import "sort"

func (packs packNodes) sort() {

	sort.Slice(packs, func(i, j int) bool {
		// Dependencies come first
		if packs[i].IsDependency() != packs[j].IsDependency() {
			return packs[i].IsDependency()
		}
		// Same dependency status: sort by level (descending), then alphabetically
		if packs[i].NDependents != packs[j].NDependents {
			return packs[i].NDependents > packs[j].NDependents
		}
		return packs[i].Name < packs[j].Name
	})

}
