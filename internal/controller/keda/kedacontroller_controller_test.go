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
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	mf "github.com/manifestival/manifestival"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	kedav1alpha1 "github.com/kedacore/keda-olm-operator/api/keda/v1alpha1"
)

var _ = Describe("Deploying KedaController manifest", func() {
	const (
		olmOperatorName      = "keda-olm-operator"
		operatorName         = "keda-operator"
		kedaManifestFilepath = "../../../config/samples/keda_v1alpha1_kedacontroller.yaml"
		timeout              = time.Second * 60
		interval             = time.Millisecond * 250
		namespace            = "keda"
	)

	var (
		ctx      = context.Background()
		err      error
		scheme   *runtime.Scheme
		manifest mf.Manifest
	)

	BeforeEach(func() {
		Eventually(func() error {
			_, err = getObject(ctx, "Pod", olmOperatorName, namespace, k8sClient)
			return err
		}, timeout, interval).Should(Succeed())

	})

	var _ = Describe("Changing namespace", func() {

		const ()

		BeforeEach(func() {
			scheme = k8sManager.GetScheme()
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			manifest, err = changeAttribute(manifest, "namespace", namespace, scheme, "")
			Expect(err).To(BeNil())

			Expect(manifest.Delete()).Should(Succeed())

			Eventually(func() error {
				_, err = getObject(ctx, "Pod", operatorName, namespace, k8sClient)
				return err
			}, timeout, interval).ShouldNot(Succeed())
		})

		Context("When deploying in \"keda\" namespace", func() {
			It("Should deploy KedaController", func() {

				Expect(manifest.Apply()).Should(Succeed())

				Eventually(func() error {
					_, err = getObject(ctx, "Pod", operatorName, namespace, k8sClient)
					return err
				}, timeout, interval).Should(Succeed())
			})
		})

		Context("When deploying not in \"keda\" namespace", func() {
			const changedNamespace = "default"

			It("Should not deploy KedaController", func() {

				manifest, err = changeAttribute(manifest, "namespace", changedNamespace, scheme, "")
				Expect(err).To(BeNil())

				Expect(manifest.Apply()).Should(Succeed())

				Eventually(func() error {
					_, err = getObject(ctx, "Pod", operatorName, namespace, k8sClient)
					return err
				}, timeout, interval).ShouldNot(Succeed())
			})
		})
	})
})

