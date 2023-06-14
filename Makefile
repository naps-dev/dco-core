.PHONY: all clean check-dependencies install-dependencies install-zarf install-go install-aws assume-role ecr-login push-to-s3 registry-login build-dco-package build-package run-tests clean clean-zarf clean-go clean-packages

REQUIRED_BINS:=bash sudo gcc curl go docker aws # jq needed for assume-role, but since that's not a required target, ignoring here.

ZARF_VER=0.27.0
GOLANG_VER_MIN=1.19.5
AWS_ROLE_SESSION_NAME=arkime-ecr
AWS_REGION=us-east-1
AWS_ACCOUNT_ID=765814079306

REF_TYPE?=branch
REF_NAME?=main
ZARF_PACKAGE?=zarf-package-dco-core-amd64.tar.zst
COMPONENT?=dco-core
DCO_REF_TYPE?=tag # In actions, defined as ${{ github.head_ref || github.ref_name }}. See README for notes.
DCO_REF_NAME?=main # In actions, defined as ${{ github.ref_type }}. See README for notes.
DCO_DIR?=dco-core

all: check-dependencies install-dependencies ecr-login registry-login build-dco-package build-package run-tests

## Check for and Install dependencies
check-dependencies: ## Check for dependencies that would be easier to install using package manager for the specific distro
	@$(foreach bin,$(REQUIRED_BINS),\
		$(if $(shell which $(bin)),$(info $(bin) found.),$(error No $(bin) in PATH)))

install-dependencies:
# Zarf
ifeq (,$(shell which zarf))
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
ifeq (,$(shell which k3d))
	$(info "k3d not installed. Installing...")
	curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v5.4.3 bash 
else
	$(info "k3d already installed")
endif

install-zarf: 
	curl -L -o zarf_v$(ZARF_VER)_Linux_amd64 https://github.com/defenseunicorns/zarf/releases/download/v$(ZARF_VER)/zarf_v$(ZARF_VER)_Linux_amd64
	chmod +x ./zarf_v$(ZARF_VER)_Linux_amd64 && mv ./zarf_v$(ZARF_VER)_Linux_amd64 /usr/local/bin/zarf
	curl -L -o zarf-init-amd64-v$(ZARF_VER).tar.zst https://github.com/defenseunicorns/zarf/releases/download/v$(ZARF_VER)/zarf-init-amd64-v$(ZARF_VER).tar.zst
	mkdir $(HOME)/.zarf-cache && mv ./zarf-init-amd64-v$(ZARF_VER).tar.zst $(HOME)/.zarf-cache/

install-go: 
	curl -L -o go.tar.gz https://dl.google.com/go/go$(GOLANG_VER_MIN).linux-amd64.tar.gz && \
	tar -C /usr/local/ -xzf go.tar.gz && \
	rm go.tar.gz

install-aws:
	curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
	unzip awscliv2.zip && \
	sudo ./aws/install

## AWS stuff
assume-role: ## This isn't needed when on a runner with the role AWS_ECR_ROLE
	# Assume the role and output the result to a JSON file
	aws sts assume-role --role-arn $(AWS_ECR_ROLE) --role-session-name  $(AWS_ROLE_SESSION_NAME) > assume-role-output.json

	# Use jq to parse the JSON file and set environment variables
	export AWS_ACCESS_KEY_ID=$(jq -r .Credentials.AccessKeyId assume-role-output.json)
	export AWS_SECRET_ACCESS_KEY=$(jq -r .Credentials.SecretAccessKey assume-role-output.json)
	export AWS_SESSION_TOKEN=$(jq -r .Credentials.SessionToken assume-role-output.json)

ecr-login: ## Login to ECR
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

push-to-s3: ## Pushes Zarf package to s3, currently omitted in all
	cd $(COMPONENT)
	aws s3 cp --no-progress "$(ZARF_PACKAGE)" "s3://$(AWS_BUCKET)/naps-dev/$(COMPONENT)/$(REF_NAME)/$(ZARF_PACKAGE)"

## Build packages and test
registry-login:
	@zarf tools registry login \
		-u "$(REGISTRY1_USERNAME)" \
		-p "$(REGISTRY1_PASSWORD)" \
		"registry1.dso.mil"

build-dco-package:
	./build-package.sh $(DCO_DIR) $(DCO_REF_TYPE) $(DCO_REF_NAME) ""

build-package:
ifeq ($(COMPONENT),"xsoar")
	# Login to registry, add license file
	@zarf tools registry login \
		-u "$(XSOAR_USERNAME)" \
		-p "$(XSOAR_PASSWORD)" \
		"xsoar-registry.pan.dev"
	@echo "$(XSOAR_LICENSE)" >> /tmp/demisto.lic
endif
ifeq ($(COMPONENT),"polarity")
	# Add license file
	@echo "$(POLARITY_LICENSE)" >> /tmp/polarity.lic
endif
	./build-package.sh $(COMPONENT) $(REF_TYPE) $(REF_NAME) $(IMAGE_TAG)

run-tests:
	go get -t ./...
	cd ./test && go test -timeout 40m

## Clean files
clean: clean-zarf clean-packages 

clean-zarf:
	rm /usr/local/bin/zarf
	rm -rf $(HOME)/.zarf-cache

clean-go:
	rm -rf /usr/local/go

clean-packages:
	rm $(COMPONENT)/$(ZARF_PACKAGE)
	rm $(DCO_DIR)/zarf-package-dco-core-amd64.tar.zst
