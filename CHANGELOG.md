# Changelog

* [CHANGELOG](./CHANGELOG.md)
* [LICENSE](./LICENSE)
* [README](./README.md)
* [CONTRIBUTING](./CONTRIBUTING.md)

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
