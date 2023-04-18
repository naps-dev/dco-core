package test


import (
    "os"
    "testing"
    "time"
    "strings"
    "context"
    "net"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"

    "github.com/gruntwork-io/terratest/modules/k8s"
    "github.com/gruntwork-io/terratest/modules/shell"
    v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestZarfPackage(t *testing.T) {
	component := os.Getenv("COMPONENT")
	kubeconfigPath := "/tmp/" + component + "_test_kubeconfig"
	clusterName := "test-" + component

	cwd, err := os.Getwd()

	if err != nil {
		t.Error("ERROR: Unable to determine working directory, exiting." + err.Error())
	} else {
		t.Log("Working directory: " + cwd)
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
			"--agents", "2",
			"--k3s-node-label", component + "-capture=true@agent:0"},
		Env: testEnv,
	}

	clusterTeardownCmd := shell.Command{
		Command: "k3d",
		Args:    []string{"cluster", "delete", "test-" + component},
		Env:     testEnv,
	}

	// if this was already running, go ahead and tear it down now.
	shell.RunCommand(t, clusterTeardownCmd)

	// to leave cluster up for examination after this run, comment this out:
	defer shell.RunCommand(t, clusterTeardownCmd)

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
		Args: []string{"package", "deploy", "../zarf-package-dco-core-amd64.tar.zst",
			"--confirm",
			"--components", "flux,big-bang-core,setup,kubevirt,cdi,metallb,metallb-config,dataplane-ek",
			"--set", "METALLB_IP_ADDRESS_POOL=" + ipstart.String() + "-" + ipend.String(),
			// "--set", "METALLB_INTERFACE", ""
		},
		Env: testEnv,
	}

	shell.RunCommand(t, zarfDeployDCOCmd)

	// Wait for DCO elastic (Big Bang deployment) to come up before deploying our component
	// Note that k3d calls the cluster test-<component>, but actual context is called k3d-test-<component>
	opts := k8s.NewKubectlOptions(contextName, kubeconfigPath, "dataplane-ek")
	k8s.WaitUntilServiceAvailable(t, opts, "dataplane-ek-es-http", 40, 30*time.Second)

	zarfDeployComponentCmd := shell.Command{
		Command: "zarf",
		Args:    []string{"package", "deploy", "../zarf-package-" + component + "-amd64.tar.zst", "--confirm"},
		Env:     testEnv,
	}

	shell.RunCommand(t, zarfDeployComponentCmd)

    // BWM ARKIME SPECIFIC
    if component == "arkime" {
        // wait for arkime service to come up before attempting to hit it
        opts = k8s.NewKubectlOptions(contextName, kubeconfigPath, "arkime")
        k8s.WaitUntilServiceAvailable(t, opts, "arkime-viewer", 40, 30*time.Second)

        // Determine IP used by the dataplane ingressgateway
        dataplane_igw := k8s.GetService(t, k8s.NewKubectlOptions(contextName, kubeconfigPath, "istio-system"), "dataplane-ingressgateway")
        loadbalancer_ip := dataplane_igw.Status.LoadBalancer.Ingress[0].IP

        // Once service is up, give another few seconds for the upstream to be healthy
        time.Sleep(30 * time.Second)

        //-------------------------------------------------------------------------
        // Sub-tests
        //-------------------------------------------------------------------------
        // virtual service is set up as: arkime-viewer.vp.bigbang.dev
        // --fail-with-body used to fail on a 400 error which can happen when headers are incorrect.
        curlCmd := shell.Command{
            Command: "curl",
            Args: []string{"--resolve", "arkime-viewer.vp.bigbang.dev:443:" + loadbalancer_ip,
                "--fail-with-body",
                "https://arkime-viewer.vp.bigbang.dev"},
            Env: testEnv,
        }

        t.Run("Arkime runs successfully w/ initial setup", func(t *testing.T) {

            shell.RunCommand(t, curlCmd)
        })

        t.Run("Arkime undeploys cleanly", func(t *testing.T) {
            zarfDeleteArkimeCmd := shell.Command{
                Command: "zarf",
                Args:    []string{"package", "remove", "../zarf-package-arkime-amd64.tar.zst", "--confirm"},
                Env:     testEnv,
            }

            shell.RunCommand(t, zarfDeleteArkimeCmd)
        })

        t.Run("Arkime skips initial setup on re-deploy", func(t *testing.T) {
            shell.RunCommand(t, zarfDeployArkimeCmd)
        })

        t.Run("Arkime runs succesfully post initial setup", func(t *testing.T) {
            k8s.WaitUntilServiceAvailable(t, opts, "arkime-viewer", 40, 30*time.Second)
            time.Sleep(30 * time.Second)
            shell.RunCommand(t, curlCmd)
        })

        //-------------------------------------------------------------------------
        // @TODO: Sensor tests
        //-------------------------------------------------------------------------
        t.Run("Arkime sensor is running", func(t *testing.T) {
            pods := k8s.ListPods(t, opts, v1.ListOptions{
                LabelSelector: "k8s-app=arkime-sensor",
            })

            for _, pod := range pods {
                t.Log("Pod log: " + k8s.GetPodLogs(t, opts, &pod, ""))
            }
        })
    }
    if component == "suricata" {
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
