REQUIRED_BINS:=bash sudo curl go docker aws # jq needed for assume-role, but since that's not a required target, ignoring here.

ZARF_VER?=0.27.1
GOLANG_VER_MIN?=1.19.5
AWS_ROLE_SESSION_NAME?=arkime-ecr
AWS_REGION?=us-east-1
REF_TYPE?=branch
REF_NAME?=main
ZARF_PACKAGE_NAME?=zarf-package-dco-core-amd64.tar.zst
COMPONENT?=dco-core
DCO_REF_TYPE?=tag # In actions, defined as ${{ github.head_ref || github.ref_name }}. See README for notes.
DCO_REF_NAME?=main # In actions, defined as ${{ github.ref_type }}. See README for notes.
DCO_DIR?=dco-core 
ZARF_CONFIG?=$(shell pwd)/bigbang/zarf-config.yaml
export ZARF_CONFIG

DOCKER_BUILD ?= docker build
DOCKER_PUSH_IMAGE ?= docker push
IMAGE_NAME ?= ghcr.io/naps-dev/dco-core
DOCKER_RUN ?= docker run --rm -v $(PWD):/app/ -w /app --env-file ./.env $(IMAGE_NAME):latest
DOCKER_RUN_IT ?= docker run -it --rm -v $(PWD):/app/ -w /app --env-file ./.env --entrypoint /bin/bash $(IMAGE_NAME):latest

.DEFAULT_GOAL := help

.PHONY: help
help: ## List of targets with descriptions
	@echo "\n--------------------- Run [TARGET] [ARGS] or "make help" for more information ---------------------\n"
	@for MAKEFILENAME in $(MAKEFILE_LIST); do \
		grep -E '[a-zA-Z_-]+:.*?## .*$$' $$MAKEFILENAME  | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'; \
	done
	@echo "\n---------------------------------------------------------------------------------------------------\n"

.PHONY: build-image
build-image: ## Builds the docker image with all dependencies to produce Zarf packages
	$(DOCKER_BUILD) --build-arg ZARF_VER=$(ZARF_VER) -f ./Dockerfile  -t $(IMAGE_NAME):latest .

.PHONY: test-image
test-image:
	$(DOCKER_RUN_IT)

.PHONY: build-all
build-all: check-dependencies install-dependencies ecr-login zarf-registry1-login build-dco-package build-package ## Builds the dco-core and COMPONENT Zarf packages

.PHONY: build-all-docker
build-all-docker: ## Builds the dco-core and COMPONENT Zarf packages in a docker container
	$(DOCKER_RUN) build-all

.PHONY: test
test: check-dependencies install-dependencies run-tests ## Runs the go tests

.PHONY: check-dependencies
check-dependencies: ## Check for dependencies that would be easier to install using package manager for the specific distro
	@$(foreach bin,$(REQUIRED_BINS),\
		$(if $(shell command -v $(bin)),$(info $(bin) found.),$(error No $(bin) in PATH)))

.PHONY: install-dependencies
install-dependencies: ## Installs certain dependency packages - currently Zarf and k3d
# Zarf
ifeq (,$(shell command -v zarf))
	$(info "Zarf not installed. Installing...")
	make install-zarf
else 
ifeq (v$(ZARF_VER),$(shell zarf version))
	$(info "Zarf $(shell zarf version) already installed")
else
	$(info "Zarf v$(ZARF_VER) installed. Installing correct version of Zarf...")
	make clean-zarf
	make install-zarf
endif
endif

# Go, this simply checks version
ifneq ($(shell go version | awk '{print $$3}' | tr -d 'go'),$(shell printf "%s\n%s" "$(GOLANG_VER_MIN)" "$(shell go version | awk '{print $$3}' | tr -d 'go')" | sort -V | tail -n 1))
	$(error "Install a newer version of go, >=$(GOLANG_VER_MIN)...")
else
	$(info "Correct Go version, >=$(GOLANG_VER_MIN), installed")
endif

# k3d
# Note - during testing, it was noted that older version of k3d (v5.0.0) don't accept certain special characters in the cluster name, this version should work but version checking for k3d is probably needed
ifeq (,$(shell command -v k3d))
	$(info "k3d not installed. Installing...")
	curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v5.4.3 bash 
else
	$(info "k3d already installed")
endif

