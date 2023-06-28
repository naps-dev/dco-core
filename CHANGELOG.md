# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added toleration and node selector to Suricata DaemonSet for controlling hardware placement
- Makefile to replicate Github Actions

## [v2.3.0-1] - 2023-06-20

### Changed
- DUBBD

## [v2.0.1-3] - 2023-06-20
### Added
- this changelog

### Updated
- Upgraded Suricata to v7.0.0-0
- split dataplane-ek
- Update LetsEncrypt certs
  
## [v2.0.1-2] - 2023-06-20
### Added
- this changelog

### Updated
- KubeVirt and CDI sub-packages in the `kubevirt/` directory
- update mixmode sensor
- upped pvc size for mockingbird
- Split metallb into separate Zarf component


## [v2.0.1-1] - 2023-05-05

 
[unreleased]: https://github.com/naps-dev/dco-core/compare/v2.3.0-1...HEAD
[v2.3.0-1]: https://github.com/naps-dev/dco-core/compare/v2.0.1-3...v2.3.0-1
[v2.0.1-3]: https://github.com/naps-dev/dco-core/compare/v2.0.1-2...v2.0.1-3
[v2.0.1-2]: https://github.com/naps-dev/dco-core/compare/v2.0.1-1...v2.0.1-2
[v2.0.1-1]: https://github.com/naps-dev/dco-core/releases/tag/v2.0.1-1
