# Arkime

## Quickstart

1. Create the package
    ```bash
    zarf package create --set GIT_REF=refs/heads/main --set IMAGE_TAG=v4.2.0-2
    ```

1. Deploy the package
    ```bash
    zarf package deploy
    ```

1. Access we web UI at [https://arkime.vp.bigbang.dev](https://arkime.vp.bigbang.dev)

## Description

[Arkime](https://arkime.com/) is a large-scale, open-source, indexed package capture and search tool. This directory packages a customized, internally managed Arkime image and Helm Chart for deployment into a broader DCO stack.

This package is a child to the [dco-core](../dco-core) Zarf package.

## Build

### Zarf Constants

| Name | Type | Purpose |
|--|--|--|
| `GIT_REF` | `string` | Provide the BRANCH (refs/heads/BRANCH) or TAG (refs/tags/TAG) git ref to identify the git reference to deploy |
| `IMAGE_TAG` | `string` | Arkime container image tag to package and deploy |

### Steps

Create the package
```bash
# For a specific branch and image tag
zarf package create --set GIT_REF=refs/heads/main --set IMAGE_TAG=v4.2.0-2

# For a specific git repository tag and matching image tag
zarf package create --set GIT_REF=refs/tags/v4.2.0-2 --set IMAGE_TAG=v4.2.0-2
```

## Installation

### Prerequisites (optional)

* Kubernetes cluster
* Cluster must have Flux installed
* Cluster must have the Zarf initilization package deployed
* Cluster must have the Big Bang package deployed

_Note: the [dco-core](../dco-core/) Zarf package includes these prerequisites and will deploy this package_

### Zarf Variables

| Name | Type | Purpose | Default |
|--|--|--|--|
| `CAPTURE_INTERFACE` | `string` | Interface the Arkime sensor will listen on | `cni0` |
| `DOMAIN` | `string` | Domain tied to Istio ingress and other parts of the Big Bang umbrella chart | `vp.bigbang.dev` |

### Steps

Deploy the package
```bash
zarf package deploy

# Optionally specify a different domain. The certificates packaged in kustomizations/bigbang/environment-bb-secret.yaml must match the provided domain
zarf package deploy --set DOMAIN=your.domain.here

# Optionally specify a different capture interface
zarf package deploy --set CAPTURE_INTERFACE=eni0
```

## Usage

### Credentials

Default development credentials for web login

| Property | Deafult Value |
|--|--|
| username | `localadmin` |
| password | `password` |

### Integrations

Arkime uses the [Dataplane ElasticSearch](../dataplane-ek/) database deployed by [dco-core](../dco-core/).

### Tests

Verify Arkime is accessible

```bash
LOADBALANCER_IP=$(kubectl get svc dataplane-ingressgateway -n istio-system --output=jsonpath="{.status.loadBalancer.ingress[0]['ip']}")

curl --resolve arkime-viewer.vp.bigbang.dev:443:"${LOADBALANCER_IP}" --fail-with-body https://arkime-viewer.vp.bigbang.dev
```

## References

* [Arkime](https://arkime.com/)
* [Upstream naps-dev Arkime repository](https://github.com/naps-dev/arkime)
