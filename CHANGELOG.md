# Changelog

* [CHANGELOG](./CHANGELOG.md)
* [LICENSE](./LICENSE)
* [README](./README.md)
* [CONTRIBUTING](./CONTRIBUTING.md)

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.1] - 2021-07-14
### Fixed
* A recent version of mailcow changed the type of a property from a string to an int. This
  release adds support for newer mailcow versions returning an int while preserving functionality
  on older mailcow versions that return a string.

## [1.3.0] - 2021-05-18
### Added
* New `scheme` option to allow API requests via http. (Thank you [maximbaz](https://github.com/maximbaz))

## [1.2.0] - 2020-09-06
### Added
* New rspamd metrics. Requires an up-to-date mailcow version as it uses a brand new API endpoint.

## [1.1.3] - 2020-09-06
### Fixed
* Errors in single providers will no longer translate to errors in the whole exporter.
  Instead, the new `mailcow_exporter_success` and `mailcow_api_success` metrics will then
  be set to 0. This is done to make the exporter provide metrics, even if parts of it fail.

## [1.1.2] - 2020-09-05
### Fixed
* non-200 API responses (such as authorization errors) no longer throw an obscure JSON
  marshalling error, but a more helpful message. In general, API error messages contain
  a lot more information now, which makes finding issues in ones setup easier.

## [1.1.1] - 2020-09-05
### Fixed
* `mailcow_container_start` accidentally reported a static value

## [1.1.0] - 2020-09-05
### Added
* Meta information about the mailcow API requests
* Container information
* Help texts
