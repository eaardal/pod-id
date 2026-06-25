package main

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podNamed(name string) v1.Pod {
	return v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func TestSplitAppNames(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single name", "api-gateway", []string{"api-gateway"}},
		{"two names", "api-gateway,invoice", []string{"api-gateway", "invoice"}},
		{"trims surrounding whitespace", "api-gateway, invoice", []string{"api-gateway", "invoice"}},
		{"drops empty parts", "api-gateway,,invoice,", []string{"api-gateway", "invoice"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitAppNames(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitAppNames(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterPodsByNames(t *testing.T) {
	pods := []v1.Pod{
		podNamed("orders-api-gateway-abc123"),
		podNamed("web-gateway-def456"),
		podNamed("payments-invoice-api-xyz789"),
	}

	t.Run("single partial name matches every pod containing it", func(t *testing.T) {
		got := filterPodsByNames(pods, []string{"gateway"})
		assertPodNames(t, got, "orders-api-gateway-abc123", "web-gateway-def456")
	})

	t.Run("multiple partial names match the union", func(t *testing.T) {
		got := filterPodsByNames(pods, []string{"api-gateway", "invoice"})
		assertPodNames(t, got, "orders-api-gateway-abc123", "payments-invoice-api-xyz789")
	})

	t.Run("a pod matched by several names appears once, in listing order", func(t *testing.T) {
		got := filterPodsByNames(pods, []string{"api", "gateway"})
		assertPodNames(t, got, "orders-api-gateway-abc123", "web-gateway-def456", "payments-invoice-api-xyz789")
	})

	t.Run("returns nil when nothing matches", func(t *testing.T) {
		if got := filterPodsByNames(pods, []string{"nope"}); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})
}

func assertPodNames(t *testing.T, got []v1.Pod, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d pods %v, want %d %v", len(got), podNames(got), len(want), want)
	}
	for i, w := range want {
		if got[i].Name != w {
			t.Errorf("pod[%d] = %q, want %q", i, got[i].Name, w)
		}
	}
}

func podNames(pods []v1.Pod) []string {
	names := make([]string, len(pods))
	for i, p := range pods {
		names[i] = p.Name
	}
	return names
}
