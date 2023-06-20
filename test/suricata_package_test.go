package test

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/logger"
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

	//Test suricata daemonsets comes up on all hardware tiers (node taint toleration test)
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		for _, agent := range agents {
			if agent.Name == nodeName {
				logger.Log(t, fmt.Sprintf("suricata pod [%s] is running on node [%s]", pod.Name, agent.Name))
			} else {
				logger.Log(t, fmt.Sprintf("suricata pod [%s] is not running on node [%s], failing test.", pod.Name, agent.Name))
				t.FailNow()
			}
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
