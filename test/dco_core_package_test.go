package test

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
)

func TestZarfPackage(t *testing.T) {
	component := os.Getenv("COMPONENT")
	refName := os.Getenv("REF_NAME")
	clusterName := component + "-test-" + refName
	tier1AgentName := "k3d" + clusterName + "-agent-1"
	kubeconfigPath := "/tmp/" + component + "_test_+" + refName + "_kubeconfig"

	// Truncate string if too long (k3d not happy with strings over 32 chars)
	if len(clusterName) > 32 {
		clusterName = clusterName[:32]
	}
	cwd, err := os.Getwd()

	if err != nil {
		t.Error("ERROR: Unable to determine working directory, exiting." + err.Error())
	} else {
		logger.Log(t, "Working directory: "+cwd)
	}

	// Additional test environment vars. Use this to make sure proper kubeconfig is being referenced by k3d
	testEnv := map[string]string{
		"KUBECONFIG": kubeconfigPath,
	}

	clusterSetupCmd := shell.Command{
		Command: "k3d",
		Args: []string{"cluster", "create", clusterName,
			"--k3s-arg", "--disable=traefik@server:*",
			"--k3s-arg", "--disable=servicelb@server:*",
			"--port", "0:443@loadbalancer",
			"--port", "0:80@loadbalancer",
			"--agents", "3",
			"--k3s-node-label", component + "-capture=true@agent:0",
			"--k3s-node-label", "hardware-tier=Tier1@agent:1"},
		Env: testEnv,
	}

	clusterTeardownCmd := shell.Command{
		Command: "k3d",
		Args:    []string{"cluster", "delete", clusterName},
		Env:     testEnv,
	}

	// if this was already running, go ahead and tear it down now.
	shell.RunCommand(t, clusterTeardownCmd)

	// to leave cluster up for examination after this run, comment this out:
	//defer shell.RunCommand(t, clusterTeardownCmd)

	// create the cluster
	shell.RunCommand(t, clusterSetupCmd)

	// set network ID to inspect
	contextName := "k3d-" + clusterName
	networkID := contextName

	// Get IP range we can use for metallb load balancer
	ipstart, ipend := DetermineIPRange(t, networkID)

	// Start up zarf
	zarfInitCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"init", "--components", "git-server", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfInitCmd)

	zarfDeployDCOCmd := shell.Command{
		Command: "zarf",
		Args: []string{"package", "deploy", "../dco-core/zarf-package-dco-core-amd64.tar.zst", "--confirm",
			"--components", "flux,bigbang,setup,kubevirt,cdi,metallb,metallb-config,dataplane-ek",
			"--set", "METALLB_IP_ADDRESS_POOL=" + ipstart.String() + "-" + ipend.String(),
		},
		Env: testEnv,
	}

	shell.RunCommand(t, zarfDeployDCOCmd)

	// Wait for DCO elastic to come up
	opts := k8s.NewKubectlOptions(contextName, kubeconfigPath, "dataplane-ek")
	k8s.WaitUntilServiceAvailable(t, opts, "dataplane-ek-es-http", 40, 30*time.Second)

	// Check that Kyverno is successfully generating policy reports
	checkAlert := shell.Command{
		Command: "kubectl",
		Args:    []string{"get", "policyreport", "-A"},
		Env:     testEnv,
	}
	shell.RunCommand(t, checkAlert)

	k8s.WaitUntilPodAvailable(t, opts, "dataplane-ek-es-master-0", 40, 30*time.Second)
	k8s.WaitUntilPodAvailable(t, opts, "dataplane-ek-es-data-0", 40, 30*time.Second)
	pods := k8s.ListPods(t, opts, metav1.ListOptions{})

	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		if nodeName != tier1AgentName {
			logger.Log(t, fmt.Sprintf("dataplane-ek Elasticsearch pod [%s] is not running on Tier1 node, failing test.", pod.Name))
			t.FailNow()
		} else {
			logger.Log(t, fmt.Sprintf("dataplane-ek Elasticsearch pod [%s] is not running on Tier1 node, failing test.", pod.Name))
		}
	}

	// Wait for Neuvector UI
	opts = k8s.NewKubectlOptions(contextName, kubeconfigPath, "neuvector")
	k8s.WaitUntilServiceAvailable(t, opts, "neuvector-service-webui", 50, 30*time.Second)

	opts = k8s.NewKubectlOptions(contextName, kubeconfigPath, "istio-system")
	retries := 0

	for retries = 0; retries < 5; retries++ {
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

	// Determine IP used by the public ingressgateway
	public_igw := k8s.GetService(t, k8s.NewKubectlOptions(contextName, kubeconfigPath, "istio-system"), "public-ingressgateway")
	public_lb_ip := public_igw.Status.LoadBalancer.Ingress[0].IP

	curlCmd := shell.Command{
		Command: "curl",
		Args: []string{
			"-k",
			"-L",
			"https://neuvector.vp.bigbang.dev:443",
			"--resolve",
			"neuvector.vp.bigbang.dev:443:" + public_lb_ip,
			"--fail-with-body"},
		Env: testEnv,
	}

	t.Run("Neuvector UI is accessible through Istio", func(t *testing.T) {
		shell.RunCommand(t, curlCmd)
	})

	retries = 0

	for retries = 0; retries < 5; retries++ {
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

	// Determine IP used by the passthrough ingressgateway
	passthrough_igw := k8s.GetService(t, k8s.NewKubectlOptions(contextName, kubeconfigPath, "istio-system"), "passthrough-ingressgateway")
	passthrough_lb_ip := passthrough_igw.Status.LoadBalancer.Ingress[0].IP

	time.Sleep(120 * time.Second)

	curlCmd = shell.Command{
		Command: "curl",
		Args: []string{
			"-k",
			"-L",
			"https://keycloak.vp.bigbang.dev:443/auth",
			"--resolve",
			"keycloak.vp.bigbang.dev:443:" + passthrough_lb_ip,
			"--fail-with-body"},
		Env: testEnv,
	}

	t.Run("Keycloak UI is accessible through Istio", func(t *testing.T) {
		shell.RunCommand(t, curlCmd)
	})

	if component == "arkime" {
		ArkimeTestZarfPackage(t, contextName, kubeconfigPath)
	}
	if component == "kasm" {
		KasmTestZarfPackage(t, contextName, kubeconfigPath)
	}
	if component == "mixmode" {
		MixmodeTestZarfPackage(t, contextName, kubeconfigPath)
	}
	if component == "mockingbird" {
		MockingbirdTestZarfPackage(t, contextName, kubeconfigPath)
	}
	if component == "polarity" {
		PolarityTestZarfPackage(t, contextName, kubeconfigPath)
	}
	if component == "suricata" {
		SuricataTestZarfPackage(t, contextName, kubeconfigPath)
	}
	if component == "xsoar" {
		XsoarTestZarfPackage(t, contextName, kubeconfigPath)
	}
}

// -------------------------------------------------------------------------
// DetermineIPRange returns the first and last IP in the subnet
// This is used to set the IP range for metallb
// -------------------------------------------------------------------------
func DetermineIPRange(t *testing.T, networkID string) (net.IP, net.IP) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Error("ERROR: Unable to create docker client, exiting." + err.Error())
	}

	network, err := cli.NetworkInspect(context.Background(), networkID, types.NetworkInspectOptions{})
	if err != nil {
		t.Error("ERROR: Unable to inspect network, exiting." + err.Error())
	}

	subnet := network.IPAM.Config[0].Subnet

	ipaddr, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		t.Error("ERROR: Unable to parse CIDR, exiting." + err.Error())
	}

	octets := ipaddr.To4()
	octets[2]++
	octets[3] = 0

	ipstart := net.IPv4(octets[0], octets[1], octets[2], octets[3])

	octets[3] = 255
	ipend := net.IPv4(octets[0], octets[1], octets[2], octets[3])

	if !ipnet.Contains(ipstart) || !ipnet.Contains(ipend) {
		t.Error("ERROR: unable to gonkulate IPs in the k3d subnet, exiting.")
	}
	return ipstart, ipend
}
