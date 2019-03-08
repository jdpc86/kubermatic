package machinecontroller

import (
	"fmt"

	kubermaticv1 "github.com/kubermatic/kubermatic/api/pkg/crd/kubermatic/v1"
	"github.com/kubermatic/kubermatic/api/pkg/resources"
	"github.com/kubermatic/kubermatic/api/pkg/resources/apiserver"
	"github.com/kubermatic/kubermatic/api/pkg/resources/certificates/triple"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	certutil "k8s.io/client-go/util/cert"
)

var (
	webhookResourceRequirements = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("32Mi"),
			corev1.ResourceCPU:    resource.MustParse("10m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("512Mi"),
			corev1.ResourceCPU:    resource.MustParse("100m"),
		},
	}
)

// WebhookDeploymentCreator returns the function to create and update the machine controller webhook deployment
func WebhookDeploymentCreator(data machinecontrollerData) resources.DeploymentCreator {
	return func(dep *appsv1.Deployment) (*appsv1.Deployment, error) {
		dep.Name = resources.MachineControllerWebhookDeploymentName
		dep.Labels = resources.BaseAppLabel(resources.MachineControllerWebhookDeploymentName, nil)
		dep.Spec.Replicas = resources.Int32(1)
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: resources.BaseAppLabel(resources.MachineControllerWebhookDeploymentName, nil),
		}
		dep.Spec.Strategy.Type = appsv1.RollingUpdateStatefulSetStrategyType
		dep.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
			MaxSurge: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 1,
			},
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 0,
			},
		}
		dep.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: resources.ImagePullSecretName}}

		volumes := []corev1.Volume{getKubeconfigVolume(), getServingCertVolume()}
		dep.Spec.Template.Spec.Volumes = volumes
		podLabels, err := data.GetPodTemplateLabels(resources.MachineControllerWebhookDeploymentName, volumes, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create pod labels: %v", err)
		}
		dep.Spec.Template.ObjectMeta = metav1.ObjectMeta{Labels: podLabels}

		apiserverIsRunningContainer, err := apiserver.IsRunningInitContainer(data)
		if err != nil {
			return nil, err
		}
		dep.Spec.Template.Spec.InitContainers = []corev1.Container{*apiserverIsRunningContainer}

		dep.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:            Name,
				Image:           data.ImageRegistry(resources.RegistryDocker) + "/kubermatic/machine-controller:" + tag,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/usr/local/bin/webhook"},
				Args: []string{
					"-kubeconfig", "/etc/kubernetes/kubeconfig/kubeconfig",
					"-logtostderr",
					"-v", "4",
					"-listen-address", "0.0.0.0:9876",
				},
				Env:                      getEnvVars(data),
				TerminationMessagePath:   corev1.TerminationMessagePathDefault,
				TerminationMessagePolicy: corev1.TerminationMessageReadFile,
				Resources:                webhookResourceRequirements,
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "/healthz",
							Port:   intstr.FromInt(9876),
							Scheme: corev1.URISchemeHTTPS,
						},
					},
					FailureThreshold: 3,
					PeriodSeconds:    10,
					SuccessThreshold: 1,
					TimeoutSeconds:   15,
				},
				LivenessProbe: &corev1.Probe{
					FailureThreshold: 8,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "/healthz",
							Port:   intstr.FromInt(9876),
							Scheme: corev1.URISchemeHTTPS,
						},
					},
					InitialDelaySeconds: 15,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      15,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      resources.MachineControllerKubeconfigSecretName,
						MountPath: "/etc/kubernetes/kubeconfig",
						ReadOnly:  true,
					},
					{
						Name:      resources.MachineControllerWebhookServingCertSecretName,
						MountPath: "/tmp/cert",
						ReadOnly:  true,
					},
				},
			},
		}

		return dep, nil
	}
}

// ServiceCreator returns the function to reconcile the DNS service
func ServiceCreator() resources.NamedServiceCreatorGetter {
	return func() (string, resources.ServiceCreator) {
		return resources.MachineControllerWebhookServiceName, func(se *corev1.Service) (*corev1.Service, error) {
			se.Name = resources.MachineControllerWebhookServiceName
			se.Labels = resources.BaseAppLabel(resources.MachineControllerWebhookDeploymentName, nil)

			se.Spec.Type = corev1.ServiceTypeClusterIP
			se.Spec.Selector = map[string]string{
				resources.AppLabelKey: resources.MachineControllerWebhookDeploymentName,
			}
			se.Spec.Ports = []corev1.ServicePort{
				{
					Name:       "",
					Port:       443,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(9876),
				},
			}

			return se, nil
		}
	}
}

func getServingCertVolume() corev1.Volume {
	return corev1.Volume{
		Name: resources.MachineControllerWebhookServingCertSecretName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  resources.MachineControllerWebhookServingCertSecretName,
				DefaultMode: resources.Int32(resources.DefaultAllReadOnlyMode),
			},
		},
	}
}

type tlsServingCertCreatorData interface {
	GetRootCA() (*triple.KeyPair, error)
	Cluster() *kubermaticv1.Cluster
}

