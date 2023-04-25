package test


import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/shell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func PolarityTestZarfPackage(t *testing.T, contextName string, kubeconfigPath string) {
	testEnv := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}

	createNamespace := shell.Command{
		Command: "kubectl",
		Args:    []string{"create", "namespace", "polarity"},
		Env:     testEnv,
	}

	shell.RunCommand(t, createNamespace)

	injectLicenseFile := shell.Command{
		Command: "kubectl",
		Args:    []string{"create", "secret", "generic", "polarity-server-lic", "--from-file=/tmp/polarity.lic", "-n", "polarity"},
		Env:     testEnv,
	}

	shell.RunCommand(t, injectLicenseFile)

	zarfDeployPolarityCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../polarity/zarf-package-polarity-amd64.tar.zst", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployPolarityCmd)

	// wait for polarity service to come up before attempting to hit it
	opts := k8s.NewKubectlOptions(contextName, kubeconfigPath, "polarity")
	k8s.WaitUntilServiceAvailable(t, opts, "polarity-web", 40, 30*time.Second)

	// Determine IP used by the dataplane ingressgateway
	dataplane_igw := k8s.GetService(t, k8s.NewKubectlOptions(contextName, kubeconfigPath, "istio-system"), "dataplane-ingressgateway")
	loadbalancer_ip := dataplane_igw.Status.LoadBalancer.Ingress[0].IP

	pods := k8s.ListPods(t, opts, metav1.ListOptions{})
	k8s.WaitUntilPodAvailable(t, opts, pods[0].Name, 40, 30*time.Second)

	// virtual service is set up as: polarity.vp.bigbang.dev
	// --fail-with-body used to fail on a 400 error which can happen when headers are incorrect.
	curlCmd := shell.Command{
		Command: "curl",
		Args: []string{"--resolve", "polarity.vp.bigbang.dev:443:" + loadbalancer_ip,
			"--fail-with-body",
			"https://polarity.vp.bigbang.dev"},
		Env: testEnv,
	}

	shell.RunCommand(t, curlCmd)
}