var _ = Describe("Testing functionality", func() {

	var _ = Describe("Default KedaController and Annotation Creation", func() {
		const (
			timeout                         = time.Second * 60
			interval                        = time.Millisecond * 250
			namespace                       = "keda"
			kedaDefaultControllerAnnotation = "keda-olm-operator/create-default-controller"
			deploymentName                  = "keda-operator"
		)

		var (
			ctx = context.Background()
			err error
		)

		Context("Default KedaController & annotation exist after operator installation", func() {
			It("Should find the correct annotation in the keda namespace", func() {
				kedaNamespace := &corev1.Namespace{}
				Eventually(func(g Gomega) {
					err = k8sClient.Get(ctx, types.NamespacedName{Name: namespace}, kedaNamespace)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(kedaNamespace.GetAnnotations()).To(HaveKeyWithValue(kedaDefaultControllerAnnotation, "true"))
				}, timeout, interval).Should(Succeed())
			})

			It("Should retrieve the default KedaController instance", func() {
				kedaInstance := &kedav1alpha1.KedaController{}
				Eventually(func() error {
					err = k8sClient.Get(ctx, types.NamespacedName{Name: namespace, Namespace: namespace}, kedaInstance)
					return err
				}, timeout, interval).Should(Succeed())
			})
		})
	})

	var _ = Describe("Changing operator parameters", func() {

		const (
			deploymentName       = "keda-operator"
			containerName        = "keda-operator"
			logLevelPrefix       = "--zap-log-level="
			kind                 = "KedaController"
			name                 = "keda"
			namespace            = "keda"
			kedaManifestFilepath = "../../../config/samples/keda_v1alpha1_kedacontroller.yaml"
		)

		var (
			ctx      = context.Background()
			timeout  = time.Second * 60
			interval = time.Millisecond * 250
			scheme   *runtime.Scheme
			manifest mf.Manifest
			err      error
			arg      string
			dep      = &appsv1.Deployment{}
		)

		BeforeEach(func() {
			By("Applying a manifest that defaults operator logLevel back to info")
			scheme = k8sManager.GetScheme()
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
			manifest, err = changeAttribute(manifest, "logLevel", "info", scheme, "default")
			Expect(err).To(BeNil())
			Expect(manifest.Apply()).Should(Succeed())

			By("Waiting for the operator deployment to reflect the changes")
			Eventually(func() error {
				return deploymentHasRolledOut(deploymentName, namespace, "default")
			}, timeout, interval).Should(Succeed())

			By("Checking to make sure loglevel is set back to info")
			u, err := getObject(ctx, "Deployment", deploymentName, namespace, k8sClient)
			Expect(err).To(BeNil())
			err = scheme.Convert(u, dep, nil)
			Expect(err).To(BeNil())

			arg, err = getDepArg(dep, logLevelPrefix, containerName)
			Expect(err).To(BeNil())
			Expect(arg).To(Equal("info"))
		})

		Context("When changing \"--zap-log-level\"", func() {
			variants := []struct {
				initialLogLevel string
				actualLogLevel  string
			}{
				{
					initialLogLevel: "debug",
					actualLogLevel:  "debug",
				},
				{
					initialLogLevel: "info",
					actualLogLevel:  "info",
				},
				{
					initialLogLevel: "error",
					actualLogLevel:  "error",
				},
				// the default in the sample kedacontroller manifest is "info", so "info" in this list can also mean "make sure it doesn't change",
				// it is supposed to ignore these two cases because the values are invalid.
				{
					initialLogLevel: "",
					actualLogLevel:  "info",
				},
				{
					initialLogLevel: "foo",
					actualLogLevel:  "info",
				},
			}

			for _, variant := range variants {
				caseName := fmt.Sprintf("Should change it, initialLoglevel='%s', actualLoglevel='%s'", variant.initialLogLevel, variant.actualLogLevel)
				It(caseName,
					func() {
						By(fmt.Sprintf("Setting operator loglevel to %s in kedaController manifest", variant.initialLogLevel))
						manifest, err = changeAttribute(manifest, "logLevel", variant.initialLogLevel, scheme, caseName)
						Expect(err).To(BeNil())
						err = manifest.Apply()
						Expect(err).To(BeNil())

						By("Waiting for the operator deployment to reflect the changes")
						Eventually(func() error {
							return deploymentHasRolledOut(deploymentName, namespace, caseName)
						}, timeout, interval).Should(Succeed())

						By("Checking to make sure the log level is " + variant.actualLogLevel)
						u, err := getObject(ctx, "Deployment", deploymentName, namespace, k8sClient)
						Expect(err).To(BeNil())
						err = scheme.Convert(u, dep, nil)
						Expect(err).To(BeNil())

						arg, err = getDepArg(dep, logLevelPrefix, containerName)
						Expect(err).To(BeNil())
						Expect(arg).To(Equal(variant.actualLogLevel))

					})
			}

		})
	})

	var _ = Describe("Changing webhook parameters", func() {

		const (
			deploymentName       = "keda-admission"
			containerName        = "keda-admission-webhooks"
			logLevelPrefix       = "--zap-log-level="
			kind                 = "KedaController"
			name                 = "keda"
			namespace            = "keda"
			kedaManifestFilepath = "../../../config/samples/keda_v1alpha1_kedacontroller.yaml"
		)

		var (
			ctx      = context.Background()
			timeout  = time.Second * 60
			interval = time.Millisecond * 250
			scheme   *runtime.Scheme
			manifest mf.Manifest
			err      error
			arg      string
			dep      = &appsv1.Deployment{}
		)

		BeforeEach(func() {
			By("Applying a manifest that defaults operator logLevel back to info")
			scheme = k8sManager.GetScheme()
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
			manifest, err = changeAttribute(manifest, "logLevel", "info", scheme, "default")
			Expect(err).To(BeNil())
			Expect(manifest.Apply()).Should(Succeed())

			By("Waiting for the operator deployment to reflect the changes")
			Eventually(func() error {
				return deploymentHasRolledOut(deploymentName, namespace, "default")
			}, timeout, interval).Should(Succeed())

			By("Checking to make sure loglevel is set back to info")
			u, err := getObject(ctx, "Deployment", deploymentName, namespace, k8sClient)
			Expect(err).To(BeNil())
			err = scheme.Convert(u, dep, nil)
			Expect(err).To(BeNil())

			arg, err = getDepArg(dep, logLevelPrefix, containerName)
			Expect(err).To(BeNil())
			Expect(arg).To(Equal("info"))
		})

		Context("When changing \"--zap-log-level\"", func() {
			variants := []struct {
				initialLogLevel string
				actualLogLevel  string
			}{
				{
					initialLogLevel: "debug",
					actualLogLevel:  "debug",
				},
				{
					initialLogLevel: "info",
					actualLogLevel:  "info",
				},
				{
					initialLogLevel: "error",
					actualLogLevel:  "error",
				},
				// the default in the sample kedacontroller manifest is "info", so "info" in this list can also mean "make sure it doesn't change",
				// it is supposed to ignore these two cases because the values are invalid.
				{
					initialLogLevel: "",
					actualLogLevel:  "info",
				},
				{
					initialLogLevel: "foo",
					actualLogLevel:  "info",
				},
			}

			for _, variant := range variants {
				caseName := fmt.Sprintf("Should change it, initialLoglevel='%s', actualLoglevel='%s'", variant.initialLogLevel, variant.actualLogLevel)

				It(caseName, func() {
					By(fmt.Sprintf("Setting admission loglevel to %s in kedaController manifest", variant.initialLogLevel))
					manifest, err = changeAttribute(manifest, "logLevel-admission", variant.initialLogLevel, scheme, caseName)
					Expect(err).To(BeNil())
					err = manifest.Apply()
					Expect(err).To(BeNil())

					By("Waiting for the admission deployment to reflect the changes")
					Eventually(func() error {
						return deploymentHasRolledOut(deploymentName, namespace, caseName)
					}, timeout, interval).Should(Succeed())

					By("Checking to make sure the log level is " + variant.actualLogLevel)
					u, err := getObject(ctx, "Deployment", deploymentName, namespace, k8sClient)
					Expect(err).To(BeNil())
					err = scheme.Convert(u, dep, nil)
					Expect(err).To(BeNil())

					arg, err = getDepArg(dep, logLevelPrefix, containerName)
					Expect(err).To(BeNil())
					Expect(arg).To(Equal(variant.actualLogLevel))
				})
			}
		})
	})
})

