# DCO Core

DCO Core is a Kubernetes application for Defensive Cyber Operations (DCO).
DCO aims to protect and defend computer networks and systems from cyber
threats and attacks. This project provides the core functionality required
to deploy and run the DCO suite.

> **⚠️ Warning**
>
> This is under active development and is not intended for public use yet.

## Features

- Enables deployment of the DCO suite in a Kubernetes cluster.
- Provides essential components and services for Defensive Cyber Operations.
- Facilitates airgapped deployment through the production of a Zarf package.

## Getting Started

### Prerequisites

- Kubernetes cluster.
- Access to the DCO Core source code repository and packages.
- Zarf version 0.27 or greater if building locally.

### Install from Zarf Package

```shell
zarf package deploy oci://ghcr.io/naps-dev/packages/dco-everything:${BRANCH-OR-TAG}-amd64 --components ${LIST-OF-COMMA-SEPARATED-COMPONENTS} --set ${OPTIONS} --confirm
```

### Building from Scratch

See [here](dco-everything/README.md)
