/*
Copyright 2026 The KEDA Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package keda

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPrometheusRBACName(t *testing.T) {
	tests := []struct {
		namespace string
		want      string
	}{
		{"openshift-keda", "openshift-keda-prometheus"},
		{"keda", "keda-prometheus"},
		{"my-namespace", "my-namespace-prometheus"},
	}
	for _, tt := range tests {
		if got := prometheusRBACName(tt.namespace); got != tt.want {
			t.Errorf("prometheusRBACName(%q) = %q, want %q", tt.namespace, got, tt.want)
		}
	}
}

func TestPrometheusRoleRules(t *testing.T) {
	rules := prometheusRoleRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]

	if len(rule.APIGroups) != 1 || rule.APIGroups[0] != "" {
		t.Fatalf("unexpected APIGroups: %v", rule.APIGroups)
	}

	wantResources := map[string]bool{"services": true, "endpoints": true, "pods": true}
	if len(rule.Resources) != len(wantResources) {
		t.Fatalf("expected %d resources, got %d: %v", len(wantResources), len(rule.Resources), rule.Resources)
	}
	for _, res := range rule.Resources {
		if !wantResources[res] {
			t.Errorf("unexpected resource %q", res)
		}
	}

	wantVerbs := map[string]bool{"get": true, "list": true, "watch": true}
	if len(rule.Verbs) != len(wantVerbs) {
		t.Fatalf("expected %d verbs, got %d: %v", len(wantVerbs), len(rule.Verbs), rule.Verbs)
	}
	for _, verb := range rule.Verbs {
		if !wantVerbs[verb] {
			t.Errorf("unexpected verb %q", verb)
		}
	}
}

func TestNamespaceHasClusterMonitoringEnabled(t *testing.T) {
	tests := []struct {
		name string
		ns   *corev1.Namespace
		want bool
	}{
		{
			name: "label set to true",
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{clusterMonitoringLabel: "true"},
				},
			},
			want: true,
		},
		{
			name: "label set to false",
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{clusterMonitoringLabel: "false"},
				},
			},
			want: false,
		},
		{
			name: "label missing",
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
			want: false,
		},
		{
			name: "nil labels",
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{},
			},
			want: false,
		},
		{
			name: "nil namespace",
			ns:   nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := namespaceHasClusterMonitoringEnabled(tt.ns); got != tt.want {
				t.Errorf("namespaceHasClusterMonitoringEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrometheusRoleBindingSubjects(t *testing.T) {
	rb := &rbacv1.RoleBinding{
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      prometheusMonitoringSAName,
				Namespace: prometheusMonitoringSANamespace,
			},
		},
	}

	if len(rb.Subjects) != 1 {
		t.Fatalf("expected 1 subject, got %d", len(rb.Subjects))
	}
	subject := rb.Subjects[0]
	if subject.Kind != rbacv1.ServiceAccountKind {
		t.Errorf("subject.Kind = %q, want %q", subject.Kind, rbacv1.ServiceAccountKind)
	}
	if subject.Name != "prometheus-k8s" {
		t.Errorf("subject.Name = %q, want %q", subject.Name, "prometheus-k8s")
	}
	if subject.Namespace != "openshift-monitoring" {
		t.Errorf("subject.Namespace = %q, want %q", subject.Namespace, "openshift-monitoring")
	}
}
