package http

import (
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHostTimeoutsSet(t *testing.T) {
	ht := HostTimeouts{}
	if err := ht.Set("*=1m,github.com=20s"); err != nil {
		t.Fatal(err)
	}
	if got, want := ht.String(), "*=1m0s,github.com=20s"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	for _, invalid := range []string{"github.com=20", "github.com=-1s"} {
		if err := ht.Set(invalid); err == nil {
			t.Errorf("expected error for %q", invalid)
		}
	}
}

func TestRequestTimeout(t *testing.T) {
	const timeout = 50 * time.Millisecond
	defer SetTimeouts(HostTimeouts{})
	SetTimeouts(HostTimeouts{AnyHost: timeout})

	stalled := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		<-r.Context().Done()
	}))
	defer stalled.Close()

	if _, _, err := Request(nethttp.MethodGet, stalled.URL); err == nil || !strings.Contains(err.Error(), "no progress") {
		t.Fatalf("expected a timeout, got %v", err)
	}

	// takes 4x the timeout in total, but makes progress all the time
	slow := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(nethttp.StatusOK)
		for range 8 {
			_, _ = w.Write([]byte("."))
			w.(nethttp.Flusher).Flush()
			time.Sleep(timeout / 2)
		}
	}))
	defer slow.Close()

	resp, done, err := Request(nethttp.MethodGet, slow.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "........" {
		t.Fatalf("got %q", body)
	}
}
