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

	agents := k8s.GetNodes(t, opts)
	for _, pod := range pods {
		agent := getAgent(pod.Spec.NodeName, agents)

		// Check if the agent exists in the agents list
		if agent == nil {
			t.Errorf("Pod %s is running on an agent that is not in the agents list", pod.Name)
			continue
		}

		// Check if it's running on the correct agent
		expectedLabels := map[string]string{
			"suricata-capture":   "true",
			"cnaps.io/node-type": "",
		}
		//pod = tier-1
		for _, label := range []string{"Tier-1", "Tier-2"} {
			expectedLabels["cnaps.io/node-type"] = label

			// Check labels
			for key, value := range expectedLabels {
				if nodeValue, ok := agent.Labels[key]; ok {
					if nodeValue != value {
						t.Errorf("Pod %s running on wrong agent [%s], expected label %s=%s, got %s=%s", pod.Name, agent.Name, key, value, key, nodeValue)
					}
				} else {
					t.Errorf("Pod %s running on wrong agent [%s], expected label %s=%s, but label is missing", pod.Name, agent.Name, key, value)
				}
			}

			// Check taints
			for _, taint := range agent.Spec.Taints {
				if taint.Key == "cnaps.io/node-taint" && taint.Value != "noncore:NoSchedule" && taint.Effect != v1.TaintEffectNoSchedule {
					t.Errorf("Pod %s running on wrong agent, expected taint %s=%s:%s", pod.Name, taint.Key, taint.Value, taint.Effect)
				}
			}
		}
	}

	// for each pod check if it is running on agents with the following labels:
	// - agent-0: labels: {"suricata-capture":"true", "cnaps.io/node-type": "Tier-1", } taints: {"cnaps.io/node-taint": "noncore:NoSchedule"}
	// - agent-1: labels: {"suricata-capture":"true", "cnaps.io/node-type": "Tier-2", } taints: {"cnaps.io/node-taint": "noncore:NoSchedule"}
	// if it's running on agents without the above labels, fail the test

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

func getAgent(nodeName string, agents []v1.Node) *v1.Node {
	for _, agent := range agents {
		if agent.Name == nodeName {
			return &agent
		}
	}
	return nil
}
