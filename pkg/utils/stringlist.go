package utils

import "strings"

type StringList []string

func (sl *StringList) String() string     { return strings.Join(*sl, ", ") }
func (sl *StringList) Set(v string) error { *sl = append(*sl, v); return nil }
