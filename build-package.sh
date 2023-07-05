#!/bin/bash
# Arguments: [1 = COMPONENT, 2 = REF_TYPE, 3 = REF_NAME, 4 (optional) = IMAGE_TAG] 

cd $1

# cleanup and init a tmp dir
rm -rf big-tmp
mkdir big-tmp

args=()
if [ "$2" = "branch" ]; then
    args+=(--set GIT_REF=refs/heads/$3)
fi
if [ "$2" = "tag" ]; then
    args+=(--set GIT_REF=refs/tags/$3)
fi
if [ $4 ]; then
    args+=(--set IMAGE_TAG="$4")
fi
args+=(--confirm)
args+=(--tmpdir $PWD/big-tmp)
args+=(--skip-sbom)

# build the zarf package
zarf package create "${args[@]}"

# cleanup tmp dir
rm -rf ./big-tmp