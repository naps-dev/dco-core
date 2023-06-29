# Makefile README
## Motivation
Some future state of this effort may exist without access to Github actions. In this case, the `Makefile` can be used to facilitate the build and test of the Zarf packages in this repository.

## How To
Help is available for the supported `make` targets by running `make` in the root project directory.   

```shell
make

--------------------- Run [TARGET] [ARGS] or make help for more information ---------------------

help                           List of targets with descriptions
build-image                    Builds the docker image with all dependencies to produce Zarf packages
build-all                      Builds the dco-core and COMPONENT Zarf packages
build-all-docker               Builds the dco-core and COMPONENT Zarf packages in a docker container (preferred)
test                           Runs the go tests
check-dependencies             Check for dependencies that would be easier to install using package manager for the specific distro
install-dependencies           Installs certain dependency packages - currently Zarf and k3d
install-zarf                   Install Zarf
install-go                     Install Go
install-aws                    Install AWS CLI
assume-role                    Assumes role user; not needed on EC2 instance with the role AWS_ECR_ROLE
ecr-login                      Login to Amazon ECR
zarf-registry1-login           Zarf registry1 login
build-dco-package              Builds dco-core Zarf package
build-package                  Builds COMPONENT Zarf package
run-tests                      Runs go tests
clean                          Clean files
clean-zarf                     Cleans Zarf install
clean-go                       Cleans go install 
clean-packages                 Cleans built packages

---------------------------------------------------------------------------------------------------
```
### Build Packages

There are two targets to help build the Zarf packages:
1) build-all
2) build-all-docker

If you would like to build the Zarf packages using a containerized environment to host dependencies (as opposed to needing them installed locally), the image needs to be build first (future: load to ghcr)
``` shell
make build-image
```
Create a .env file at the root directory to hold your credentials and variables, for example:
```yaml
REGISTRY1_USERNAME=[Username]
REGISTRY1_PASSWORD=[Password]
AWS_ACCOUNT_ID=[Account ID]
AWS_ACCESS_KEY_ID=[ID]
AWS_SECRET_ACCESS_KEY=[Key]
AWS_DEFAULT_REGION=us-east-1

COMPONENT=arkime
REF_NAME=v4.2.0-2
REF_TYPE=tag
ZARF_PACKAGE_NAME=zarf-package-arkime-amd64.tar.zst
IMAGE_TAG=v4.2.0-2
```

Then run your build-all-docker target, for example
```
make build-all-docker
```

When building an individual package, the following values should be set
- REF_NAME 
- REF_TYPE 
- ZARF_PACKAGE_NAME
- COMPONENT 
- IMAGE_TAG (optional)

Iron Bank login credentials are required as so: 
- REGISTRY1_USERNAME
- REGISTRY1_PASSWORD

For ECR login, AWS_ACCOUNT_ID is required

AWS access credentials will need to be passed if not running on instance with proper priviledges:
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_DEFAULT_REGION

Finally, ZARF_CONFIG will need to be updated if it's not located in ./bigbang/zarf-config.yaml

The `test` target is currently separate from `build-all` and can only be run with the dependencies installed locally after the packages are built. (COMPONENT and REF_NAME required) 

Note - "assume_role" and "push-to-s3" are currently unused and have not been fully tested.

#### XSOAR
If this component is being built, it also requires:
- XSOAR_USERNAME (for xsoar registry)
- XSOAR_PASSWORD (for xsoar registry)
- XSOAR_LICENSE

#### Polarity
If this component is being built, it also requires:
- POLARITY_LICENSE

### Substitutions for github variables
Several github variables are used in the Actions. While these are not currently replicated in the Makefile, the analogous definitions might look as follows:

#### `github.ref_type`
This can be extracted by `$(ref_type)` using the following bash script function

``` shell
function ref_type {
    if git show-ref --verify --quiet refs/heads/$(git rev-parse --abbrev-ref HEAD); then
        echo "branch"
    elif git show-ref --verify --quiet refs/tags/$(git describe --tags --exact-match); then
        echo "tag"
    else
        echo "other"
    fi
}
```

#### `github.head_ref || github.ref_name`
The following can be used to get the current git branch, which seems analogous to these variables (?)
``` shell
$(git rev-parse --abbrev-ref HEAD)
```