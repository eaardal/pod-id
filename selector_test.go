package main

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podWithLabels(name string, labels map[string]string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func TestResolveSelector(t *testing.T) {
	tests := []struct {
		name    string
		pods    []v1.Pod
		want    string
		wantErr bool
	}{
		{
			name: "single app: all pods share the same app label",
			pods: []v1.Pod{
				podWithLabels("my-service-abc", map[string]string{"app": "my-service"}),
				podWithLabels("my-service-def", map[string]string{"app": "my-service"}),
			},
			want: "app=my-service",
		},
		{
			name: "prefers app.kubernetes.io/name over app",
			pods: []v1.Pod{
				podWithLabels("my-service-abc", map[string]string{
					"app":                    "legacy",
					"app.kubernetes.io/name": "my-service",
				}),
				podWithLabels("my-service-def", map[string]string{
					"app":                    "legacy",
					"app.kubernetes.io/name": "my-service",
				}),
			},
			want: "app.kubernetes.io/name=my-service",
		},
		{
			name: "falls back to a lower-priority key when the preferred one is absent",
			pods: []v1.Pod{
				podWithLabels("my-service-abc", map[string]string{"k8s-app": "my-service"}),
				podWithLabels("my-service-def", map[string]string{"k8s-app": "my-service"}),
			},
			want: "k8s-app=my-service",
		},
		{
			name: "set-based selector when matched pods span multiple apps",
			pods: []v1.Pod{
				podWithLabels("api-gateway-abc", map[string]string{"app": "api-gateway"}),
				podWithLabels("gateway-worker-def", map[string]string{"app": "gateway-worker"}),
			},
			want: "app in (api-gateway,gateway-worker)",
		},
		{
			name: "set-based selector lists values sorted and deduplicated",
			pods: []v1.Pod{
				podWithLabels("zebra-abc", map[string]string{"app": "zebra"}),
				podWithLabels("alpha-def", map[string]string{"app": "alpha"}),
				podWithLabels("mango-ghi", map[string]string{"app": "mango"}),
				podWithLabels("alpha-jkl", map[string]string{"app": "alpha"}),
			},
			want: "app in (alpha,mango,zebra)",
		},
		{
			name: "errors when no known label key is present on all pods",
			pods: []v1.Pod{
				podWithLabels("my-service-abc", map[string]string{"team": "payments"}),
				podWithLabels("my-service-def", map[string]string{"team": "payments"}),
			},
			wantErr: true,
		},
		{
			name: "errors when a known key is present on only some pods",
			pods: []v1.Pod{
				podWithLabels("my-service-abc", map[string]string{"app": "my-service"}),
				podWithLabels("my-service-def", map[string]string{"team": "payments"}),
			},
			wantErr: true,
		},
		{
			name:    "errors on empty pod list",
			pods:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSelector(tt.pods)

			if tt.wantErr {
				if err == nil {
					t.Errorf("resolveSelector() expected an error, got selector %q", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolveSelector() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("resolveSelector() = %q, want %q", got, tt.want)
			}
		})
	}
}
