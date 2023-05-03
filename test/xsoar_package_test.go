package test

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/shell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func XsoarTestZarfPackage(t *testing.T, contextName string, kubeconfigPath string) {
	testEnv := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}

	createNamespace := shell.Command{
		Command: "kubectl",
		Args:    []string{"create", "namespace", "xsoar"},
		Env:     testEnv,
	}

	shell.RunCommand(t, createNamespace)

	injectLicenseFile := shell.Command{
		Command: "kubectl",
		Args:    []string{"create", "configmap", "xsoar-lic", "--from-file=/tmp/demisto.lic", "-n", "xsoar"},
		Env:     testEnv,
	}

	shell.RunCommand(t, injectLicenseFile)

	zarfDeployXsoarCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../xsoar/zarf-package-xsoar-amd64.tar.zst", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployXsoarCmd)

	// wait for xsoar service to come up before attempting to hit it
	opts := k8s.NewKubectlOptions(contextName, kubeconfigPath, "xsoar")
	k8s.WaitUntilServiceAvailable(t, opts, "xsoar", 40, 30*time.Second)
	pods := k8s.ListPods(t, opts, metav1.ListOptions{})
	k8s.WaitUntilPodAvailable(t, opts, pods[0].Name, 40, 30*time.Second)

	// Determine IP used by the dataplane ingressgateway
	dataplane_igw := k8s.GetService(t, k8s.NewKubectlOptions(contextName, kubeconfigPath, "istio-system"), "dataplane-ingressgateway")
	loadbalancer_ip := dataplane_igw.Status.LoadBalancer.Ingress[0].IP

	// Once service is up, give another few seconds for the upstream to be healthy
	time.Sleep(30 * time.Second)

	// virtual service is set up as: xsoar.vp.bigbang.dev
	// --fail-with-body used to fail on a 400 error which can happen when headers are incorrect.
	curlCmd := shell.Command{
		Command: "curl",
		Args: []string{"--resolve", "xsoar.vp.bigbang.dev:443:" + loadbalancer_ip,
			"--fail-with-body",
			"https://xsoar.vp.bigbang.dev"},
		Env: testEnv,
	}

	shell.RunCommand(t, curlCmd)
}
