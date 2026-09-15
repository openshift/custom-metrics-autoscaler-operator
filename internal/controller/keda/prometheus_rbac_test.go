/*
Copyright 2020 The KEDA Authors

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
	if got := prometheusRBACName("openshift-keda"); got != "openshift-keda-prometheus" {
		t.Fatalf("prometheusRBACName() = %q, want openshift-keda-prometheus", got)
	}
}

func TestPrometheusRoleRules(t *testing.T) {
	rules := prometheusRoleRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]
	if len(rule.APIGroups) != 1 || rule.APIGroups[0] != "" {
		t.Fatalf("unexpected APIGroups: %#v", rule.APIGroups)
	}
	if len(rule.Resources) != 3 {
		t.Fatalf("unexpected resources: %#v", rule.Resources)
	}
	wantResources := map[string]bool{"services": true, "endpoints": true, "pods": true}
	for _, res := range rule.Resources {
		if !wantResources[res] {
			t.Fatalf("unexpected resource %q", res)
		}
	}
	if len(rule.Verbs) != 3 {
		t.Fatalf("unexpected verbs: %#v", rule.Verbs)
	}
	wantVerbs := map[string]bool{"get": true, "list": true, "watch": true}
	for _, verb := range rule.Verbs {
		if !wantVerbs[verb] {
			t.Fatalf("unexpected verb %q", verb)
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
			name: "enabled",
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{clusterMonitoringLabel: "true"},
				},
			},
			want: true,
		},
		{
			name: "disabled value",
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{clusterMonitoringLabel: "false"},
				},
			},
			want: false,
		},
		{
			name: "missing label",
			ns: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
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
				t.Fatalf("namespaceHasClusterMonitoringEnabled() = %v, want %v", got, tt.want)
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
	if subject.Name != "prometheus-k8s" || subject.Namespace != "openshift-monitoring" {
		t.Fatalf("unexpected subject: %#v", subject)
	}
}
