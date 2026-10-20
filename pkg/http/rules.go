package http

import (
	"fmt"
	"maps"
	"sort"
	"strings"
)

// AnyHost is the rule which applies to all hosts not matched by another rule
const AnyHost = "*"

// parseRules splits "host=value[,host=value...]" and hands each pair to set
func parseRules(v string, set func(host, value string) error) error {
	for rule := range strings.SplitSeq(v, ",") {
		host, value, ok := strings.Cut(strings.TrimSpace(rule), "=")
		if !ok || host == "" {
			return fmt.Errorf("invalid rule %q, expected host=value", rule)
		}
		if err := set(strings.ToLower(host), value); err != nil {
			return fmt.Errorf("invalid value in %q: %w", rule, err)
		}
	}
	return nil
}

// formatRules is the counterpart of parseRules
func formatRules[V any](rules map[string]V) string {
	pairs := make([]string, 0, len(rules))
	for host, v := range rules {
		pairs = append(pairs, fmt.Sprintf("%s=%v", host, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

// matchRule returns the most specific rule in rules for host. A rule
// "github.com" matches "github.com" and all its subdomains.
func matchRule[V any](rules map[string]V, host string) string {
	host = strings.ToLower(host)
	for {
		if _, ok := rules[host]; ok {
			return host
		}
		idx := strings.IndexByte(host, '.')
		if idx == -1 {
			return AnyHost
		}
		host = host[idx+1:]
	}
}

// withDefault returns a copy of rules with AnyHost set to def, unless rules
// already contains AnyHost
func withDefault[V any](rules map[string]V, def V) map[string]V {
	merged := map[string]V{AnyHost: def}
	maps.Copy(merged, rules)
	return merged
}
