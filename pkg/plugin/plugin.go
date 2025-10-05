package plugin

import (
	"crypto/sha1" /* #nosec */
	"encoding/hex"
	"fmt"
	"math"
	neturl "net/url"
	"runtime"
	"strconv"
	"strings"

	"github.com/mgumz/vopher/pkg/vopher"
)

type Plugin struct {
	Name    string
	Ext     string
	URL     *neturl.URL
	Opts    Opts
	Archive vopher.Archive // Used to extract/view content of plugin

	ln int // Line in `vopher.list`, might be used for sorting
}

func (pl *Plugin) String() string {
	return fmt.Sprintf("Plugin{%q, %q, strip=%d}",
		pl.Name, pl.URL.String(), pl.Opts.StripDir)
}

func (p *Plugin) optionsFromFields(fields []string) error {

	postUpdateOS := "postupdate." + runtime.GOOS + "="

	for _, field := range fields {
		switch {
		case strings.HasPrefix(field, "strip="):
			strip, err := strconv.ParseUint(field[len("strip="):], 10, 8)
			if err == nil {
				if strip > 0 && strip < math.MaxInt32 {
					p.Opts.StripDir = int(strip)
					continue
				}
			}
			return fmt.Errorf("strange 'strip' field")
		case strings.HasPrefix(field, "postupdate=") && p.Opts.PostUpdate == "":
			p.Opts.PostUpdate = field[len("postupdate="):]
		case strings.HasPrefix(field, postUpdateOS):
			p.Opts.PostUpdate = field[len(postUpdateOS):]
		case strings.HasPrefix(field, "sha1="):
			p.Opts.SHA1 = strings.ToLower(field[len("sha1="):])
		case strings.HasPrefix(field, "branch="):
			p.Opts.Branch = field[len("branch="):]
		case strings.HasPrefix(field, "min-version="):
			p.Opts.MinVersion = field[len("min-version="):]
		case strings.HasPrefix(field, "depends-on="):
			p.Opts.DependsOn = parseDependsOn(field[len("depends-on="):])
		}
	}

	if p.Opts.PostUpdate != "" {
		decoded, err := neturl.QueryUnescape(p.Opts.PostUpdate)
		if err != nil {
			return err
		}
		p.Opts.PostUpdate = decoded
	}

	if p.Opts.SHA1 != "" && len(p.Opts.SHA1) != hex.EncodedLen(sha1.Size) {
		return fmt.Errorf("'sha1' field does not match size of a sha1")
	}

	return nil
}
