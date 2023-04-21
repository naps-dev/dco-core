# Polarity

This repo contains the polarity helm chart, zarf yaml, and github actions to build zartifact.

Before running the zartifact, create a Polarity license secret:

```
kubectl create secret generic polarity-server-lic \
  --from-file=polarity.lic=<path-to-polarity.lic> \
  --namespace=polarity
```
Note - "polarity-server-lic" string can be configured in the chart's values.yaml