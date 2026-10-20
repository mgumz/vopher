package http

import (
	"errors"
	"strconv"
)

const defaultJobs = 8

// HostJobs maps a host-rule to the max. number of concurrent requests, see
// matchRule. 0 means unlimited.
type HostJobs map[string]int

// String implements flag.Value
func (hj *HostJobs) String() string {
	if hj == nil {
		return ""
	}
	return formatRules(*hj)
}

// Set implements flag.Value, accepts "host=n[,host=n...]"
func (hj *HostJobs) Set(v string) error {
	if *hj == nil {
		*hj = HostJobs{}
	}
	return parseRules(v, func(host, value string) error {
		jobs, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if jobs < 0 {
			return errors.New("negative number of jobs")
		}
		(*hj)[host] = jobs
		return nil
	})
}

// slots holds one semaphore per rule; a nil semaphore means unlimited
var slots = newSlots(HostJobs{})

func newSlots(hj HostJobs) map[string]chan struct{} {
	rules := withDefault(hj, defaultJobs)
	s := make(map[string]chan struct{}, len(rules))
	for host, n := range rules {
		var sem chan struct{}
		if n > 0 {
			sem = make(chan struct{}, n)
		}
		s[host] = sem
	}
	return s
}

// SetJobs configures the per-host limits. Call it before any request is made.
func SetJobs(hj HostJobs) { slots = newSlots(hj) }

// AcquireSlot blocks until a request to host may be made. The caller has to
// call the returned release func once the request is done.
func AcquireSlot(host string) (release func()) {
	sem := slots[matchRule(slots, host)]
	if sem == nil {
		return func() {}
	}
	sem <- struct{}{}
	return func() { <-sem }
}
