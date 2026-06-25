package main

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podNamed(name string) v1.Pod {
	return v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func TestFilterPodsByName(t *testing.T) {
	pods := []v1.Pod{
		podNamed("api-gateway-abc123"),
		podNamed("api-gateway-def456"),
		podNamed("billing-worker-xyz789"),
	}

	t.Run("returns all pods whose name contains the substring", func(t *testing.T) {
		got := filterPodsByName(pods, "gateway")

		if len(got) != 2 {
			t.Fatalf("got %d pods, want 2", len(got))
		}
		if got[0].Name != "api-gateway-abc123" || got[1].Name != "api-gateway-def456" {
			t.Errorf("unexpected matches: %q, %q", got[0].Name, got[1].Name)
		}
	})

	t.Run("returns nil when nothing matches", func(t *testing.T) {
		if got := filterPodsByName(pods, "nope"); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})
}
