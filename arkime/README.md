# ARKIME
This repo contains a helm chart for deploying the open source network traffic analysis tool [Arkime](https://arkime.com/) alongside [Bigbang](https://repo1.dso.mil/platform-one/big-bang/bigbang). Additionally, this repo manages a [zarf](https://zarf.dev) package for performing offline installs.

## Build Assumptions
* Access to the image repo
* Required assets (see Dockerfile) copied to `assets/` prior to `docker build`

## Deployment Assumptions
* A pre-existing Bigbang deployment from [dco-foundation](https://github.com/naps-dev/dco-foundation)
