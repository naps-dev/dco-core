# Mockingbird

## Quickstart

1. Build

    ```bash
    zarf package create --set GIT_REF=refs/heads/main --confirm
    ```

2. Deploy

    ```bash
    zarf package deploy --confirm
    ```

## Description

Mockingbird is a malware analysis tool.

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
* Cluster must have the Zarf initialization package deployed

> Note: the [dco-core](../dco-core/) Zarf package includes these prerequisites

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

## References

* [Zarf](https://zarf.dev/docs)
* [Mockingbird](https://docs-bigbang.dso.mil/latest/)

