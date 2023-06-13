# dataplane-ek

## Quickstart

1. Create the package

    ```bash
    zarf package create
    ```

2. Deploy the package

    ```bash
    zarf package deploy
    ```

## Description

The `dataplane-ek` Zarf package is responsible for packaging and deploying [Elastic Search and Kibana](https://repo1.dso.mil/big-bang/product/packages/elasticsearch-kibana) from Big Bang for deployment into a broader DCO stack.

## Build

### Steps

1. Create the package

    ```bash
    zarf package create
    ```

## Installation

### Prerequisites (optional)

* Kubernetes cluster
* Cluster must have Flux installed
* Cluster must have the Zarf initialization package deployed
* Cluster must have the Big Bang package deployed

### Zarf Variables

| Name | Type | Purpose | Default |
|--|--|--|--|
| `DOMAIN` | `string` | Domain tied to Istio ingress and other parts of the Big Bang umbrella chart | `vp.bigbang.dev` |
| `KIBANA_COUNT` | `string` | The desired Kibana replica count | `1` |
| `ES_MASTER_COUNT` | `string` | The desired Elastic Search Master replica count | `1` |
| `ES_DATA_COUNT` | `string` | The desired Elastic Search Data replica count | `1` |

### Steps

Deploy the package

```bash
zarf package deploy

# Optionally specify a different domain. The certificates packaged in kustomizations/bigbang/environment-bb-secret.yaml must match the provided domain
zarf package deploy --set DOMAIN=your.domain.here

# Optionally specify a different kibana count
zarf package deploy --set KIBANA_COUNT=2

# Optionally specify a different elastic search master count
zarf package deploy --set ES_MASTER_COUNT=2

# Optionally specify a different elastic search data replica count
zarf package deploy --set ES_DATA_COUNT=2
```

### Tests

Verify Kibana is accessible

```bash
curl --resolve "dataplane-kibana.vp.bigbang.dev:443:<ip address>" --fail-with-body https://dataplane-kibana.vp.bigbang.dev
```

## References

* [Zarf](https://zarf.dev/docs)
* [Big Bang Elastic Search Kibana](https://repo1.dso.mil/big-bang/product/packages/elasticsearch-kibana)
