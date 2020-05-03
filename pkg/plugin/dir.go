package plugin

type DirEntry struct {
	Name     string
	Exists   int
	IsPlugin int
	IsDir    int
}

type DirEntryByName []*DirEntry

func (a DirEntryByName) Len() int           { return len(a) }
func (a DirEntryByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DirEntryByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
