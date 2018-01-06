package main

// TODO:
//
// if vim7 is compiled with the +clientserver feature one can talk back to vim.
// this is usefull to create the actual ui directly in vim. the downside is:
// +clientserver relies on win32 / xclipboard support as it seems.
//
// idea: vopher opens a fifo and writes to it. vim reads from it.
//
// vim8 is different, it got async-comminication features.
