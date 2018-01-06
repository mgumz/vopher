package main

import "fmt"

func actSample() {
	fmt.Println(`# Sample vopher.list file
# Comments start with '#', blank lines are ignored
#
# The following are some examples of fetching Tim Pope's "fugitive" plugin,
# designed for use with Git. An actual list of plugins will follow these
# examples that works out of the box, so consider this a quick guide to
# understanding vopher's list-file format. To put this output into a file,
# simply redirect it using your shell's '>' operator, e.g.:
# 'vopher sample > sample.list'.
#
# Fetch tpope's 'vim-fugitive' plugin, the master branch (rox, btw) and places
# the content of the zip-file into -dir <folder>/vim-fugitive.
# https://github.com/tpope/vim-fugitive
#
# Fetch tpope's 'vim-fugitive' plugin, but grab the tagged release 'v2.1'
# instead of 'master'.
# https://github.com/tpope/vim-fugitive#v2.1
#
# Fetch tpope's 'vim-fugitive' plugin and place it under -dir <folder>/foo
# foo https://github.com/tpope/vim-fugitive
#
# Fetch tpope's 'vim-fugitive' plugin, but do not strip any directories from the
# filenames in the zip. The default is to strip the first directory name, but
# sometimes you need to have more control.
# vim-fugitive https://github.com/tpope/vim-fugitive strip=0
#
# Fetch tpope's 'vim-fugitive' plugin directly via a link to the zip. It's wise
# to name the plugin, otherwise the plugin-folder would be 'v2.1' when installed.
# vim-fugitive https://github.com/tpope/vim-fugitive/archive/v2.1.zip
#
# Same as before, but check the zip against the sha1 given
# https://github.com/tpope/vim-fugitive/archive/v2.1.zip sha1=90437a3bd5f248bf5061f9afd3cc4a22fca4a11c
#
# SAMPLE PLUGIN LIST
# The following are a set of plugins that are rather popular and are a good
# place to start. Edit anything after this as you please. You can add entries
# from any URL found in "vopher search <term>"'s output or any URL that serves
# an archive containing the plugin.
commentary https://github.com/tpope/vim-commentary
exchange   https://github.com/tommcdo/vim-exchange
fugitive   https://github.com/tpope/vim-fugitive
gnupg      https://github.com/jamessan/vim-gnupg
surround   https://github.com/tpope/vim-surround
`)
}
