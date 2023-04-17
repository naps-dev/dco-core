package test

import (
    "os"
    "testing"
    "time"
    "strings"
    "github.com/gruntwork-io/terratest/modules/k8s"
    "github.com/gruntwork-io/terratest/modules/shell"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestZarfPackage(t *testing.T) {
    cwd, err := os.Getwd()

    if (err != nil){
        t.Error("ERROR: Unable to determine working directory, exiting." + err.Error())
    } else {
        t.Log("Working directory: " + cwd)
    }

    // Additional test environment vars. Use this to make sure proper kubeconfig is being referenced by k3d
    testEnv := map[string]string{
        "KUBECONFIG": "/tmp/test_kubeconfig_suricata",
    }

    clusterSetupCmd := shell.Command{
        Command: "k3d",
        Args:    []string{"cluster", "create", "test-suricata",
                          "--k3s-arg", "--disable=traefik@server:*",
                          "--port", "0:443@loadbalancer",
                          "--port", "0:80@loadbalancer"},
        Env:     testEnv,
    }

    clusterTeardownCmd := shell.Command{
        Command: "k3d",
        Args:    []string{"cluster", "delete", "test-suricata"},
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
        Args:    []string{"package", "deploy", "../zarf-package-dco-core-amd64.tar.zst", "--confirm"},
        Env:     testEnv,
    }

    shell.RunCommand(t, zarfDeployDCOCmd)

    // Wait for DCO elastic to come up before deploying suricata
    // Note that k3d calls the cluster test-suricata, but actual context is called k3d-test-suricata
    opts := k8s.NewKubectlOptions("k3d-test-suricata", "/tmp/test_kubeconfig_suricata", "dataplane-ek");
    k8s.WaitUntilServiceAvailable(t, opts, "dataplane-ek-es-http", 40, 30*time.Second)

    zarfDeploysuricataCmd := shell.Command{
        Command: "zarf",
        Args:    []string{"package", "deploy", "../zarf-package-suricata-amd64.tar.zst", "--confirm"},
        Env:     testEnv,
    }

    shell.RunCommand(t, zarfDeploysuricataCmd)

    //Test pods come up
    opts = k8s.NewKubectlOptions("k3d-test-suricata", "/tmp/test_kubeconfig_suricata", "suricata")
    x := 0
    pods := k8s.ListPods(t, opts, metav1.ListOptions{})
    for x < 30 {
        if len(pods) > 0 {
            break
        } else if x == 29 {
            t.Errorf("Could not start Suricata pod (Timeout)")
        }
        time.Sleep(10*time.Second)
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
