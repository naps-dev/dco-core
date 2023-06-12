# README
### How to build without GitHub Actions
The purpose of this branch is to facilitate the build of Zarf packages in an environment that does use GitHub actions. The `Makefile` and `build-package.sh` have been added.

The `Dockerfile` generates a docker image with the necessary components. The Makefile is invoked from the image as follows:

### Sample make commands
```
make all REF_NAME="v4.2.0-2" REF_TYPE="tag" ZARF_PACKAGE="zarf-package-arkime-amd64.tar.zst" IMAGE_TAG="v4.2.0-2" COMPONENT="arkime"
```
Requires the following to be environment variables
- AWS_ECR_ROLE
- REGISTRY1_USERNAME
- REGISTRY1_PASSWORD