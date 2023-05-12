## Zarf package for metallb (and associated configuration)

### Description

This folder and associated Zarf package contains two Zarf components: `metallb` and
`metallb-config`. These are intended for use as sub-components to the `dco-core`
parent package.

The first component `metallb` is a Zarf wrapper around the upstream
`metallb` Helm chart at [MetalLB's GitHub repo](https://github.com/metallb/metallb). 
It bundles the two `metallb` containers `speaker` and `controller`.

The second component `metallb-config` is a Zarf wrapper around _our own_ 
[Helm chart](charts/metallb-config) which contains an `IPAddressPool` and `L2Advertisement`. 

### Zarf variable flow

The IP range used in the `IPAddressPool` and the network interface used for the `L2Advertisement`
are passed down to the actual k8s objects via the following chain:

1. Overall `dco-core` Zarf package contains variables `METALLB_IP_ADDRESS_POOL` and
   `METALLB_INTERFACE`.
2. The `metallb` Zarf package gets those variables from the parent package and
   updates the `###ZARF_VAR_METALLB_IP_ADDRESS_POOL###` and `###ZARF_VAR_METALLB_INTERFACE###`
   sections in [manifests/metallb-config.yaml](manifests/metallb-config.yaml). This is
   used by Kustomize to populate the fields in [kustomizations/metallb-config/base/helmrelease.yaml]
   (kustomizations/metallb-config/base/helmrelease.yaml).
3. Kustomize populates the values from this into the Helm chart values at
   [charts/metallb-config/values.yaml](charts/metallb-config/values.yaml).
4. Helm uses the values there and renders them into the templates at
   [charts/metallb-config/templates](charts/metallb-config/templates/)

### Dependencies:

This assumes Flux has been deployed on a working Kubernetes cluster.
