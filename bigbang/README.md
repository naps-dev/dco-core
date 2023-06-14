## Zarf Package for Big Bang

### Description

This folder and associated Zarf package contains a single Zarf component: `bigbang`. This
is an _opinionated_ release of [Big Bang](https://docs-bigbang.dso.mil/latest/), which includes
only certain Big Bang applications.

### Zarf variables

This package defines these Zarf variables: `BIGBANG_VERSION` and `FLUX_VERSION` which specify the versions of
bigbang and flux, respectively, to be included in the zarf package.

### Additional Customization

Big Bang defines cascading Helm chart
[values](https://docs-bigbang.dso.mil/latest/docs/understanding-bigbang/configuration/base-config/#Values).
We selectively override/merge our own custom values provided in our
(values.yaml file)[./kustomizations/bigbang/values.yaml].

### Gateway Configuration

Big Bang provides [Istio](https://istio.io/) which we use for a Service Mesh
and also ingress gateways. Big Bang deployments typically define a single
ingress gateway `admin-ingressgateway`. We add two additional ingress
gateways: `dataplane-ingressgateway` and `tenant-ingressgateway`. The
data plane gateway is used to provide some level of isolation between the Big
Bang services and those added later.

Both the admin-ingressgateway and dataplane-ingressgateway will do TLS
termination at the gateway. The tenant gateway is used for traffic to
the keycloak service because keycloak insists on doing its own TLS termination.

> ⚠️ **The default k3d load balancer is
> unable to cope with multiple ingress gateways.** We recommend starting k3d
> with `--k3s-arg --disable=servicelb@server:*` and then running Metal LB on
> the cluster. This will allow you to run this package using all the gateways
> specified. For more details see
> [this](https://github.com/keunlee/k3d-metallb-starter-kit). This
> technique is used for the automated
> [unit test](../test/dco_core_package_test.go).

### Dependencies:

This requires Flux to be present on the Kubernetes cluster. In DCO Core, this
is done in the [dco-core umbrella package](../dco-core/zarf.yaml) prior to the
Big Bang component deployment.
