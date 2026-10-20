package http

import (
	"testing"
	"time"
)

func TestHostJobsSet(t *testing.T) {
	hj := HostJobs{}
	if err := hj.Set("*=4,GitHub.com=5"); err != nil {
		t.Fatal(err)
	}
	if err := hj.Set("codeberg.org=0"); err != nil {
		t.Fatal(err)
	}
	if got, want := hj.String(), "*=4,codeberg.org=0,github.com=5"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	for _, invalid := range []string{"github.com", "=5", "github.com=x", "github.com=-1"} {
		if err := hj.Set(invalid); err == nil {
			t.Errorf("expected error for %q", invalid)
		}
	}
}

func TestMatchRule(t *testing.T) {
	rules := HostJobs{AnyHost: 8, "github.com": 5, "gist.github.com": 1}
	tests := map[string]string{
		"github.com":                "github.com",
		"GitHub.com":                "github.com",
		"codeload.github.com":       "github.com",
		"gist.github.com":           "gist.github.com",
		"a.gist.github.com":         "gist.github.com",
		"notgithub.com":             AnyHost,
		"raw.githubusercontent.com": AnyHost,
		"codeberg.org":              AnyHost,
	}
	for host, want := range tests {
		if got := matchRule(rules, host); got != want {
			t.Errorf("%q: got %q, want %q", host, got, want)
		}
	}
}

func TestAcquireSlot(t *testing.T) {
	defer SetJobs(HostJobs{})
	SetJobs(HostJobs{"github.com": 1, AnyHost: 0})

	release := AcquireSlot("github.com")
	acquired := make(chan struct{})
	go func() {
		AcquireSlot("gist.github.com")()
		close(acquired)
	}()
	select {
	case <-acquired:
		t.Fatal("subdomain got a slot while the rule was exhausted")
	case <-time.After(50 * time.Millisecond):
	}
	AcquireSlot("codeberg.org")() // unlimited, must not block
	release()
	<-acquired
}
