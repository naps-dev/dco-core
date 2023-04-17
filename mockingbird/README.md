# mockingbird

## VM OCI Image Creation

The GitLab pipeline uses `virt-v2v` and `virt-customize` to convert the Mockingbird VM then builds an OCI container. The converted VM is saved on S3 at `naps-dev-artifact/naps-dev/mockingbird/[branch or tag]` ([ex](https://s3.console.aws.amazon.com/s3/object/naps-dev-artifacts?region=us-east-1&prefix=naps-dev/mockingbird)) and the image uploaded to ECR at `ECR \ Repositories \ mockingbird \ [branch or tag]` ([ex](https://us-east-1.console.aws.amazon.com/ecr/repositories/private/765814079306/mockingbird?region=us-east-1)).

## Build Zarf Package

`zarf package create`

## Deploy Zarf Package

This package assumes the cluster is running KubeVirt and CDI and has the space and execution resources needed.

Ports 22, 80, and 8090 are exposed via a Kubernetes Service.

`zarf package deploy`
