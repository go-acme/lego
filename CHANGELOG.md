# Changelog

## [Unreleased]

### Added:
- CLI: The `--exclude` or `-x` switch. To exclude a challenge from being solved.
- CLI: The `--httpPort`. To set the listen port of HTTP based challenges.
- CLI: The `--tlsPort`. To set the listen port of TLS based challenges.
- lib: ExcludeChallenges function. Pass an array of challenge identifiers to exclude them from solving.
- lib: SetHTTPPort function. Pass a port to set the listen port for HTTP based challenges.
- lib: SetTLSPort function. Pass a port to set the listen port of TLS based challenges.

### Changed:
- lib: NewClient does no longer accept the optPort parameter
- lib: ObtainCertificate now returns a SAN certificate if you pass more then one domain.

### Removed:
- CLI: The `--port` switch was removed.

### Fixed:
- CLI: Fix logic using the `--days` parameter for renew

## [0.1.1] - 2015-12-18

### Added:
- CLI: Added a way to automate renewal through a cronjob using the --days parameter to renew

### Changed:
- lib: Improved log output on challenge failures.

### Fixed:
- CLI: The short parameter for domains would not get accepted
- CLI: The cli did not return proper exit codes on error library errors.
- lib: RenewCertificate did not properly renew SAN certificates.

### Security
- lib: Fix possible DOS on GetOCSPForCert

## [0.1.0] - 2015-12-03
- Initial release

[Unreleased]: https://github.com/xenolf/lego/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/xenolf/lego/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/xenolf/lego/tree/v0.1.0
