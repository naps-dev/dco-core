package test

import (
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/shell"
)

func TestZarfPackage(t *testing.T) {
	zarfPackage := os.Getenv("MINIMAL_ZARF_PACKAGE")

	cwd, err := os.Getwd()

	if err != nil {
		t.Error("ERROR: Unable to determine working directory, exiting." + err.Error())
	} else {
		t.Log("Working directory: " + cwd)
	}

	// Additional test environment vars. Use this to make sure proper kubeconfig is being referenced by k3d
	testEnv := map[string]string{
		"KUBECONFIG": "/tmp/test_kubeconfig_dco_core_minimal",
	}

	clusterSetupCmd := shell.Command{
		Command: "k3d",
		Args: []string{"cluster", "create", "test-dco-core-minimal",
			"--k3s-arg", "--disable=traefik@server:*",
			"--port", "0:443@loadbalancer",
			"--port", "0:80@loadbalancer"},
		Env: testEnv,
	}

	clusterTeardownCmd := shell.Command{
		Command: "k3d",
		Args:    []string{"cluster", "delete", "test-dco-core-minimal"},
		Env:     testEnv,
	}

	// if this was already running, go ahead and tear it down now.
	shell.RunCommand(t, clusterTeardownCmd)

	// to leave cluster up for examination after this run, comment this out:
	defer shell.RunCommand(t, clusterTeardownCmd)

	shell.RunCommand(t, clusterSetupCmd)

	zarfInitCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"init", "--components", "git-server", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfInitCmd)

	zarfDeployDCOCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../" + zarfPackage, "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployDCOCmd)

	// Wait for DCO elastic (Big Bang minimal deployment) to come up
	opts := k8s.NewKubectlOptions("k3d-test-dco-core-minimal", "/tmp/test_kubeconfig_dco_core_minimal", "dataplane-ek")
	k8s.WaitUntilServiceAvailable(t, opts, "dataplane-ek-es-http", 40, 30*time.Second)
}
