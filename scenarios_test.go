package main

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

// TestSelectorScenarios encodes the end-to-end behaviour for the input shapes
// pod-id supports in selector mode: a single partial name matching one app, a
// single partial name matching multiple apps, and comma-separated partial names
// that each match one or several apps. The pipeline mirrors main(): split the
// raw argument into partial names, match pods by any of them, then derive a
// selector covering all the matched pods.
func TestSelectorScenarios(t *testing.T) {
	pods := []v1.Pod{
		podWithLabels("orders-api-gateway-abc", map[string]string{"app": "orders-api-gateway"}),
		podWithLabels("web-gateway-def", map[string]string{"app": "web-gateway"}),
		podWithLabels("payments-invoice-api-ghi", map[string]string{"app": "payments-invoice-api"}),
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "a) single partial name matching one app",
			input: "api-gateway",
			want:  "app=orders-api-gateway",
		},
		{
			name:  "b) single partial name matching multiple apps",
			input: "gateway",
			want:  "app in (orders-api-gateway,web-gateway)",
		},
		{
			name:  "c) comma-separated names each matching one app",
			input: "api-gateway,invoice",
			want:  "app in (orders-api-gateway,payments-invoice-api)",
		},
		{
			name:  "d) comma-separated names matching multiple apps in total",
			input: "gateway,invoice",
			want:  "app in (orders-api-gateway,payments-invoice-api,web-gateway)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matching := filterPodsByNames(pods, splitAppNames(tt.input))

			got, err := resolveSelector(matching)
			if err != nil {
				t.Fatalf("resolveSelector() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("input %q => %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
