package plugin

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"
)

func Test_ScanPluginFile(t *testing.T) {

	const nPlugins = 8
	sample := `
# ignore this
http://example.com/plugin1
http://example.com/plugin2.zip
http://example.com/plugin3#tag-a
plugin4 http://example.com/pluginx#tag-a

http://example.com/plugin5 opt1=0 opt2=0 strip=2
plugin6 http://example.com/pluginx opt1=0 opt2=0 strip=2
http://example.com/plugin7 postupdate=a 
  plugin8 http://example.com/leading-ws  
`

	scanned := make(List)
	sr := strings.NewReader(sample)

	if err := scanned.Parse(io.NopCloser(sr)); err != nil {
		t.Fatal(err)
	}

	if len(scanned) != nPlugins {
		t.Fatal("expected", nPlugins, "plugins, got", len(scanned))
	}

	for i := range nPlugins {
		name := fmt.Sprintf("plugin%d", i+1)
		plugin, ok := scanned[name]
		if !ok {
			t.Fatalf("expected plugin %q, not found", name)
		}
		t.Logf("%s %s", plugin.Name, plugin.URL)
	}
}

func Test_ScanPluginOptions(t *testing.T) {

	samples := []struct {
		fields   []string
		expected Plugin
	}{
		{[]string{"a", "b", "c"},
			Plugin{Opts: Opts{StripDir: 0}}},
		{[]string{"strip=1", "b", "c"},
			Plugin{Opts: Opts{StripDir: 1}}},
		{[]string{"a", "strip=1", "c"},
			Plugin{Opts: Opts{StripDir: 1}}},
		{[]string{"postupdate=foo", "strip=1", "c"},
			Plugin{Opts: Opts{StripDir: 1, PostUpdate: "foo"}}},
		{[]string{"postupdate=foo", "strip=1", "postupdate." + runtime.GOOS + "=bar"},
			Plugin{Opts: Opts{StripDir: 1, PostUpdate: "bar"}}},
		{[]string{"postupdate." + runtime.GOOS + "=bar", "strip=1", "postupdate=foo"},
			Plugin{Opts: Opts{StripDir: 1, PostUpdate: "bar"}}},
		{[]string{"postupdate=%22foo%20bar%20blub%22"},
			Plugin{Opts: Opts{PostUpdate: "foo bar blub"}}},
		{[]string{"postupdate=foo/bar/blub"},
			Plugin{Opts: Opts{PostUpdate: "foo/bar/blub"}}},
		{[]string{"postupdate=foo+bar/blub"},
			Plugin{Opts: Opts{PostUpdate: "foo bar/blub"}}},
	}

	for i := range samples {
		plugin := Plugin{}
		if err := plugin.optionsFromFields(samples[i].fields); err != nil {
			t.Fatalf("%d: %v, expected %v, got %v", i, err, samples[i].expected, plugin)
		}
		t.Logf("%d: %v => %v", i, samples[i].fields, plugin)
	}
}
