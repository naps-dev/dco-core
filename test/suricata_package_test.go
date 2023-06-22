package test

import (
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
		if len(pods) > 0 {
			break
		} else if x == 29 {
			t.Errorf("Could not start Suricata pod (Timeout)")
		}
		time.Sleep(10 * time.Second)
		pods = k8s.ListPods(t, opts, metav1.ListOptions{})
		x += 1
	}
	k8s.WaitUntilPodAvailable(t, opts, pods[0].Name, 40, 30*time.Second)

	// Test that the pods are running on the correct agents
	agents := k8s.GetNodes(t, opts)
	var actualNodeTypes map[string]bool
	expectedNodeTypes := map[string]bool{"Tier-1": true, "Tier-2": true}
	for _, pod := range pods {
		//isRunningOnExpectedAgent := false
		//expectedNodeTypes := getAgentsWithLabel(agents, []string{"Tier-1", "Tier-2"})

		// Check if any expected agent exists
		if len(agents) == 0 {
			t.Errorf("Pod %s is running on an agent that is not in the agents list or does not have matching labels", pod.Name)
			continue
		}

		for _, agent := range agents {
			if isPodRunningOnAgent(pod, &agent) {
				//isRunningOnExpectedAgent = true
				actualNodeTypes[agent.Labels["cnaps.io/node-type"]] = true
				break
			}
		}

		if isEqual(expectedNodeTypes, actualNodeTypes) != true {
			t.Errorf("Pod %s is not running on any of the expected agents [%s]", pod.Name, expectedNodeTypes)
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

//func getAgentsWithLabel(agents []v1.Node, labels []string) []*v1.Node {
//	var matchingAgents []*v1.Node
//
//	for _, agent := range agents {
//		for _, label := range labels {
//			if agent.Labels["cnaps.io/node-type"] == label {
//				matchingAgents = append(matchingAgents, &agent)
//				break
//			}
//		}
//	}
//
//	return matchingAgents
//}

func isEqual(expected, actual map[string]bool) bool {
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
