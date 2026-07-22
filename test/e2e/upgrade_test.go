//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kedav1alpha1 "github.com/kedacore/keda-olm-operator/api/keda/v1alpha1"
)

const (
	workloadNamespace = "keda-upgrade-test"

	scaledDeploymentName = "test-consumer"
	scaledObjectName     = "test-scaledobject"

	upgradePollInterval = 2 * time.Second
	upgradePollTimeout  = 5 * time.Minute
	scalingPollTimeout  = 3 * time.Minute
)

var (
	scaledObjectGVK = schema.GroupVersionKind{
		Group:   "keda.sh",
		Version: "v1alpha1",
		Kind:    "ScaledObject",
	}
)

// TestKedaUpgradeSetup runs under the previous KEDA operator version (installed
// by e2e-olm-upgrade-install). It creates a KedaController, deploys a workload
// with a ScaledObject, and verifies the scaling pipeline works. All resources
// are left in place so TestKedaUpgradeVerify can check them after the upgrade.
func TestKedaUpgradeSetup(t *testing.T) {
	ctx := t.Context()
	c := newClient(t)

	step(t, "install KEDA via KedaController", func(t *testing.T) {
		kc := &kedav1alpha1.KedaController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "keda",
				Namespace: namespace,
			},
		}
		if err := c.Create(ctx, kc); err != nil && !apierrors.IsAlreadyExists(err) {
			t.Fatalf("creating KedaController: %v", err)
		}
	})

	step(t, "core KEDA deployments become ready", func(t *testing.T) {
		for _, name := range coreDeployments {
			waitForDeploymentReady(t, ctx, c, name)
		}
	})

	step(t, "KedaController status shows success", func(t *testing.T) {
		version := waitForKedaControllerSuccess(t, ctx, c)
		t.Logf("pre-upgrade KEDA version: %s", version)
	})

	step(t, "log pre-upgrade deployment images", func(t *testing.T) {
		for _, name := range coreDeployments {
			logDeploymentImage(t, ctx, c, name)
		}
	})

	step(t, "create workload namespace", func(t *testing.T) {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: workloadNamespace},
		}
		if err := c.Create(ctx, ns); err != nil && !apierrors.IsAlreadyExists(err) {
			t.Fatalf("creating workload namespace: %v", err)
		}
	})

	step(t, "deploy scaling workload", func(t *testing.T) {
		createScalingWorkload(t, ctx, c)
	})

	step(t, "ScaledObject becomes ready under previous version", func(t *testing.T) {
		waitForScaledObjectReady(t, ctx, c)
	})

	step(t, "HPA created under previous version", func(t *testing.T) {
		verifyHPACreated(t, ctx, c)
	})
}

// TestKedaUpgradeVerify runs after the operator has been upgraded to the
// current build (by e2e-olm-upgrade-apply). It verifies that the ScaledObject
// and HPA created by TestKedaUpgradeSetup are still functional, then cleans up.
func TestKedaUpgradeVerify(t *testing.T) {
	ctx := t.Context()
	c := newClient(t)

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		cleanupUpgradeResources(t, cleanupCtx, c)
	})

	step(t, "core KEDA deployments ready after upgrade", func(t *testing.T) {
		for _, name := range coreDeployments {
			waitForDeploymentReady(t, ctx, c, name)
		}
	})

	step(t, "KedaController status shows success after upgrade", func(t *testing.T) {
		version := waitForKedaControllerSuccess(t, ctx, c)
		t.Logf("post-upgrade KEDA version: %s", version)
	})

	step(t, "log post-upgrade deployment images", func(t *testing.T) {
		for _, name := range coreDeployments {
			logDeploymentImage(t, ctx, c, name)
		}
	})

	step(t, "pre-existing ScaledObject still ready after upgrade", func(t *testing.T) {
		waitForScaledObjectReady(t, ctx, c)
	})

	step(t, "pre-existing HPA still present after upgrade", func(t *testing.T) {
		verifyHPACreated(t, ctx, c)
	})

	step(t, "cleanup upgrade test resources", func(t *testing.T) {
		cleanupUpgradeResources(t, ctx, c)
	})
}

func waitForKedaControllerSuccess(t *testing.T, ctx context.Context, c client.Client) string {
	t.Helper()
	var kc kedav1alpha1.KedaController
	err := wait.PollUntilContextTimeout(ctx, upgradePollInterval, upgradePollTimeout, true, func(ctx context.Context) (bool, error) {
		if err := c.Get(ctx, client.ObjectKey{Name: "keda", Namespace: namespace}, &kc); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("checking KedaController status: %w", err)
		}
		return kc.Status.Phase == kedav1alpha1.PhaseInstallSucceeded, nil
	})
	if err != nil {
		t.Fatalf("KedaController did not reach %q: %v", kedav1alpha1.PhaseInstallSucceeded, err)
	}
	return kc.Status.Version
}

