package test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/stretchr/testify/require"
)

func TestZarfPackage(t *testing.T) {
	zarfPackage := os.Getenv("ZARF_PACKAGE")

	cwd, err := os.Getwd()

	if err != nil {
		t.Error("ERROR: Unable to determine working directory, exiting." + err.Error())
	} else {
		logger.Log(t, "Working directory: "+cwd)
	}

	// Additional test environment vars. Use this to make sure proper kubeconfig is being referenced by k3d
	testEnv := map[string]string{
		"KUBECONFIG": "/tmp/test_kubeconfig_dco_core",
	}

	clusterSetupCmd := shell.Command{
		Command: "k3d",
		Args: []string{"cluster", "create", "test-dco-core",
			"--k3s-arg", "--disable=traefik@server:*",
			"--port", "0:443@loadbalancer",
			"--port", "0:80@loadbalancer"},
		Env: testEnv,
	}

	clusterTeardownCmd := shell.Command{
		Command: "k3d",
		Args:    []string{"cluster", "delete", "test-dco-core"},
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

	// Wait for DCO elastic to come up
	opts := k8s.NewKubectlOptions("k3d-test-dco-core", "/tmp/test_kubeconfig_dco_core", "dataplane-ek")
	k8s.WaitUntilServiceAvailable(t, opts, "dataplane-ek-es-http", 40, 30*time.Second)

	// Check that Kyverno is successfully generating policy reports
	checkAlert := shell.Command{
		Command: "kubectl",
		Args:    []string{"get", "policyreport", "-A"},
		Env:     testEnv,
	}

	shell.RunCommand(t, checkAlert)

	// Get the port for the curl
	k3dInspect := docker.Inspect(t, "k3d-test-dco-core-serverlb")

	httpsPort := k3dInspect.GetExposedHostPort(443)
	httpsPortStr := strconv.Itoa(int(httpsPort))

	// Wait for Neuvector UI
	opts = k8s.NewKubectlOptions("k3d-test-dco-core", "/tmp/test_kubeconfig_dco_core", "neuvector")
	k8s.WaitUntilServiceAvailable(t, opts, "neuvector-service-webui", 50, 30*time.Second)

	// Attempt to force IP on public-ingressgateway service
	dataplaneResourcePath, err := filepath.Abs("dataplane-ingressgateway.yaml")
	require.NoError(t, err)

	// Attempt to force IP on passthrough-ingressgateway service
	passthroughResourcePath, err := filepath.Abs("passthrough-ingressgateway.yaml")
	require.NoError(t, err)

	opts = k8s.NewKubectlOptions("k3d-test-dco-core", "/tmp/test_kubeconfig_dco_core", "istio-system")
	retries := 0

	for retries = 0; retries < 5; retries++ {
		// Delete dataplane-ingressgateway to free up the IP and sleep a bit
		logger.Log(t, "Delete dataplane-ingressgateway")
		err = k8s.KubectlDeleteE(t, opts, dataplaneResourcePath)
		require.NoError(t, err)

		// Delete passthrough-ingressgateway to free up the IP and sleep a bit
		logger.Log(t, "Delete passthrough-ingressgateway")
		err = k8s.KubectlDeleteE(t, opts, passthroughResourcePath)
		require.NoError(t, err)

		logger.Log(t, "Sleep 45s")
		time.Sleep(45 * time.Second)

		// Get public-ingressgateway service
		logger.Log(t, "Check public-ingressgateway for LoadBalancer IP, attempt", retries+1)
		publicSvc := k8s.GetService(t, opts, "public-ingressgateway")

		if len(publicSvc.Status.LoadBalancer.Ingress) > 0 {
			retries = 0
			logger.Log(t, "Success! LoadBalancer IP is assigned to public-ingressgateway")
			break
		}
	}

	if retries > 0 {
		logger.Log(t, "Failed to align LoadBalancer IP with public-ingressgateway")
		t.FailNow()
	}

	curlCmd := shell.Command{
		Command: "curl",
		Args: []string{
			"-k",
			"-L",
			"https://neuvector.vp.bigbang.dev:" + httpsPortStr,
			"--resolve",
			"neuvector.vp.bigbang.dev:" + httpsPortStr + ":127.0.0.1",
			"--fail-with-body"},
		Env: testEnv,
	}

	t.Run("Neuvector UI is accessible through Istio", func(t *testing.T) {
		shell.RunCommand(t, curlCmd)
	})

	// Attempt to force IP on passthrough-ingressgateway service
	publicResourcePath, err := filepath.Abs("public-ingressgateway.yaml")
	require.NoError(t, err)

	retries = 0

	for retries = 0; retries < 5; retries++ {
		// Delete dataplane-ingressgateway to free up the IP and sleep a bit
		logger.Log(t, "Delete dataplane-ingressgateway")
		err = k8s.KubectlDeleteE(t, opts, dataplaneResourcePath)
		require.NoError(t, err)

		// Delete public-ingressgateway to free up the IP and sleep a bit
		logger.Log(t, "Delete public-ingressgateway")
		err = k8s.KubectlDeleteE(t, opts, publicResourcePath)
		require.NoError(t, err)

		logger.Log(t, "Sleep 45s")
		time.Sleep(45 * time.Second)

		// Get passthrough-ingressgateway service
		logger.Log(t, "Check passthrough-ingressgateway for LoadBalancer IP, attempt", retries+1)
		passthroughSvc := k8s.GetService(t, opts, "passthrough-ingressgateway")

		if len(passthroughSvc.Status.LoadBalancer.Ingress) > 0 {
			retries = 0
			logger.Log(t, "Success! LoadBalancer IP is assigned to passthrough-ingressgateway")
			break
		}
	}

	if retries > 0 {
		logger.Log(t, "Failed to align LoadBalancer IP with public-ingressgateway")
		t.FailNow()
	}

	curlCmd = shell.Command{
		Command: "curl",
		Args: []string{
			"-k",
			"-L",
			"https://keycloak.vp.bigbang.dev:" + httpsPortStr + "/auth",
			"--resolve",
			"keycloak.vp.bigbang.dev:" + httpsPortStr + ":127.0.0.1",
			"--fail-with-body"},
		Env: testEnv,
	}

	t.Run("Keycloak UI is accessible through Istio", func(t *testing.T) {
		shell.RunCommand(t, curlCmd)
	})
}
