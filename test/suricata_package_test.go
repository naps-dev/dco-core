package test

import (
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
		Args:    []string{"package", "deploy", "../suricata/zarf-package-suricata-amd64.tar.zst", "--confirm"},
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

	//Test alert provided by suricata devs
	createAlert := shell.Command{
		Command: "kubectl",
		Args:    []string{"--namespace", "suricata", "exec", "-it", pods[0].Name, "--", "/bin/bash", "-c", "curl -A BlackSun www.google.com"},
		Env:     testEnv,
	}

	shell.RunCommand(t, createAlert)

	checkAlert := shell.Command{
		Command: "kubectl",
		Args:    []string{"--namespace", "suricata", "exec", "-it", pods[0].Name, "--", "/bin/bash", "-c", "tail /var/log/suricata/fast.log"},
		Env:     testEnv,
	}

	output := shell.RunCommandAndGetOutput(t, checkAlert)

	got := strings.Contains(output, "Suspicious User Agent")

	if got != true {
		t.Errorf("tail /var/log/suricata/fast.log did not contain \"Suspicious User Agent\"")
	}
}