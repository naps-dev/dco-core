package test

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/shell"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ArkimeTestZarfPackage(t *testing.T, contextName string, kubeconfigPath string) {
	testEnv := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}

	zarfDeployArkimeCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../arkime/zarf-package-arkime-amd64.tar.zst", "--confirm", "--no-progress"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployArkimeCmd)

	//Test pods come up
	opts := k8s.NewKubectlOptions(contextName, kubeconfigPath, "Arkime")
	x := 0
	pods := k8s.ListPods(t, opts, metav1.ListOptions{})
	for x < 30 {
		if len(pods) > 1 {
			break
		} else if x == 29 {
			t.Errorf("Could not start Arkime pods (Timeout)")
		}
		time.Sleep(10 * time.Second)
		pods = k8s.ListPods(t, opts, metav1.ListOptions{})
		x += 1
	}
	k8s.WaitUntilPodAvailable(t, opts, pods[0].Name, 40, 30*time.Second)
	k8s.WaitUntilPodAvailable(t, opts, pods[1].Name, 40, 30*time.Second)

	// Test that the pods are running on the correct agents
	agents := k8s.GetNodes(t, opts)
	actualNodeTypes := map[string]bool{}
	expectedNodeTypes := map[string]bool{"Tier-1": true, "Tier-2": true}
	for _, pod := range pods {
		for _, agent := range agents {
			fmt.Printf("Agent name: [%s] \n", agent.Name)
			if isPodRunningOnAgent(pod, &agent) {
				actualNodeTypes[agent.Labels["cnaps.io/node-type"]] = true
			}
		}
	}

	if isEqual(expectedNodeTypes, actualNodeTypes) != true {
		for k, v := range expectedNodeTypes {
			t.Errorf("Expected Node Type: %s, %t", k, v)
		}
		for k, v := range actualNodeTypes {
			t.Errorf("Actual Node Type: %s, %t", k, v)
		}
	}	

	// Wait for arkime service to come up before attempting to hit it
	opts := k8s.NewKubectlOptions(contextName, kubeconfigPath, "arkime")
	k8s.WaitUntilServiceAvailable(t, opts, "arkime-viewer", 40, 30*time.Second)

	// Determine IP used by the dataplane ingressgateway
	dataplane_igw := k8s.GetService(t, k8s.NewKubectlOptions(contextName, kubeconfigPath, "istio-system"), "dataplane-ingressgateway")
	loadbalancer_ip := dataplane_igw.Status.LoadBalancer.Ingress[0].IP

	// Once service is up, give another few seconds for the upstream to be healthy
	time.Sleep(30 * time.Second)

	//-------------------------------------------------------------------------
	// Sub-tests
	//-------------------------------------------------------------------------
	// virtual service is set up as: arkime-viewer.vp.bigbang.dev
	// --fail-with-body used to fail on a 400 error which can happen when headers are incorrect.
	curlCmd := shell.Command{
		Command: "curl",
		Args: []string{"--resolve", "arkime-viewer.vp.bigbang.dev:443:" + loadbalancer_ip,
			"--fail-with-body",
			"https://arkime-viewer.vp.bigbang.dev"},
		Env: testEnv,
	}

	t.Run("Arkime runs successfully w/ initial setup", func(t *testing.T) {

		shell.RunCommand(t, curlCmd)
	})

	t.Run("Arkime undeploys cleanly", func(t *testing.T) {
		zarfDeleteArkimeCmd := shell.Command{
			Command: "zarf",
			Args:    []string{"package", "remove", "../arkime/zarf-package-arkime-amd64.tar.zst", "--confirm", "--no-progress"},
			Env:     testEnv,
		}

		shell.RunCommand(t, zarfDeleteArkimeCmd)
	})

	t.Run("Arkime skips initial setup on re-deploy", func(t *testing.T) {
		shell.RunCommand(t, zarfDeployArkimeCmd)
	})

	t.Run("Arkime runs succesfully post initial setup", func(t *testing.T) {
		k8s.WaitUntilServiceAvailable(t, opts, "arkime-viewer", 40, 30*time.Second)
		time.Sleep(30 * time.Second)
		shell.RunCommand(t, curlCmd)
	})

	//-------------------------------------------------------------------------
	// @TODO: Sensor tests
	//-------------------------------------------------------------------------
	t.Run("Arkime sensor is running", func(t *testing.T) {
		pods := k8s.ListPods(t, opts, v1.ListOptions{
			LabelSelector: "k8s-app=arkime-sensor",
		})

		for _, pod := range pods {
			t.Log("Pod log: " + k8s.GetPodLogs(t, opts, &pod, ""))
		}
	})
}

func isEqual(expected map[string]bool, actual map[string]bool) bool {
	if len(expected) != len(actual) {
		return false
	}
	for k, v := range expected {
		if actual[k] != v {
			return false
		}
	}
	return true
}

func isPodRunningOnAgent(pod v1.Pod, agent *v1.Node) bool {
	return pod.Spec.NodeName == agent.Name
}