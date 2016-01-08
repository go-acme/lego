# Changelog

## [Unreleased]

### Added:
- CLI: The `--exclude` or `-x` switch. To exclude a challenge from being solved.
- CLI: The `--http` switch. To set the listen address and port of HTTP based challenges. Supports `host:port` and `:port` for any interface.
- CLI: The `--tls` switch. To set the listen address and port of TLS based challenges. Supports `host:port` and `:port` for any interface.
- lib: ExcludeChallenges function. Pass an array of challenge identifiers to exclude them from solving.
- lib: SetHTTPAddress function. Pass a port to set the listen port for HTTP based challenges.
- lib: SetTLSAddress function. Pass a port to set the listen port of TLS based challenges.
- lib: acme.UserAgent variable. Use this to customize the user agent on all requests sent by lego.

### Changed:
- lib: NewClient does no longer accept the optPort parameter
- lib: ObtainCertificate now returns a SAN certificate if you pass more then one domain.
- lib: GetOCSPForCert now returns the parsed OCSP response instead of just the status.

### Removed:
- CLI: The `--port` switch was removed.
- lib: RenewCertificate does no longer offer to also revoke your old certificate.

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