var _ = Describe("Testing audit flags", func() {
	const (
		metricsServerName    = "keda-metrics-apiserver"
		kind                 = "KedaController"
		name                 = "keda"
		namespace            = "keda"
		kedaManifestFilepath = "../../../config/samples/keda_v1alpha1_kedacontroller.yaml"
	)

	var (
		ctx      = context.Background()
		timeout  = time.Second * 60
		interval = time.Millisecond * 250
		scheme   *runtime.Scheme
		manifest mf.Manifest
		err      error
		dep      = &appsv1.Deployment{}
	)

	When("Manipulating parameters", func() {

		// TODO(jkyros): come back to refactor this test, it doesn't proeprly reset between test cases, and
		// if it did, it would be failing, because it's currently impossible to deconfigure audit because of
		// how our manifestival code only operates if the settings are different from default/empty.
		BeforeEach(func() {
			scheme = k8sManager.GetScheme()
			dep = &appsv1.Deployment{}
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
		})

		Context("to add audit configuration flags", func() {
			vars := []struct {
				argument string
				prefix   string
				value    string
			}{
				{
					argument: "auditLogFormat",
					prefix:   "--audit-log-format=",
					value:    "json",
				},
				{
					argument: "auditMaxAge",
					prefix:   "--audit-log-maxage=",
					value:    "1",
				},
				{
					argument: "auditMaxBackup",
					prefix:   "--audit-log-maxbackup=",
					value:    "2",
				},
				{
					argument: "auditLogMaxSize",
					prefix:   "--audit-log-maxsize=",
					value:    "3",
				},
			}
			for _, variant := range vars {
				caseName := fmt.Sprintf("adds '%s' with value '%s'", variant.argument, variant.value)
				It(caseName, func() {
					manifest, err := changeAttribute(manifest, variant.argument, variant.value, scheme, caseName)
					Expect(err).To(BeNil())

					Expect(manifest.Apply()).To(Succeed())
					Eventually(func() error {
						return deploymentHasRolledOut(metricsServerName, namespace, caseName)

					}, timeout, interval).Should(Succeed())
					u, err := getObject(ctx, "Deployment", metricsServerName, namespace, k8sClient)
					Expect(err).To(BeNil())

					Expect(scheme.Convert(u, dep, nil)).To(Succeed())
					arg, err := getDepArg(dep, variant.prefix, metricsServerName)
					Expect(err).To(BeNil())
					Expect(arg).To(Equal(variant.value))
				})
			}

		})
	})

	var _ = Describe("Configuring replicas", func() {
		const (
			kedaManifestFilepath  = "../../../config/samples/keda_v1alpha1_kedacontroller.yaml"
			operatorName          = "keda-operator"
			metricsServerName     = "keda-metrics-apiserver"
			admissionWebhooksName = "keda-admission"
			namespace             = "keda"
		)

		var (
			ctx      = context.Background()
			timeout  = time.Second * 60
			interval = time.Millisecond * 250
			scheme   *runtime.Scheme
			manifest mf.Manifest
			err      error
			dep      = &appsv1.Deployment{}
		)

		BeforeEach(func() {
			scheme = k8sManager.GetScheme()
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
		})

		Context("When setting operator.replicas", func() {
			It("Should set replicas to 2 in keda-operator Deployment", func() {
				By("Setting operator.replicas to 2")
				manifest, err = changeAttribute(manifest, "operatorReplicas", "2", scheme, "operator-replicas-2")
				Expect(err).To(BeNil())
				Expect(manifest.Apply()).Should(Succeed())

				By("Waiting for the operator deployment to reflect the changes")
				Eventually(func() error {
					return deploymentHasRolledOut(operatorName, namespace, "operator-replicas-2")
				}, timeout, interval).Should(Succeed())

				By("Checking that the deployment has 2 replicas")
				u, err := getObject(ctx, "Deployment", operatorName, namespace, k8sClient)
				Expect(err).To(BeNil())
				err = scheme.Convert(u, dep, nil)
				Expect(err).To(BeNil())
				Expect(dep.Spec.Replicas).ToNot(BeNil())
				Expect(*dep.Spec.Replicas).To(Equal(int32(2)))
			})

			It("Should default to 1 replica when replicas is not set (nil)", func() {
				By("Setting operator.replicas to nil (empty string)")
				manifest, err = changeAttribute(manifest, "operatorReplicas", "", scheme, "operator-replicas-nil")
				Expect(err).To(BeNil())
				Expect(manifest.Apply()).Should(Succeed())

				By("Waiting for the operator deployment to reflect the changes")
				Eventually(func() error {
					return deploymentHasRolledOut(operatorName, namespace, "operator-replicas-nil")
				}, timeout, interval).Should(Succeed())

				By("Checking that the deployment defaults to 1 replica")
				u, err := getObject(ctx, "Deployment", operatorName, namespace, k8sClient)
				Expect(err).To(BeNil())
				err = scheme.Convert(u, dep, nil)
				Expect(err).To(BeNil())
				Expect(dep.Spec.Replicas).ToNot(BeNil())
				Expect(*dep.Spec.Replicas).To(Equal(int32(1)))
			})

			It("Should update replicas when changed after initial deployment", func() {
				By("Initially setting operator.replicas to 1")
				manifest, err = changeAttribute(manifest, "operatorReplicas", "1", scheme, "operator-replicas-update-1")
				Expect(err).To(BeNil())
				Expect(manifest.Apply()).Should(Succeed())

				Eventually(func() error {
					return deploymentHasRolledOut(operatorName, namespace, "operator-replicas-update-1")
				}, timeout, interval).Should(Succeed())

				By("Changing operator.replicas to 2")
				manifest, err = changeAttribute(manifest, "operatorReplicas", "2", scheme, "operator-replicas-update-2")
				Expect(err).To(BeNil())
				Expect(manifest.Apply()).Should(Succeed())

				By("Waiting for the deployment to update")
				Eventually(func() error {
					return deploymentHasRolledOut(operatorName, namespace, "operator-replicas-update-2")
				}, timeout, interval).Should(Succeed())

				By("Checking that the deployment now has 2 replicas")
				u, err := getObject(ctx, "Deployment", operatorName, namespace, k8sClient)
				Expect(err).To(BeNil())
				err = scheme.Convert(u, dep, nil)
				Expect(err).To(BeNil())
				Expect(dep.Spec.Replicas).ToNot(BeNil())
				Expect(*dep.Spec.Replicas).To(Equal(int32(2)))
			})
		})

		Context("When setting metricsServer.replicas", func() {
			It("Should set replicas to 2 in keda-metrics-apiserver Deployment", func() {
				By("Setting metricsServer.replicas to 2")
				manifest, err = changeAttribute(manifest, "metricsServerReplicas", "2", scheme, "metricsserver-replicas-2")
				Expect(err).To(BeNil())
				Expect(manifest.Apply()).Should(Succeed())

				By("Waiting for the metrics server deployment to reflect the changes")
				Eventually(func() error {
					return deploymentHasRolledOut(metricsServerName, namespace, "metricsserver-replicas-2")
				}, timeout, interval).Should(Succeed())

				By("Checking that the deployment has 2 replicas")
				u, err := getObject(ctx, "Deployment", metricsServerName, namespace, k8sClient)
				Expect(err).To(BeNil())
				err = scheme.Convert(u, dep, nil)
				Expect(err).To(BeNil())
				Expect(dep.Spec.Replicas).ToNot(BeNil())
				Expect(*dep.Spec.Replicas).To(Equal(int32(2)))
			})
		})

		Context("When setting admissionWebhooks.replicas", func() {
			It("Should set replicas to 2 in keda-admission Deployment", func() {
				By("Setting admissionWebhooks.replicas to 2")
				manifest, err = changeAttribute(manifest, "admissionWebhooksReplicas", "2", scheme, "admission-replicas-2")
				Expect(err).To(BeNil())
				Expect(manifest.Apply()).Should(Succeed())

				By("Waiting for the admission webhooks deployment to reflect the changes")
				Eventually(func() error {
					return deploymentHasRolledOut(admissionWebhooksName, namespace, "admission-replicas-2")
				}, timeout, interval).Should(Succeed())

				By("Checking that the deployment has 2 replicas")
				u, err := getObject(ctx, "Deployment", admissionWebhooksName, namespace, k8sClient)
				Expect(err).To(BeNil())
				err = scheme.Convert(u, dep, nil)
				Expect(err).To(BeNil())
				Expect(dep.Spec.Replicas).ToNot(BeNil())
				Expect(*dep.Spec.Replicas).To(Equal(int32(2)))
			})
		})
	})
})

