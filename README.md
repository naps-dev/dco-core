# Defensive Cyber Operations Foundation

## BigBang Upgrade Process
Tentative instructions for upgrading the version of Bigbang included in the zarf package. This only addresses how to update the zarf package -- additional consideration will be needed for updating any existing deployments.

Instructions:
* Update Bigbang repository URL reference tags in:
  * `./kustomizations/bigbang/kustomization.yaml`
  * `./zarf.yaml`
* View the [Bigbang release notes](https://repo1.dso.mil/platform-one/big-bang/bigbang/-/releases) for the target version ([example -- 1.46.0](https://repo1.dso.mil/platform-one/big-bang/bigbang/-/releases/1.46.0))
* Adjust the version tags in `components.big-bang-core-standard-assets.repos`  according to the version shown in the `BB Version` column of the "Packages" table in the release notes
* Download the `package-images.yaml` from the release notes
* Remove unused packages from the yaml
* Use `yq '.package-image-list.*.images' ~/Downloads/package-images.yaml | yq 'unique'` to filter the list
* Update the component `images` according to the `yq` output^
* test test test