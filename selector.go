package main

import (
	"fmt"
	"sort"
	"strings"

	v1 "k8s.io/api/core/v1"
)

// preferredLabelKeys are the label keys, in priority order, that pod-id uses to
// derive a `kubectl logs -l <selector>` value from a set of matched pods. The
// Kubernetes recommended label (app.kubernetes.io/name) is preferred, followed
// by the common conventional fallbacks.
var preferredLabelKeys = []string{
	"app.kubernetes.io/name",
	"app",
	"k8s-app",
	"app.kubernetes.io/instance",
}

// resolveSelector derives a label selector covering every pod in the given set,
// so the matched pods can be read together with
// `kubectl logs -l <selector> --prefix`.
//
// It tries each preferred label key in priority order and uses the first key
// that is present on all pods. When every pod shares the same value it returns
// an equality selector ("key=value"). When the matched pods span more than one
// app it returns a set-based selector ("key in (value1,value2)") so all the
// matched apps are still covered rather than the query being rejected.
func resolveSelector(pods []v1.Pod) (string, error) {
	if len(pods) == 0 {
		return "", fmt.Errorf("no pods to derive a selector from")
	}

	for _, key := range preferredLabelKeys {
		values, present := labelValues(pods, key)
		if !present {
			continue
		}

		return formatSelector(key, values), nil
	}

	return "", fmt.Errorf("could not derive a selector: none of the known label keys %v are present on all matched pods", preferredLabelKeys)
}

// labelValues collects the distinct values of the given label key across all
// pods. It reports the sorted, deduplicated values and whether the key is
// present on every pod. If any pod lacks the key, present is false so the caller
// falls back to the next preferred key.
func labelValues(pods []v1.Pod, key string) (values []string, present bool) {
	seen := make(map[string]struct{})

	for _, pod := range pods {
		v, ok := pod.Labels[key]
		if !ok {
			return nil, false
		}
		seen[v] = struct{}{}
	}

	values = make([]string, 0, len(seen))
	for v := range seen {
		values = append(values, v)
	}
	sort.Strings(values)

	return values, true
}

// formatSelector renders a selector for one label key. A single shared value
// produces an equality selector ("key=value"); multiple values produce a
// set-based selector ("key in (value1,value2)").
func formatSelector(key string, values []string) string {
	if len(values) == 1 {
		return fmt.Sprintf("%s=%s", key, values[0])
	}
	return fmt.Sprintf("%s in (%s)", key, strings.Join(values, ","))
}