func getDepArg(dep *appsv1.Deployment, prefix string, containerName string) (string, error) {
	for _, container := range dep.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			for _, arg := range container.Args {
				if strings.HasPrefix(arg, prefix) {
					return strings.TrimPrefix(arg, prefix), nil
				}
			}
			return "", errors.New("Could not find an argument with given prefix: " + prefix)
		}
	}
	return "", errors.New("Could not find a container: " + containerName)
}

func changeAttribute(manifest mf.Manifest, attr string, value string, scheme *runtime.Scheme, annotation string) (mf.Manifest, error) {
	transformer := func(u *unstructured.Unstructured) error {
		kedaControllerInstance := &kedav1alpha1.KedaController{}
		if err := scheme.Convert(u, kedaControllerInstance, nil); err != nil {
			return err
		}

		// Annotations might be nil, so we need to make sure we account for that
		if kedaControllerInstance.Spec.Operator.DeploymentAnnotations == nil {
			kedaControllerInstance.Spec.Operator.DeploymentAnnotations = make(map[string]string)
		}

		if kedaControllerInstance.Spec.AdmissionWebhooks.DeploymentAnnotations == nil {
			kedaControllerInstance.Spec.AdmissionWebhooks.DeploymentAnnotations = make(map[string]string)
		}
		if kedaControllerInstance.Spec.MetricsServer.DeploymentAnnotations == nil {
			kedaControllerInstance.Spec.MetricsServer.DeploymentAnnotations = make(map[string]string)
		}
		// When we push through an attribute change, we also set an annotation matching the test case
		// so we can be sure the controller reacted to our kedaControllerInstance updates and we have "our"
		// changes, not just some changes that might be from a previous test case
		kedaControllerInstance.Spec.Operator.DeploymentAnnotations["testCase"] = annotation
		kedaControllerInstance.Spec.AdmissionWebhooks.DeploymentAnnotations["testCase"] = annotation
		kedaControllerInstance.Spec.MetricsServer.DeploymentAnnotations["testCase"] = annotation

		switch attr {
		case "namespace":
			kedaControllerInstance.Namespace = value
		case "logLevel":
			kedaControllerInstance.Spec.Operator.LogLevel = value
		// TODO(jkyros): this breaks pattern with the rest of these cases but multiple operands have
		// the same field, but we kind of bolted the admission tests on here without doing a refactor and
		// this makes it work for now
		case "logLevel-admission":
			kedaControllerInstance.Spec.AdmissionWebhooks.LogLevel = value
		// metricsServer audit arguments
		case "auditLogFormat":
			kedaControllerInstance.Spec.MetricsServer.LogFormat = value
		case "auditMaxAge":
			kedaControllerInstance.Spec.MetricsServer.MaxAge = value
		case "auditMaxBackup":
			kedaControllerInstance.Spec.MetricsServer.MaxBackup = value
		case "auditLogMaxSize":
			kedaControllerInstance.Spec.MetricsServer.MaxSize = value
		case "operatorReplicas":
			if value == "" {
				kedaControllerInstance.Spec.Operator.Replicas = nil
			} else {
				replicas, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return err
				}
				replicasInt32 := int32(replicas)
				kedaControllerInstance.Spec.Operator.Replicas = &replicasInt32
			}
		case "metricsServerReplicas":
			if value == "" {
				kedaControllerInstance.Spec.MetricsServer.Replicas = nil
			} else {
				replicas, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return err
				}
				replicasInt32 := int32(replicas)
				kedaControllerInstance.Spec.MetricsServer.Replicas = &replicasInt32
			}
		case "admissionWebhooksReplicas":
			if value == "" {
				kedaControllerInstance.Spec.AdmissionWebhooks.Replicas = nil
			} else {
				replicas, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return err
				}
				replicasInt32 := int32(replicas)
				kedaControllerInstance.Spec.AdmissionWebhooks.Replicas = &replicasInt32
			}
		default:
			return errors.New("Not a valid attribute")
		}
		return scheme.Convert(kedaControllerInstance, u, nil)
	}

	return manifest.Transform(transformer)
}

// deploymentHasRolledOut waits for the specified deployment to possess the specified annotation
//
//nolint:unparam
func deploymentHasRolledOut(deploymentName string, namespace string, deploymentAnnotation string) error {
	u, err := getObject(ctx, "Deployment", deploymentName, namespace, k8sClient)
	if err != nil {
		return err
	}
	// The default manifest has no annotation, so I'm not checking whether it's present, only the value,
	// because the default case will not have the annotation, and we want that to be okay in the case where we
	// apply the default.
	testcase := u.GetAnnotations()["testCase"]
	if deploymentAnnotation == testcase {
		By("Observing that the test case annotation is now: " + testcase)
		return nil
	}
	return fmt.Errorf("Deployment has not rolled out, annotation is still '%s'", testcase)
}
