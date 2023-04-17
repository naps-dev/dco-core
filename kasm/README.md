# Kasm Workspaces Helm Chart
This is a proof-of-concept Helm chart for running Kasm workspaces in the cluster.

**Must be run with priveleged context**
This container requires root privileges to run successfully. An exception must be made in OPA Gatekeeper to allow this container to run.

Exception should be made for default/kasm-workspaces in the CRD K8sPSPPrivilegedContainer2 under constratins.gatekeeper.sh

### Note on Startup
Startup takes ~2 minutes for kasm to fully spin up -- the container will not be marked ready until startup is complete.

