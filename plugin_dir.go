package main

type PluginDirEntry struct {
	name     string
	exists   int
	isPlugin int
	isDir    int
}

type PluginDirEntryByName []*PluginDirEntry

func (a PluginDirEntryByName) Len() int           { return len(a) }
func (a PluginDirEntryByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PluginDirEntryByName) Less(i, j int) bool { return a[i].name < a[j].name }