.PHONY: install-zarf
install-zarf: ## Install Zarf
	curl -L -o zarf_v$(ZARF_VER)_Linux_amd64 https://github.com/defenseunicorns/zarf/releases/download/v$(ZARF_VER)/zarf_v$(ZARF_VER)_Linux_amd64
	chmod +x ./zarf_v$(ZARF_VER)_Linux_amd64 && sudo mv ./zarf_v$(ZARF_VER)_Linux_amd64 /usr/local/bin/zarf
	curl -L -o zarf-init-amd64-v$(ZARF_VER).tar.zst https://github.com/defenseunicorns/zarf/releases/download/v$(ZARF_VER)/zarf-init-amd64-v$(ZARF_VER).tar.zst
	mkdir $(HOME)/.zarf-cache && mv ./zarf-init-amd64-v$(ZARF_VER).tar.zst $(HOME)/.zarf-cache/

.PHONY: install-go
install-go: ## Install Go
	curl -L -o go.tar.gz https://dl.google.com/go/go$(GOLANG_VER_MIN).linux-amd64.tar.gz && \
	tar -C /usr/local/ -xzf go.tar.gz && \
	rm go.tar.gz

.PHONY: install-aws
install-aws: ## Install AWS CLI
	curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
	unzip awscliv2.zip && \
	sudo ./aws/install

.PHONY: assume-role
assume-role: ## Assumes role user; not needed on EC2 instance with the role AWS_ECR_ROLE
	# Assume the role and output the result to a JSON file
	aws sts assume-role --role-arn $(AWS_ECR_ROLE) --role-session-name  $(AWS_ROLE_SESSION_NAME) > assume-role-output.json

	# Use jq to parse the JSON file and set environment variables
	export AWS_ACCESS_KEY_ID=$(jq -r .Credentials.AccessKeyId assume-role-output.json)
	export AWS_SECRET_ACCESS_KEY=$(jq -r .Credentials.SecretAccessKey assume-role-output.json)
	export AWS_SESSION_TOKEN=$(jq -r .Credentials.SessionToken assume-role-output.json)

.PHONY: ecr-login
ecr-login: ## Login to Amazon ECR
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

.PHONY: push-to-s3
push-to-s3: ## Pushes Zarf package to s3
	cd $(COMPONENT)
	aws s3 cp --no-progress "$(ZARF_PACKAGE_NAME)" "s3://$(AWS_BUCKET)/naps-dev/$(COMPONENT)/$(REF_NAME)/$(ZARF_PACKAGE_NAME)"

.PHONY: zarf-registry1-login
zarf-registry1-login: ## Zarf registry1 login
	@zarf tools registry login \
		-u "$(REGISTRY1_USERNAME)" \
		-p "$(REGISTRY1_PASSWORD)" \
		"registry1.dso.mil"

.PHONY: build-dco-package
build-dco-package: ## Builds dco-core Zarf package
	./build-package.sh $(DCO_DIR) $(DCO_REF_TYPE) $(DCO_REF_NAME)

.PHONY: build-package
build-package: ## Builds COMPONENT Zarf package
ifeq ($(COMPONENT),test)
	$(error TEST!!!)
endif
ifeq ($(COMPONENT),xsoar)
	# Login to registry, add license file
	@zarf tools registry login \
		-u "$(XSOAR_USERNAME)" \
		-p "$(XSOAR_PASSWORD)" \
		"xsoar-registry.pan.dev"
	@echo "$(XSOAR_LICENSE)" >> /tmp/demisto.lic
endif
ifeq ($(COMPONENT),polarity)
	# Add license file
	@echo "$(POLARITY_LICENSE)" >> /tmp/polarity.lic
endif
	./build-package.sh $(COMPONENT) $(REF_TYPE) $(REF_NAME) $(IMAGE_TAG)

.PHONY: run-tests
run-tests: ## Runs go tests
	$(info zarf_config=$(ZARF_CONFIG))
	go get -t ./...
	cd ./test && go test -timeout 40m

.PHONY: clean
clean: clean-zarf clean-packages ## Clean files

.PHONY: clean-zarf
clean-zarf: ## Cleans Zarf install
	sudo rm /usr/local/bin/zarf
	rm -rf $(HOME)/.zarf-cache

.PHONY: clean-go
clean-go: ## Cleans go install 
	rm -rf /usr/local/go

.PHONY: clean-packages
clean-packages: ## Cleans built packages
	rm $(COMPONENT)/$(ZARF_PACKAGE_NAME)
	rm $(DCO_DIR)/zarf-package-dco-core-amd64.tar.zst
ifeq ($(COMPONENT),xsoar)
	rm /tmp/demisto.lic
endif
ifeq ($(COMPONENT),polarity)
	rm /tmp/polarity.lic
endif
