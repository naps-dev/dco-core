# MixMode Server 8.14

The MixMode server or 'brain'.
Please see [docs/instructions.txt](docs/instructions.txt) for setup instructions from the vendor.

## VM OCI Image Creation

The GitLab pipeline uses `virt-v2v` and `virt-customize` to convert the MixMode VM then builds an OCI container. The converted VM is saved on S3 at `naps-dev-artifact/naps-dev/mixmode/[branch or tag]` ([ex](https://s3.console.aws.amazon.com/s3/object/naps-dev-artifacts?region=us-east-1&prefix=naps-dev/mixmode/feature/mm/mixmode-8.11-sda.qcow2)) and the image uploaded to ECR at `ECR \ Repositories \ mixmode \ [branch or tag]` ([ex](https://us-east-1.console.aws.amazon.com/ecr/repositories/private/765814079306/mixmode?region=us-east-1)).

## Build Zarf Package

`zarf package create`

## Deploy Zarf Package

This package assumes the cluster is running KubeVirt and CDI and has the space and execution resources needed.

`zarf package deploy`

Please note: the [Virtual Machine ](chart/templates/virtualmachine.yaml) performs a number of 

## Log in to the web application

https://mixmode-enterprise.vp-stg.bigbang.dev

Username: provider
Password: *see below*

#### Get the Password

You must have the mixmode private key. It is available on S3 in `naps-dev-artifacs > mixmode-key > mixmode`

```
kubectl virt ssh -i ~/.ssh/mixmode --username centos -n mixmode mixmode-enterprise -c "grep LDAP /opt/mixmode/env | cut -d'=' -f2 | tr -d '\"'"
```