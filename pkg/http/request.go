package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	neturl "net/url"
	"time"
)

// Request issues method to url. It waits for a free slot for the host of
// url, see AcquireSlot, and aborts the request once it made no progress for
// the timeout configured for that host, see SetTimeouts. The caller has to
// call done once the response body is consumed; done also closes the body.
func Request(method, url string) (resp *nethttp.Response, done func(), err error) {

	u, err := neturl.Parse(url)
	if err != nil {
		return nil, nil, err
	}
	host := u.Hostname()

	release := AcquireSlot(host)
	ctx, cancel := context.WithCancelCause(context.Background())
	timeout := timeoutFor(host)
	var timer *time.Timer
	if timeout > 0 {
		timer = time.AfterFunc(timeout, func() {
			cancel(fmt.Errorf("no progress for %s", timeout))
		})
	}
	finish := func() {
		if timer != nil {
			timer.Stop()
		}
		cancel(nil)
		release()
	}

	req, err := nethttp.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		finish()
		return nil, nil, err
	}
	resp, err = nethttp.DefaultClient.Do(req) // #nosec G107
	if err != nil {
		finish()
		return nil, nil, causeOf(ctx, err)
	}

	resp.Body = &progressReader{ctx: ctx, body: resp.Body, timer: timer, timeout: timeout}
	done = func() {
		_ = resp.Body.Close()
		finish()
	}
	return resp, done, nil
}

// progressReader restarts the timeout of a request on every read which
// returned data
type progressReader struct {
	ctx     context.Context
	body    io.ReadCloser
	timer   *time.Timer
	timeout time.Duration
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.body.Read(p)
	if n > 0 && pr.timer != nil {
		pr.timer.Reset(pr.timeout)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		err = causeOf(pr.ctx, err)
	}
	return n, err
}

func (pr *progressReader) Close() error { return pr.body.Close() }

// causeOf replaces the generic "context canceled" by the reason of the
// cancellation
func causeOf(ctx context.Context, err error) error {
	if cause := context.Cause(ctx); cause != nil && !errors.Is(cause, context.Canceled) {
		return cause
	}
	return err
}
