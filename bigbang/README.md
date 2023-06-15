# Big Bang Zarf Package

## Quickstart

1. Build
    ```bash
    zarf package create --set GIT_REF=refs/heads/main --confirm
    ```
1. Deploy
    ```bash
    zarf package deploy --confirm
    ```

## Description

The `big bang` Zarf package is responsible for packaging and deploying PlatformOne's `big bang`. This is an _opinionated_ release of [Big Bang](https://docs-bigbang.dso.mil/latest/), which includes only certain applications.

This package is a child to the [dco-core](../dco-core/) Zarf package.

| Application | Enabled |
| -- | -- |
| Network Policies | x |
| Kiali | x |
| Cluster Auditor | x |
| Gatekeeper | x |
| Istio | &check; |
| Istio Operator | &check; |
| Jaeger | &check; |
| Kyverno | &check; |
| Kyverno Policies | &check; |
| Kyverno Reporter | x |
| ElasticSearch and Kibana | &check; |
| ECK Operator | &check; |
| FluentBit | &check; |
| Promtail | x |
| Loki | x |
| Neuvector | &check; |
| Tempo | x |
| Monitoring - Prometheus, Grafana, and Alert Manager | &check; |
| Twistlock | x |

Add-ons
| Add-on | Enabled |
| -- | -- |
| ArgoCD | x |
| AuthService | x |
| MinIO Operator | x |
| MinIO | x |
| GitLab | x |
| GitLab Runner | x |
| Nexus | x |
| SonarQube | x |
| HA Proxy | x |
| Anchore | x |
| Mattermost Operator | x |
| Mattermost | x |
| Velero | x |
| Keycloak | &check; |
| Vault | x |
| Metrics Server | x |

## Build

### Zarf Constants

| Name | Type | Purpose |
|--|--|--|
| `GIT_REF` | `string` | Provide the BRANCH (refs/heads/BRANCH) or TAG (refs/tags/TAG) git ref to identify the git reference to deploy |

### Steps

Create the package
```bash
# For a specific branch
zarf package create --set GIT_REF=refs/heads/main

# For a specific tag
zarf package create --set GIT_REF=refs/tags/v2.1.0
```

## Installation

### Prerequisites

* Kubernetes cluster
* Cluster must have Flux installed
* Cluster must have the Zarf initilization package deployed

_Note: the [dco-core](../dco-core/) Zarf package includes these prerequisites and will deploy this package_

### Zarf Variables

| Name | Type | Purpose | Default |
|--|--|--|--|
| `DOMAIN` | `string` | Domain tied to Istio ingress and other parts of the Big Bang umbrella chart | `vp.bigbang.dev` |

### Steps

Deploy the package
```bash
zarf package deploy

# Optionally specify a different domain. The certificates packaged in kustomizations/bigbang/environment-bb-secret.yaml must match the provided domain
zarf package deploy --set DOMAIN=your.domain.here
```

## Usage

Please see the official [Big Bang](https://docs-bigbang.dso.mil/latest/) umbrella chart documents for detailed configuration and usage information.

## FAQ

### How are the Istio Ingress Gateways Configured?

Big Bang provides [Istio](https://istio.io/) which we use for a Service Mesh
and also ingress gateways. Big Bang deployments typically define a single
ingress gateway `public-ingressgateway`. We add two additional ingress
gateways: `dataplane-ingressgateway` and `passthrough-ingressgateway`. The
data plane gateway is used to provide some level of isolation between the Big
Bang services and those added later.

Both the public-ingressgateway and dataplane-ingressgateway will do TLS
termination at the gateway. The passthrough gateway is used for traffic to
the keycloak service because keycloak insists on doing its own TLS termination.

> ⚠️ **The default k3d load balancer is unable to cope with multiple ingress gateways.** We recommend starting k3d with `--k3s-arg --disable=servicelb@server:*` and then running Metal LB on the cluster. This will allow you to run this package using all the gateways specified. For more details see [this](https://github.com/keunlee/k3d-metallb-starter-kit). This technique is used for the automated [unit test](../test/dco_core_package_test.go).

## References

* [Zarf](https://zarf.dev/docs)
* [Big Bang](https://docs-bigbang.dso.mil/latest/)

