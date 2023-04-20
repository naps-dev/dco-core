package test

import (
    "os"
    "strconv"
    "testing"
    "time"
    "github.com/gruntwork-io/terratest/modules/docker"
    "github.com/gruntwork-io/terratest/modules/k8s"
    "github.com/gruntwork-io/terratest/modules/shell"  
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"  
)

func TestZarfPackage(t *testing.T) {
	cwd, err := os.Getwd()

	if err != nil {
		t.Error("ERROR: Unable to determine working directory, exiting." + err.Error())
	} else {
		t.Log("Working directory: " + cwd)
	}

	// Additional test environment vars. Use this to make sure proper kubeconfig is being referenced by k3d
	testEnv := map[string]string{
		"KUBECONFIG": "/tmp/test_kubeconfig_kasm",
	}

	clusterSetupCmd := shell.Command{
		Command: "k3d",
		Args: []string{"cluster", "create", "test-kasm",
			"--k3s-arg", "--disable=traefik@server:*",
			"--port", "0:443@loadbalancer",
			"--port", "0:80@loadbalancer"},
		Env: testEnv,
	}

	clusterTeardownCmd := shell.Command{
		Command: "k3d",
		Args:    []string{"cluster", "delete", "test-kasm"},
		Env:     testEnv,
	}

	// if this was already running, go ahead and tear it down now.
	shell.RunCommand(t, clusterTeardownCmd)

	// to leave cluster up for examination after this run, comment this out:
	defer shell.RunCommand(t, clusterTeardownCmd)

	shell.RunCommand(t, clusterSetupCmd)

	// Identify port being used to forward to internal HTTPS on the Host OS to the K3D Docker Container
	// equivalent to: docker inspect k3d-test-kasm-serverlb --format '{{(index .NetworkSettings.Ports "443/tcp" 0).HostPort}}'
	k3dInspect := docker.Inspect(t, "k3d-test-kasm-serverlb")

	httpPort := k3dInspect.GetExposedHostPort(80)
	httpPortStr := strconv.Itoa(int(httpPort))

	httpsPort := k3dInspect.GetExposedHostPort(443)
	httpsPortStr := strconv.Itoa(int(httpsPort))

	t.Log("Using HTTP Port  " + httpPortStr)
	t.Log("Using HTTPS Port " + httpsPortStr)

	zarfInitCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"init", "--components", "git-server", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfInitCmd)

	zarfDeployDCOCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../zarf-package-dco-core-amd64.tar.zst", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployDCOCmd)

	createNamespace := shell.Command{
        Command: "kubectl",
        Args:    []string{"create", "namespace", "kasm"},
        Env:     testEnv,
    }

    shell.RunCommand(t, createNamespace)

	// Wait for DCO elastic to come up before deploying kasm
	// Note that k3d calls the cluster test-kasm, but actual context is called k3d-test-kasm
	opts := k8s.NewKubectlOptions("k3d-test-kasm", "/tmp/test_kubeconfig_kasm", "dataplane-ek")
	k8s.WaitUntilServiceAvailable(t, opts, "dataplane-ek-es-http", 40, 30*time.Second)

	zarfDeployKasmCmd := shell.Command{
		Command: "zarf",
        Args:    []string{"package", "deploy", "../zarf-package-kasm-amd64.tar.zst", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployKasmCmd)

	// wait for service to come up before attempting to hit it
	opts = k8s.NewKubectlOptions("k3d-test-kasm", "/tmp/test_kubeconfig_kasm", "kasm")
    k8s.WaitUntilServiceAvailable(t, opts, "kasm-kasm-workspaces", 40, 30*time.Second)
    pods := k8s.ListPods(t, opts, metav1.ListOptions{})
    k8s.WaitUntilPodAvailable(t, opts, pods[0].Name, 40, 30*time.Second)

    // virtual service is set up as: kasm.vp.bigbang.dev
    // --fail-with-body used to fail on a 400 error which can happen when headers are incorrect.
    curlCmd := shell.Command{
        Command: "curl",
        Args:    []string{"--resolve", "kasm.vp.bigbang.dev:" + httpsPortStr + ":127.0.0.1",
                          "--fail-with-body",
                          "--insecure",
                          "https://kasm.vp.bigbang.dev:" + httpsPortStr },
        Env:     testEnv,
    }

    shell.RunCommand(t, curlCmd)
}
