# Makefile README
## Motivation
Some future state of this effort may exist without access to Github actions. In this case, the `Makefile` can be used to facilitate the build and test of the Zarf packages in this repository.

### How To
This Makefile should only be executed on Linux, AMD64 architecture; most of the Zarf packages are only compatible with that architecture.

Most of the variables needed are pre-set in the Makefile, for individual packages the following values should be set
- REF_NAME 
- REF_TYPE 
- ZARF_PACKAGE
- COMPONENT 
- IMAGE_TAG (optional)

Additionally, the Iron Bank login credentials are required as so: 
- REGISTRY1_USERNAME
- REGISTRY1_PASSWORD

AWS access credentials will need to be passed if not running on instance with proper priviledges:
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_DEFAULT_REGION

This makefile assumes that users might be running in different environments with packages possibly not pre-installed. Two recipes attempt to handle this:
- check-dependencies simply identifies if necessary dependencies are available in the environment. These are dependencies that are probably best managed by a package manager. With go and aws in particular, the out of the box installation wasn't compatible with alpine, so created install targets for both but omitting from `all`.
- install-dependencies checks if a dependency is there, sometimes if it's the right version, and then installs if not. The makefile will install zarf and k3d. It will check versioning for go.

Note - "assume_role" and "push" are currently unused and have not been fully tested.

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

```
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
```
$(git rev-parse --abbrev-ref HEAD)
```