func createScalingWorkload(t *testing.T, ctx context.Context, c client.Client) {
	t.Helper()

	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      scaledDeploymentName,
			Namespace: workloadNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": scaledDeploymentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": scaledDeploymentName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "consumer",
							Image:   "registry.k8s.io/pause:3.9",
							Command: []string{"/pause"},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("16Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("32Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
	if err := c.Create(ctx, dep); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("creating workload deployment: %v", err)
	}

	t.Log("waiting for workload deployment to become ready")
	err := wait.PollUntilContextTimeout(ctx, upgradePollInterval, upgradePollTimeout, true, func(ctx context.Context) (bool, error) {
		d := &appsv1.Deployment{}
		if err := c.Get(ctx, client.ObjectKey{Name: scaledDeploymentName, Namespace: workloadNamespace}, d); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		for _, cond := range d.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("workload deployment did not become ready: %v", err)
	}

	so := newScaledObject()
	if err := c.Create(ctx, so); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("creating ScaledObject: %v", err)
	}
}

func newScaledObject() *unstructured.Unstructured {
	so := &unstructured.Unstructured{}
	so.SetGroupVersionKind(scaledObjectGVK)
	so.SetName(scaledObjectName)
	so.SetNamespace(workloadNamespace)

	so.Object["spec"] = map[string]any{
		"scaleTargetRef": map[string]any{
			"name": scaledDeploymentName,
		},
		"minReplicaCount": int64(1),
		"maxReplicaCount": int64(5),
		"cooldownPeriod":  int64(10),
		"triggers": []any{
			map[string]any{
				"type":       "cpu",
				"metricType": string(autoscalingv2.UtilizationMetricType),
				"metadata": map[string]any{
					"value": "50",
				},
			},
		},
	}
	return so
}

func waitForScaledObjectReady(t *testing.T, ctx context.Context, c client.Client) {
	t.Helper()
	t.Log("waiting for ScaledObject to become ready")

	err := wait.PollUntilContextTimeout(ctx, upgradePollInterval, scalingPollTimeout, true, func(ctx context.Context) (bool, error) {
		so := &unstructured.Unstructured{}
		so.SetGroupVersionKind(scaledObjectGVK)
		if err := c.Get(ctx, client.ObjectKey{Name: scaledObjectName, Namespace: workloadNamespace}, so); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("getting ScaledObject: %w", err)
		}

		conditions, found, err := unstructured.NestedSlice(so.Object, "status", "conditions")
		if err != nil || !found {
			return false, nil
		}

		for _, c := range conditions {
			cond, ok := c.(map[string]any)
			if !ok {
				continue
			}
			condType, _, _ := unstructured.NestedString(cond, "type")
			condStatus, _, _ := unstructured.NestedString(cond, "status")
			if condType == "Ready" && condStatus == "True" {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("ScaledObject %s/%s did not become ready: %v", workloadNamespace, scaledObjectName, err)
	}
}

func verifyHPACreated(t *testing.T, ctx context.Context, c client.Client) {
	t.Helper()

	hpaName := "keda-hpa-" + scaledObjectName
	t.Logf("checking HPA %s/%s exists", workloadNamespace, hpaName)

	err := wait.PollUntilContextTimeout(ctx, upgradePollInterval, scalingPollTimeout, true, func(ctx context.Context) (bool, error) {
		hpa := &autoscalingv2.HorizontalPodAutoscaler{}
		if err := c.Get(ctx, client.ObjectKey{Name: hpaName, Namespace: workloadNamespace}, hpa); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("getting HPA: %w", err)
		}

		t.Logf("HPA %s: currentReplicas=%d, desiredReplicas=%d, minReplicas=%d, maxReplicas=%d",
			hpaName,
			hpa.Status.CurrentReplicas,
			hpa.Status.DesiredReplicas,
			*hpa.Spec.MinReplicas,
			hpa.Spec.MaxReplicas)

		return true, nil
	})
	if err != nil {
		t.Fatalf("HPA %s/%s was not created by KEDA: %v", workloadNamespace, hpaName, err)
	}
}

func cleanupUpgradeResources(t *testing.T, ctx context.Context, c client.Client) {
	t.Helper()

	so := &unstructured.Unstructured{}
	so.SetGroupVersionKind(scaledObjectGVK)
	so.SetName(scaledObjectName)
	so.SetNamespace(workloadNamespace)
	if err := c.Delete(ctx, so); err != nil && !apierrors.IsNotFound(err) {
		t.Logf("cleanup: failed to delete ScaledObject: %v", err)
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      scaledDeploymentName,
			Namespace: workloadNamespace,
		},
	}
	if err := c.Delete(ctx, dep); err != nil && !apierrors.IsNotFound(err) {
		t.Logf("cleanup: failed to delete workload deployment: %v", err)
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: workloadNamespace}}
	if err := c.Delete(ctx, ns); err != nil && !apierrors.IsNotFound(err) {
		t.Logf("cleanup: failed to delete workload namespace: %v", err)
	}

	kc := &kedav1alpha1.KedaController{}
	if err := c.Get(ctx, client.ObjectKey{Name: "keda", Namespace: namespace}, kc); err != nil {
		return
	}
	if err := c.Delete(ctx, kc); err != nil && !apierrors.IsNotFound(err) {
		t.Logf("cleanup: failed to delete KedaController: %v", err)
		return
	}
	waitForDeploymentGone(t, ctx, c, coreDeployments[0])
}