// TLSServingCertificateCreator returns a function to create/update the secret with the machine-controller-webhook tls certificate
func TLSServingCertificateCreator(data tlsServingCertCreatorData) resources.NamedSecretCreatorGetter {
	return func() (string, resources.SecretCreator) {
		return resources.MachineControllerWebhookServingCertSecretName, func(se *corev1.Secret) (*corev1.Secret, error) {
			if se.Data == nil {
				se.Data = map[string][]byte{}
			}

			ca, err := data.GetRootCA()
			if err != nil {
				return nil, fmt.Errorf("failed to get root ca: %v", err)
			}
			commonName := fmt.Sprintf("%s.%s.svc.cluster.local.", resources.MachineControllerWebhookServiceName, data.Cluster().Status.NamespaceName)
			altNames := certutil.AltNames{
				DNSNames: []string{
					resources.MachineControllerWebhookServiceName,
					fmt.Sprintf("%s.%s", resources.MachineControllerWebhookServiceName, data.Cluster().Status.NamespaceName),
					commonName,
					fmt.Sprintf("%s.%s.svc", resources.MachineControllerWebhookServiceName, data.Cluster().Status.NamespaceName),
					fmt.Sprintf("%s.%s.svc.", resources.MachineControllerWebhookServiceName, data.Cluster().Status.NamespaceName),
				},
			}
			if b, exists := se.Data[resources.MachineControllerWebhookServingCertCertKeyName]; exists {
				certs, err := certutil.ParseCertsPEM(b)
				if err != nil {
					return nil, fmt.Errorf("failed to parse certificate (key=%s) from existing secret: %v", resources.MachineControllerWebhookServingCertCertKeyName, err)
				}
				if resources.IsServerCertificateValidForAllOf(certs[0], commonName, altNames, ca.Cert) {
					return se, nil
				}
			}

			newKP, err := triple.NewServerKeyPair(ca,
				commonName,
				resources.MachineControllerWebhookServiceName,
				data.Cluster().Status.NamespaceName,
				"",
				nil,
				// For some reason the name the APIServer validates against must be in the SANs, having it as CN is not enough
				[]string{commonName})
			if err != nil {
				return nil, fmt.Errorf("failed to generate serving cert: %v", err)
			}
			se.Data[resources.MachineControllerWebhookServingCertCertKeyName] = certutil.EncodeCertPEM(newKP.Cert)
			se.Data[resources.MachineControllerWebhookServingCertKeyKeyName] = certutil.EncodePrivateKeyPEM(newKP.Key)
			// Include the CA for simplicity
			se.Data[resources.CACertSecretKey] = certutil.EncodeCertPEM(ca.Cert)

			return se, nil
		}
	}
}

// MutatingwebhookConfiguration returns the MutatingwebhookConfiguration for the machine controler
func MutatingwebhookConfiguration(c *kubermaticv1.Cluster, data *resources.TemplateData, existing *admissionregistrationv1beta1.MutatingWebhookConfiguration) (*admissionregistrationv1beta1.MutatingWebhookConfiguration, error) {
	mutatingWebhookConfiguration := existing
	if mutatingWebhookConfiguration == nil {
		mutatingWebhookConfiguration = &admissionregistrationv1beta1.MutatingWebhookConfiguration{}
	}
	mutatingWebhookConfiguration.Name = resources.MachineControllerMutatingWebhookConfigurationName

	ca, err := data.GetRootCA()
	if err != nil {
		return nil, fmt.Errorf("failed to get root ca: %v", err)
	}

	mutatingWebhookConfiguration.Webhooks = []admissionregistrationv1beta1.Webhook{
		{
			Name:              fmt.Sprintf("%s-machinedeployments", resources.MachineControllerMutatingWebhookConfigurationName),
			NamespaceSelector: &metav1.LabelSelector{},
			FailurePolicy:     failurePolicyPtr(admissionregistrationv1beta1.Fail),
			Rules: []admissionregistrationv1beta1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1beta1.OperationType{admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update},
					Rule: admissionregistrationv1beta1.Rule{
						APIGroups:   []string{clusterAPIGroup},
						APIVersions: []string{clusterAPIVersion},
						Resources:   []string{"machinedeployments"},
					},
				},
			},
			ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
				URL:      strPtr(fmt.Sprintf("https://%s.%s.svc.cluster.local./machinedeployments", resources.MachineControllerWebhookServiceName, c.Status.NamespaceName)),
				CABundle: certutil.EncodeCertPEM(ca.Cert),
			},
		},
		{
			Name:              fmt.Sprintf("%s-machines", resources.MachineControllerMutatingWebhookConfigurationName),
			NamespaceSelector: &metav1.LabelSelector{},
			FailurePolicy:     failurePolicyPtr(admissionregistrationv1beta1.Fail),
			Rules: []admissionregistrationv1beta1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1beta1.OperationType{admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update},
					Rule: admissionregistrationv1beta1.Rule{
						APIGroups:   []string{clusterAPIGroup},
						APIVersions: []string{clusterAPIVersion},
						Resources:   []string{"machines"},
					},
				},
			},
			ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
				URL:      strPtr(fmt.Sprintf("https://%s.%s.svc.cluster.local./machines", resources.MachineControllerWebhookServiceName, c.Status.NamespaceName)),
				CABundle: certutil.EncodeCertPEM(ca.Cert),
			},
		},
	}

	return mutatingWebhookConfiguration, nil
}

func strPtr(a string) *string {
	return &a
}

func failurePolicyPtr(a admissionregistrationv1beta1.FailurePolicyType) *admissionregistrationv1beta1.FailurePolicyType {
	return &a
}
