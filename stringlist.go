package main

import "strings"

type stringList []string

func (sl *stringList) String() string     { return strings.Join(*sl, ", ") }
func (sl *stringList) Set(v string) error { *sl = append(*sl, v); return nil }
