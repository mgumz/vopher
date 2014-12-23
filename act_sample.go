package main

import "fmt"

func act_sample() {
	fmt.Println(`# sample vopher.list file
# a comment starts with a '#', the whole line gets ignored.
# empty lines are ignored as well.

# fetch tpope's 'vim-fugitive' plugin, the master branch (rox, btw)
# and places the content of the zip-file into -dir <folder>/vim-fugitive.
https://github.com/tpope/vim-fugitive

# fetch tpope's 'vim-fugitive' plugin, but grab the tagged release 'v2.1'
# instead of 'master'.
https://github.com/tpope/vim-fugitive#v2.1

# acquire tpope's 'vim-fugitive' plugin and place it under -dir <folder>/foo
foo https://github.com/tpope/vim-fugitive

# acquire tpope's 'vim-fugitive' plugin directly via a link to the zip. it's
# wise to name the plugin, otherwise the plugin-folder would be 'v2.1'
vim-fugitive https://github.com/tpope/vim-fugitive/archive/v2.1.zip

# fetch tpope's 'vim-fugitive' plugin, but do not strip any directories
# from the filenames in the zip. the default is to strip the first directory
# name, but sometimes you need to have more control.
vim-fugitive https://github.com/tpope/vim-fugitive strip=0`)
}
