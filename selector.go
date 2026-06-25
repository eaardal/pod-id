package main

import (
	"fmt"

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

// resolveSelector derives a single label selector ("key=value") shared by every
// pod in the given set, so the matched pods can be read together with
// `kubectl logs -l <selector> --prefix`.
//
// It tries each preferred label key in priority order and uses the first key
// that is present on all pods. If that key's value differs between pods, the
// matched pods span more than one app and an error is returned so the caller can
// narrow the query rather than silently dropping pods.
func resolveSelector(pods []v1.Pod) (string, error) {
	if len(pods) == 0 {
		return "", fmt.Errorf("no pods to derive a selector from")
	}

	for _, key := range preferredLabelKeys {
		value, present, consistent := labelValue(pods, key)

		if !present {
			continue
		}

		if !consistent {
			return "", fmt.Errorf("matched pods span multiple apps for label %q; narrow your query to a single app", key)
		}

		return fmt.Sprintf("%s=%s", key, value), nil
	}

	return "", fmt.Errorf("could not derive a selector: none of the known label keys %v are present on all matched pods", preferredLabelKeys)
}

// labelValue inspects the given label key across all pods. It reports the shared
// value, whether the key is present on every pod, and whether every pod agrees
// on the value.
func labelValue(pods []v1.Pod, key string) (value string, present bool, consistent bool) {
	first := true

	for _, pod := range pods {
		v, ok := pod.Labels[key]
		if !ok {
			return "", false, false
		}

		if first {
			value = v
			first = false
			continue
		}

		if v != value {
			return value, true, false
		}
	}

	return value, true, true
}
