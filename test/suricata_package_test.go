package test

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/shell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SuricataTestZarfPackage(t *testing.T, contextName string, kubeconfigPath string) {
	testEnv := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}

	zarfDeploySuricataCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../suricata/zarf-package-suricata-amd64.tar.zst", "--confirm", "--no-progress"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeploySuricataCmd)

	//Test pods come up
	opts := k8s.NewKubectlOptions(contextName, kubeconfigPath, "suricata")
	x := 0
	pods := k8s.ListPods(t, opts, metav1.ListOptions{})
	for x < 30 {
		if len(pods) > 1 {
			break
		} else if x == 29 {
			t.Errorf("Could not start Suricata pods (Timeout)")
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
	fmt.Printf("Pods available: [%d] \n", len(pods))
	for _, pod := range pods {
		fmt.Printf("Pod name: [%s] \n", pod.Name)

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

	//Test alert provided by suricata devs
	createAlert := shell.Command{
		Command: "kubectl",
		Args:    []string{"--namespace", "suricata", "exec", "-i", pods[0].Name, "--", "/bin/bash", "-c", "curl -A BlackSun www.google.com"},
		Env:     testEnv,
	}

	shell.RunCommand(t, createAlert)

	checkAlert := shell.Command{
		Command: "kubectl",
		Args:    []string{"--namespace", "suricata", "exec", "-i", pods[0].Name, "--", "/bin/bash", "-c", "tail /var/log/suricata/fast.log"},
		Env:     testEnv,
	}

	output := shell.RunCommandAndGetOutput(t, checkAlert)

	got := strings.Contains(output, "Suspicious User Agent")

	if got != true {
		t.Errorf("tail /var/log/suricata/fast.log did not contain \"Suspicious User Agent\"")
	}
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
