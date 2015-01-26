package main

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
	"testing"
)

func Test_ScanPluginFile(t *testing.T) {

	const N_PLUGINS = 7
	sample := `
# ignore this
http://example.com/plugin1
http://example.com/plugin2.zip
http://example.com/plugin3#tag-a
plugin4 http://example.com/pluginx#tag-a

http://example.com/plugin5 opt1=0 opt2=0 strip=2
plugin6 http://example.com/pluginx opt1=0 opt2=0 strip=2
http://example.com/plugin7 postupdate=a`

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

func Test_ScanPluginOptions(t *testing.T) {

	samples := []struct {
		fields   []string
		expected Plugin
	}{
		{[]string{"a", "b", "c"}, Plugin{strip_dir: 0}},
		{[]string{"strip=1", "b", "c"}, Plugin{strip_dir: 1}},
		{[]string{"a", "strip=1", "c"}, Plugin{strip_dir: 1}},
		{[]string{"postupdate=foo", "strip=1", "c"}, Plugin{strip_dir: 1, postupdate: "foo"}},
		{[]string{"postupdate=foo", "strip=1", "postupdate." + runtime.GOOS + "=bar"}, Plugin{strip_dir: 1, postupdate: "bar"}},
		{[]string{"postupdate." + runtime.GOOS + "=bar", "strip=1", "postupdate=foo"}, Plugin{strip_dir: 1, postupdate: "bar"}},
		{[]string{"postupdate=%22foo%20bar%20blub%22"}, Plugin{postupdate: "foo bar blub"}},
		{[]string{"postupdate=foo/bar/blub"}, Plugin{postupdate: "foo/bar/blub"}},
		{[]string{"postupdate=foo+bar/blub"}, Plugin{postupdate: "foo bar/blub"}},
	}

	for i := range samples {
		plugin := Plugin{}
		if err := plugin.OptionsFromFields(samples[i].fields); err != nil {
			t.Fatalf("%d: %v, expected %v, got %v", i, samples[i].expected, plugin)
		}
		t.Logf("%d: %v => %v", i, samples[i].fields, plugin)
	}
}
