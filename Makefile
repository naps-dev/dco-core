.PHONY: all clean checkout install-zarf assume-role ecr-login registry-login build-dco-package build-package install-go run-tests

ZARF_VER="0.26.2"
GOLANG_VER="1.19.5"
PATH="$PATH:/usr/local/go/bin"
AWS_ROLE_SESSION_NAME="arkime-ecr"
AWS_REGION="us-east-1"
AWS_ACCOUNT_ID="765814079306"

REF_TYPE?="branch"
REF_NAME?="main"
ZARF_PACKAGE?="zarf-package-dco-core-amd64.tar.zst"
COMPONENT?="dco-core"
DCO_REF_TYPE?="tag"
DCO_REF_NAME?="main"
DCO_DIR?="dco-core"

all: install-zarf ecr-login registry-login build-dco-package build-package install-go run-tests

install-zarf: 
	curl -L -o zarf_v$(ZARF_VER)_Linux_amd64 https://github.com/defenseunicorns/zarf/releases/download/v$(ZARF_VER)/zarf_v$(ZARF_VER)_Linux_amd64
	chmod +x ./zarf_v$(ZARF_VER)_Linux_amd64 && mv ./zarf_v$(ZARF_VER)_Linux_amd64 /usr/local/bin/zarf
	curl -L -o zarf-init-amd64-v$(ZARF_VER).tar.zst https://github.com/defenseunicorns/zarf/releases/download/v$(ZARF_VER)/zarf-init-amd64-v$(ZARF_VER).tar.zst
	mkdir $(HOME)/.zarf-cache && mv ./zarf-init-amd64-v$(ZARF_VER).tar.zst $(HOME)/.zarf-cache/

assume-role: ## This isn't needed when on a runner with the role AWS_ECR_ROLE
	# Assume the role and output the result to a JSON file
	aws sts assume-role --role-arn $(AWS_ECR_ROLE) --role-session-name  $(AWS_ROLE_SESSION_NAME) > assume-role-output.json

	# Use jq to parse the JSON file and set environment variables
	export AWS_ACCESS_KEY_ID=$(jq -r .Credentials.AccessKeyId assume-role-output.json)
	export AWS_SECRET_ACCESS_KEY=$(jq -r .Credentials.SecretAccessKey assume-role-output.json)
	export AWS_SESSION_TOKEN=$(jq -r .Credentials.SessionToken assume-role-output.json)

ecr-login:
	# Login to ECR
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

registry-login:
	zarf tools registry login \
		-u "$(REGISTRY1_USERNAME)" \
		-p "$(REGISTRY1_PASSWORD)" \
		"registry1.dso.mil"

build-dco-package:
	./build-package.sh $(DCO_DIR) $(DCO_REF_TYPE) $(DCO_REF_NAME) ""

build-package:
	./build-package.sh $(COMPONENT) $(REF_TYPE) $(REF_NAME) $(IMAGE_TAG)

install-go:
	curl -L -o go$(GOLANG_VER).linux-amd64.tar.gz https://dl.google.com/go/go$(GOLANG_VER).linux-amd64.tar.gz && \
	    tar -C /usr/local/ -xzf go$(GOLANG_VER).linux-amd64.tar.gz && \
	    rm go$(GOLANG_VER).linux-amd64.tar.gz

run-tests:
	go get -t ./...
	cd ./test && go test -timeout 40m

clean:
	rm /usr/local/bin/zarf
	rm -rf $(HOME)/.zarf-cache
	rm $(COMPONENT)/$(ZARF_PACKAGE)
	rm $(DCO_DIR)/zarf-package-dco-core-amd64.tar.zst
	rm /usr/local/go