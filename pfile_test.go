package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func Test_ScanPluginFile(t *testing.T) {

	const N_PLUGINS = 6
	sample := `
# ignore this
http://example.com/plugin1
http://example.com/plugin2.zip
http://example.com/plugin3#tag-a
plugin4 http://example.com/pluginx#tag-a

http://example.com/plugin5 opt1=0 opt2=0 strip=2
plugin6 http://example.com/pluginx opt1=0 opt2=0 strip=2`

	sr := strings.NewReader(sample)
	scanned, err := ScanPluginReader(ioutil.NopCloser(sr))
	if err != nil {
		t.Fatal(err)
	}

	if len(scanned) != N_PLUGINS {
		t.Fatal("expected", N_PLUGINS, "plugins, got", len(scanned))
	}

	var plugin Plugin
	var ok bool
	for i := 0; i < N_PLUGINS; i++ {
		name := fmt.Sprintf("plugin%d", i+1)
		if plugin, ok = scanned[name]; !ok {
			t.Fatalf("expected plugin %q, not found", name)
		}
		t.Logf("%s", &plugin)
	}
}
