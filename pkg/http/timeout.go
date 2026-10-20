package http

import (
	"errors"
	"time"
)

const defaultTimeout = 30 * time.Second

// HostTimeouts maps a host-rule to the time a request may go without
// progress before it is aborted, see matchRule. 0 means no timeout.
type HostTimeouts map[string]time.Duration

// String implements flag.Value
func (ht *HostTimeouts) String() string {
	if ht == nil {
		return ""
	}
	return formatRules(*ht)
}

// Set implements flag.Value, accepts "host=duration[,host=duration...]"
func (ht *HostTimeouts) Set(v string) error {
	if *ht == nil {
		*ht = HostTimeouts{}
	}
	return parseRules(v, func(host, value string) error {
		d, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		if d < 0 {
			return errors.New("negative timeout")
		}
		(*ht)[host] = d
		return nil
	})
}

var timeouts = withDefault(HostTimeouts{}, defaultTimeout)

// SetTimeouts configures the per-host timeouts. Call it before any request
// is made.
func SetTimeouts(ht HostTimeouts) { timeouts = withDefault(ht, defaultTimeout) }

func timeoutFor(host string) time.Duration {
	return timeouts[matchRule(timeouts, host)]
}
