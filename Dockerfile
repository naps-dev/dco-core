# Base image for docker
FROM docker:24.0.2

ARG ZARF_VER

# Install necessary tools
RUN apk update \
    && apk upgrade \
    && apk add --no-cache bash curl sudo make go jq git aws-cli 

# Install k3d
RUN wget -q -O - https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v5.5.1 bash

# Install zarf
RUN curl -L -o zarf_v${ZARF_VER}_Linux_amd64 https://github.com/defenseunicorns/zarf/releases/download/v${ZARF_VER}/zarf_v${ZARF_VER}_Linux_amd64 && \
	chmod +x ./zarf_v${ZARF_VER}_Linux_amd64 && sudo mv ./zarf_v${ZARF_VER}_Linux_amd64 /usr/local/bin/zarf && \
	curl -L -o zarf-init-amd64-v${ZARF_VER}.tar.zst https://github.com/defenseunicorns/zarf/releases/download/v${ZARF_VER}/zarf-init-amd64-v${ZARF_VER}.tar.zst  && \
	mkdir ${HOME}/.zarf-cache && mv ./zarf-init-amd64-v${ZARF_VER}.tar.zst ${HOME}/.zarf-cache/

# Define the entrypoint as the make command
ENTRYPOINT ["make"]