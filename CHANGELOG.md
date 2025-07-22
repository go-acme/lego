# Changelog

## [v4.25.1](https://github.com/go-acme/lego/releases/tag/v4.25.1) (2025-07-21)

### Fixed

- **[cli]** fix: wrong CLI flag type

## [v4.25.0](https://github.com/go-acme/lego/releases/tag/v4.25.0) (2025-07-21)

The binary size of this release is about ~50% smaller compared to previous releases.

This will also reduce the module cache usage by 320 MB (this will only affect users of lego as a library or who build lego themselves).

### Added

- **[dnsprovider]** Add DNS provider for ZoneEdit
- **[cli]** Add an option to define dynamically the renew date
- **[lib,cli]** Add an option to disable common name in CSR

### Changed
- 
- **[dnsprovider]** vinyldns: add an option to add quotes around the TXT record value
- **[dnsprovider]** ionos: increase default propagation timeout

### Fixed

- **[cli]** fix: enforce domain into renewal command

## [v4.24.0](https://github.com/go-acme/lego/releases/tag/v4.24.0) (2025-07-07)

### Added

- **[dnsprovider]** Add DNS provider for Azion
- **[dnsprovider]** Add DNS provider for DynDnsFree.de
- **[dnsprovider]** Add DNS provider for ConoHa v3
- **[dnsprovider]** Add DNS provider for RU Center
- **[dnsprovider]** gcloud: add service account impersonation

### Changed

- **[dnsprovider]** pdns: improve error messages
- **[dnsprovider]** cloudflare: add quotation marks to TXT record
- **[dnsprovider]** googledomains: provider deprecation
- **[dnsprovider]** mijnhost: improve record filter

### Fixed

- **[dnsprovider]** exoscale: fix find record
- **[dnsprovider]** nicmanager: fix mode env var name and value
- **[lib,cli]** Check order identifiers difference between client and server

## [v4.23.1](https://github.com/go-acme/lego/releases/tag/v4.23.1) (2025-04-16)

Due to an error related to Snapcraft, some artifacts of the v4.23.0 release have not been published.

This release contains the same things as v4.23.0. 

## [v4.23.0](https://github.com/go-acme/lego/releases/tag/v4.23.0) (2025-04-16)

### Added

- **[dnsprovider]** Add DNS provider for Active24
- **[dnsprovider]** Add DNS provider for BookMyName
- **[dnsprovider]** Add DNS provider for Axelname
- **[dnsprovider]** Add DNS provider for Baidu Cloud
- **[dnsprovider]** Add DNS provider for Metaregistrar
- **[dnsprovider]** Add DNS provider for F5 XC
- **[dnsprovider]** Add INFOBLOX_CA_CERTIFICATE option
- **[dnsprovider]** route53: adds option to use private zone
- **[dnsprovider]** edgedns: add account switch key option
- **[dnsprovider]** infoblox: update API client to v2
- **[lib,cli]** Add delay option for TLSALPN challenge

### Changed

- **[dnsprovider]** designate: speed up API requests by using filters
- **[dnsprovider]** cloudflare: make base URL configurable
- **[dnsprovider]** websupport: migrate to API v2
- **[dnsprovider]** dnssimple: use GetZone

### Fixed

- **[ari]** Fix retry on `alreadyReplaced` error
- **[cli,log]** Fix malformed log messages
- **[cli]** Kill hook when the command is stuck
- **[dnsprovider]** pdns: fix TXT record cleanup for wildcard domains
- **[dnsprovider]** allinkl: remove `ReturnInfo`

## [v4.22.2](https://github.com/go-acme/lego/releases/tag/v4.22.2) (2025-02-17)

### Fixed

- **[dnsprovider]** acme-dns: use new registred account

## [v4.22.1](https://github.com/go-acme/lego/releases/tag/v4.22.1) (2025-02-17)

### Fixed

- **[dnsprovider]** acme-dns: continue the process when the CNAME is handled by the storage

### Added

## [v4.22.0](https://github.com/go-acme/lego/releases/tag/v4.22.0) (2025-02-17)

### Added

- **[cli]** Add `--private-key` flag to set the private key.
- **[cli]** Add `LEGO_DEBUG_ACME_HTTP_CLIENT` environment variable to debug the calls to the ACME server.
- **[cli]** Add `LEGO_EMAIL` environment variable for specifying email.
- **[cli]** Add `--hook-timeout` flag to run and renew commands.
- **[dnsprovider]** Add DNS provider for myaddr.{tools,dev,io}
- **[dnsprovider]** Add DNS provider for Spaceship
- **[dnsprovider]** acme-dns: add HTTP storage
- **[lib,cli,httpprovider]** Add `--http.delay` option for HTTP challenge.
- **[lib,cli,profiles]** Add support for Profiles Extension.
- **[lib]** Add an option to set CSR email addresses

### Changed

- **[lib]** rewrite status management
- **[dnsprovider]** docs: improve units and default values

### Removed

- **[dnsprovider]** netcup: remove TTL option

### Fixed

- **[cli,log]** remove extra debug logs

## [v4.21.0](https://github.com/go-acme/lego/releases/tag/v4.21.0) (2024-12-20)

### Added

- **[dnsprovider]** Add DNS provider for Rainyun/雨云
- **[dnsprovider]** Add DNS provider for West.cn/西部数码
- **[dnsprovider]** Add DNS provider for ManageEngine CloudDNS
- **[cli]** feat: add --force-cert-domains flag to renew

### Fixed

- **[cli]** create client only when needed
- **[cli]** clone the transport with tls-skip-verify
- **[cli]** use retryable client for ACME server calls
- **[dnsprovider]** bunny: fix zone detection
- **[dnsprovider]** inwx: delete only the TXT record related to the DNS challenge
- **[dnsprovider]** infomaniak: increase default propagation timeout
- **[dnsprovider]** dnsmadeeasy: use default transport
- **[dnsprovider]** netcup: increase default propagation values
- **[dnsprovider]** otc: use default transport

## [v4.20.4](https://github.com/go-acme/lego/releases/tag/v4.20.4) (2024-11-21)

Publish the Snap to the Snapcraft stable channel.

## [v4.20.3](https://github.com/go-acme/lego/releases/tag/v4.20.3) (2024-11-21)

### Fixed

- **[dnsprovider]** technitium: fix status code handling
- **[dnsprovider]** directadmin: fix timeout configuration
- **[httpprovider]** fix: HTTP server IPv6 matching

## [v4.20.2](https://github.com/go-acme/lego/releases/tag/v4.20.2) (2024-11-11)

### Added

- **[dnsprovider]** Add DNS provider for Technitium
- **[dnsprovider]** Add DNS provider for Regfish
- **[dnsprovider]** Add DNS provider for Timeweb Cloud
- **[dnsprovider]** Add DNS provider for Volcano Engine
- **[dnsprovider]** Add DNS provider for Core-Networks
- **[dnsprovider]** rfc2136: add support for tsig-keygen generated file
- **[cli]** Add option to skip the TLS verification of the ACME server
- Add documentation for env var only options

### Changed

- **[cli,ari]** Attempt to check ARI unless explicitly disabled
- **[dnsprovider]** Improve propagation check error messages
- **[dnsprovider]** cloudxns: provider deprecation
- **[dnsprovider]** brandit: provider deprecation

### Fixed

- **[dnsprovider]** regru: update authentication method
- **[dnsprovider]** selectelv2: fix non-ASCII domain
- **[dnsprovider]** limacity: fix error message
- **[dnsprovider]** volcengine: set API information within the default configuration
- **[log]** Parse printf verbs in log line output

## v4.20.1 (2024-11-11)

Cancelled due to CI failure.

## v4.20.0 (2024-11-11)

Cancelled due to CI failure.

## [v4.19.2](https://github.com/go-acme/lego/releases/tag/v4.19.2) (2024-10-06)

### Fixed

- **[lib]** go1.22 compatibility

## [v4.19.1](https://github.com/go-acme/lego/releases/tag/v4.19.1) (2024-10-06)

### Fixed

- **[dnsprovider]** selectelv2: use baseURL from configuration
- **[dnsprovider]** epik: add User-Agent

## [v4.19.0](https://github.com/go-acme/lego/releases/tag/v4.19.0) (2024-10-03)

### Added

- **[dnsprovider]** Add DNS provider for HuaweiCloud
- **[dnsprovider]** Add DNS provider for SelfHost.(de|eu)
- **[lib,cli,dnsprovider]** Add `dns.propagation-rns` option
- **[cli,dnsprovider]** Add `dns.propagation-wait` flag
- **[lib,dnsprovider]** Add `PropagationWait` function

### Changed

- **[dnsprovider]** ionos: follow CNAME
- **[lib,dnsprovider]** Reducing the lock strength of the soa cache entry
- **[lib,cli,dnsprovider]** Deprecation of `dns.disable-cp`, replaced by `dns.propagation-disable-ans`.

### Fixed

- **[dnsprovider]** Use UTC instead of GMT when possible
- **[dnsprovider]** namesilo: restrict CleanUp
- **[dnsprovider]** godaddy: fix cleanup

## [v4.18.0](https://github.com/go-acme/lego/releases/tag/v4.18.0) (2024-08-30)

### Added

- **[dnsprovider]** Add DNS provider for mijn.host
- **[dnsprovider]** Add DNS provider for Lima-City
- **[dnsprovider]** Add DNS provider for DirectAdmin
- **[dnsprovider]** Add DNS provider for Mittwald
- **[lib,cli]** feat: add option to handle the overall request limit
- **[lib]** feat: expose certificates pool creation

### Changed

- **[cli]** feat: add LEGO_ISSUER_CERT_PATH to run hook
- **[dnsprovider]** bluecat: skip deploy
- **[dnsprovider]** ovh: allow to use ovh.conf file
- **[dnsprovider]** designate: allow manually overwriting DNS zone

### Fixed

- **[ari]** fix: avoid Int63n panic in ShouldRenewAt()

## [v4.17.4](https://github.com/go-acme/lego/releases/tag/v4.17.4) (2024-06-12)

### Fixed

- **[dnsprovider]** Update dependencies

## [v4.17.3](https://github.com/go-acme/lego/releases/tag/v4.17.3) (2024-05-28)

### Added

- **[dnsprovider]** Add DNS provider for Selectel v2
- **[dnsprovider]** route53: adds option to not wait for changes
- **[dnsprovider]** ovh: add OAuth2 authentication
- **[dnsprovider]** azuredns: use TenantID also for cli authentication
- **[dnsprovider]** godaddy: documentation about new API limitations
- **[cli]** feat: add LEGO_ISSUER_CERT_PATH to hook

### Changed

- **[dnsprovider]** dode: update API URL
- **[dnsprovider]** exec: stream command output
- **[dnsprovider]** oracle: update API client
- **[dnsprovider]** azuredns: servicediscovery for zones
- **[dnsprovider]** scaleway: add alternative env var names
- **[dnsprovider]** exoscale: simplify record creation
- **[dnsprovider]** httpnet: add provider to NewDNSChallengeProviderByName
- **[cli]** feat: fills LEGO_CERT_PFX_PATH and LEGO_CERT_PEM_PATH only when needed
- **[lib,ari]** feat: renewal retry after value

### Fixed

- **[dnsprovider]** pdns: reconstruct zone URLs to enable non-root folder API endpoints
- **[dnsprovider]** alidns: fix link to API documentation

## v4.17.2 (2024-05-28)

Canceled due to a release failure related to Snapcraft.

The Snapcraft release are disabled for now.

## v4.17.1 (2024-05-28)

Canceled due to a release failure related to oci-go-sdk.

The module `github.com/oracle/oci-go-sdk/v65` uses `github.com/gofrs/flock` but flock doesn't support some platform (like Solaris):
- https://github.com/gofrs/flock/issues/60

Due to that we will remove the Solaris build.

## v4.17.0 (2024-05-28)

Canceled due to a release failure related to Snapcraft.

## [v4.16.1](https://github.com/go-acme/lego/releases/tag/v4.16.1) (2024-03-10)

### Fixed

- **[cli,ari]** fix: don't generate ARI cert ID if ARI is not enable

## [v4.16.0](https://github.com/go-acme/lego/releases/tag/v4.16.0) (2024-03-09)

### Added

- **[dnsprovider]** Add DNS provider for Shellrent
- **[dnsprovider]** Add DNS provider for Mail-in-a-Box
- **[dnsprovider]** Add DNS provider for CPanel and WHM

### Changed

- **[lib,ari]** Implement 'replaces' field in newOrder and draft-ietf-acme-ari-03 CertID changes
- **[log]** feat: improve errors and logs related to DNS call
- **[lib]** update to go-jose/go-jose/v4 v4.0.1

### Fixed

- **[dnsprovider]** nifcloud: fix bug in case of same auth zone
- **[dnsprovider]** bunny: Support delegated subdomains
- **[dnsprovider]** easydns: fix zone detection
- **[dnsprovider]** ns1: fix record creation

## [v4.15.0](https://github.com/go-acme/lego/releases/tag/v4.15.0) (2024-01-28)

### Added

- **[dnsprovider]** Add DNS provider for http.net
- **[dnsprovider]** Add DNS provider for Webnames

### Changed

- **[cli]** Add environment variable for specifying alternate directory URL
- **[cli]** Add format option for PFX encoding
- **[lib]** Support simplified issuance for very long domain names at Let's Encrypt
- **[lib]** Update CertID format as per draft-ietf-acme-ari-02
- **[dnsprovider]** azuredns: allow OIDC authentication
- **[dnsprovider]** azuredns: provide the ability to select authentication methods
- **[dnsprovider]** efficientip: add insecure skip verify option
- **[dnsprovider]** gandiv5: add Personal Access Token support
- **[dnsprovider]** gcloud: support GCE_ZONE_ID to bypass zone list
- **[dnsprovider]** liquidweb: add LWAPI_ prefix for env vars
- **[dnsprovider]** liquidweb: detect zone automatically
- **[dnsprovider]** pdns: optional custom API version
- **[dnsprovider]** regru: client certificate support
- **[dnsprovider]** regru: HTTP method changed to POST
- **[dnsprovider]** scaleway: add cname support

### Fixed

- **[dnsprovider]** cloudru: change default URLs
- **[dnsprovider]** constellix: follow rate limiting headers
- **[dnsprovider]** desec: increase default propagation interval
- **[dnsprovider]** gandiv5: Add "Bearer" prefix to the auth header
- **[dnsprovider]** inwx: improve sleep calculation
- **[dnsprovider]** inwx: wait before generating new TOTP TANs
- **[dnsprovider]** ionos: fix DNS record removal
- **[dnsprovider]** ipv64: remove unused option
- **[dnsprovider]** nifcloud: fix API requests
- **[dnsprovider]** otc: sequential challenge

## [v4.14.1](https://github.com/go-acme/lego/releases/tag/v4.14.1) (2023-09-20)

### Fixed

- **[dnsprovider]** bunny: fix zone detection
- **[dnsprovider]** bunny: use NRDCG fork
- **[dnsprovider]** ovh: update client to v1.4.2

## v4.14.1 (2023-09-19)

Cancelled due to CI failure.

## [v4.14.0](https://github.com/go-acme/lego/releases/tag/v4.14.0) (2023-08-20)

### Added

- **[dnsprovider]** Add DNS provider for Yandex 360
- **[dnsprovider]** Add DNS provider for cloud.ru
- **[httpprovider]** Adding S3 support for HTTP domain validation

### Changed

- **[cli]** Allow to set EAB kid and hmac via environment variables
- **[dnsprovider]** Migrate to aws-sdk-go-v2 (lightsail, route53)

### Fixed

- **[dnsprovider]** nearlyfreespeech: fix authentication
- **[dnsprovider]** pdns: fix notify
- **[dnsprovider]** route53: avoid unexpected records deletion

## [v4.13.3](https://github.com/go-acme/lego/releases/tag/v4.13.3) (2023-07-25)

### Fixed

- **[dnsprovider]** azuredns: fix configuration from env vars
- **[dnsprovider]** gcore: change API domain

## [v4.13.2](https://github.com/go-acme/lego/releases/tag/v4.13.2) (2023-07-21)

### Fixed

- **[dnsprovider]** servercow: fix regression

## [v4.13.1](https://github.com/go-acme/lego/releases/tag/v4.13.1) (2023-07-20)

### Added

- **[dnsprovider]** Add DNS provider for IPv64
- **[dnsprovider]** Add DNS provider for Metaname
- **[dnsprovider]** Add DNS provider for RcodeZero
- **[dnsprovider]** Add DNS provider for Efficient IP
- **[dnsprovider]** azure: new implementation based on the new API client
- **[lib]** Experimental option to force DNS queries to use TCP

### Changed

- **[dnsprovider]** cloudflare: update api client to v0.70.0

### Fixed

- **[dnsprovider,cname]** fix: ensure case-insensitive comparison of CNAME records
- **[cli]** fix: list command
- **[lib]** fix: ARI explanationURL

## v4.13.0 (2023-07-20)

Cancelled due to a CI issue (no space left on device).

## [v4.12.2](https://github.com/go-acme/lego/releases/tag/v4.12.2) (2023-06-19)

### Fixed

- **[dnsprovider]** dnsmadeeasy: fix DeleteRecord
- **[lib]** fix: read status code from response

## [v4.12.1](https://github.com/go-acme/lego/releases/tag/v4.12.1) (2023-06-06)

### Fixed

- **[dnsprovider]** pdns: fix record value

## [v4.12.0](https://github.com/go-acme/lego/releases/tag/v4.12.0) (2023-05-28)

### Added

- **[lib,cli]** Initial ACME Renewal Info (ARI) Implementation
- **[dnsprovider]** Add DNS provider for Derak Cloud
- **[dnsprovider]** route53: pass ExternalID property to STS:AssumeRole API operation
- **[lib,cli]** Support custom duration for certificate

### Changed

- **[dnsprovider]** Refactor DNS provider and client implementations

### Fixed

- **[dnsprovider]** autodns: fixes wrong zone in api call if CNAME is used
- **[cli]** fix: archive only domain-related files on revoke

## [v4.11.0](https://github.com/go-acme/lego/releases/tag/v4.11.0) (2023-05-02)

### Added

- **[lib]** Support for certificate with raw IP SAN (RFC8738)
- **[dnsprovider]** Add Brandit.com as DNS provider
- **[dnsprovider]** Add DNS provider for Bunny
- **[dnsprovider]** Add DNS provider for Nodion
- **[dnsprovider]** Add Google Domains as DNS provider
- **[dnsprovider]** Add DNS provider for Plesk

### Changed

- **[cli]** feat: add LEGO_CERT_PEM_PATH and LEGO_CERT_PFX_PATH to run hook
- **[lib,cli]** feat: add RSA 3072
- **[dnsprovider]** gcloud: update google APIs to latest version
- **[lib,dnsprovider,cname]** chore: replace GetRecord by GetChallengeInfo

### Fixed

- **[dnsprovider]** rimuhosting: fix API base URL

## [v4.10.2](https://github.com/go-acme/lego/releases/tag/v4.10.2) (2023-02-26)

Fix Docker image builds.

## [v4.10.1](https://github.com/go-acme/lego/releases/tag/v4.10.1) (2023-02-25)

### Fixed

- **[dnsprovider,cname]** acmedns: fix CNAME support
- **[dnsprovider]** dynu: fix subdomain support

## [v4.10.0](https://github.com/go-acme/lego/releases/tag/v4.10.0) (2023-02-10)

### Added

- **[dnsprovider]** Add DNS provider for dnsHome.de
- **[dnsprovider]** Add DNS provider for Liara
- **[dnsprovider]** Add DNS provider for UltraDNS
- **[dnsprovider]** Add DNS provider for Websupport

### Changed

- **[dnsprovider]** ibmcloud: add support for subdomains
- **[dnsprovider]** infomaniak: CNAME support
- **[dnsprovider]** namesilo: add cleanup before add a DNS record
- **[dnsprovider]** route53: Allow static credentials to be supplied
- **[dnsprovider]** tencentcloud: support punycode domain

### Fixed

- **[dnsprovider]** alidns: filter on record type
- **[dnsprovider]** arvancloud: replace arvancloud.com by arvancloud.ir
- **[dnsprovider]** hetzner: improve zone ID detection
- **[dnsprovider]** luadns: removed dot suffix from authzone while searching for zone
- **[dnsprovider]** pdns: fix usage of notify only when zone kind is Master or Slave
- **[dnsprovider]** return an error when extracting record name

## [v4.9.1](https://github.com/go-acme/lego/releases/tag/v4.9.1) (2022-11-25)

### Changed

- **[lib,cname]** cname: add log about CNAME entries
- **[dnsprovider]** regru: improve error handling

### Fixed

- **[dnsprovider,cname]** fix CNAME support for multiple DNS providers
- **[dnsprovider,cname]** duckdns: fix CNAME support
- **[dnsprovider,cname]** oraclecloud: use fqdn to resolve zone
- **[dnsprovider]** hurricane: fix CNAME support
- **[lib,cname]** cname: stop trying to traverse cname if none have been found

## [v4.9.0](https://github.com/go-acme/lego/releases/tag/v4.9.0) (2022-10-03)

### Added

- **[dnsprovider]** Add DNS provider for CIVO
- **[dnsprovider]** Add DNS provider for VK Cloud
- **[dnsprovider]** Add DNS provider for YandexCloud
- **[dnsprovider]** digitalocean: configurable base URL
- **[dnsprovider]** loopia: add configurable API endpoint
- **[dnsprovider]** pdns: notify secondary servers after updates

### Changed

- **[dnsprovider]** allinkl: removed deprecated sha1 hashing
- **[dnsprovider]** auroradns: update authentification
- **[dnsprovider]** dnspod: deprecated. Use Tencent Cloud instead.
- **[dnsprovider]** exoscale: migrate to API v2 endpoints
- **[dnsprovider]** gcloud: update golang.org/x/oauth2
- **[dnsprovider]** lightsail: cleanup
- **[dnsprovider]** sakuracloud: update api client library
- **[cname]** take out CNAME support from experimental features
- **[lib,cname]** add recursive CNAME lookup support
- **[lib]** Remove embedded issuer certificates from issued certificate if bundle is false

### Fixed

- **[dnsprovider]** luadns: fix cname support
- **[dnsprovider]** njalla: fix record id unmarshal error
- **[dnsprovider]** tencentcloud: fix subdomain error

## [v4.8.0](https://github.com/go-acme/lego/releases/tag/v4.8.0) (2022-06-30)

### Added

- **[dnsprovider]** Add DNS provider for Variomedia
- **[dnsprovider]** Add NearlyFreeSpeech DNS Provider
- **[cli]** Add a --user-agent flag to lego-cli

### Changed

- new logo
- **[cli]** feat: sleep at renewal
- **[cli]** cli/renew: skip random sleep if stdout is a terminal
- **[dnsprovider]** hetzner: set min TTL to 60s
- **[docs]** refactoring and cleanup

## [v4.7.0](https://github.com/go-acme/lego/releases/tag/v4.7.0) (2022-05-27)

### Added

- **[dnsprovider]** Add DNS provider for iwantmyname
- **[dnsprovider]** Add DNS Provider for IIJ DNS Platform Service
- **[dnsprovider]** Add DNS provider for Vercel
- **[dnsprovider]** route53: add assume role ARN
- **[dnsprovider]** dnsimple: add debug option
- **[cli]** feat: add `LEGO_CERT_PEM_PATH` and `LEGO_CERT_PFX_PATH`

### Changed

- **[dnsprovider]** gcore: change dns api url
- **[dnsprovider]** bluecat: rewrite provider implementation

### Fixed

- **[dnsprovider]** rfc2136: fix TSIG secret
- **[dnsprovider]** tencentcloud: fix InvalidParameter.DomainInvalid error when using DNS challenges
- **[lib]** fix: panic in certcrypto.ParsePEMPrivateKey

## [v4.6.0](https://github.com/go-acme/lego/releases/tag/v4.6.0) (2022-01-18)

### Added

- **[dnsprovider]** Add DNS provider for UKFast SafeDNS
- **[dnsprovider]** Add DNS Provider for Tencent Cloud
- **[dnsprovider]** azure: add support for Azure Private Zone DNS
- **[dnsprovider]** exec: add sequence interval
- **[cli]** Add a `--pfx`, and `--pfx.pas`s option to generate a PKCS#12 (`.pfx`) file.
- **[lib]** Extended support of cert pool (`LEGO_CA_CERTIFICATES` and `LEGO_CA_SYSTEM_CERT_POOL`)
- **[lib,httpprovider]** added uds capability to http challenge server

### Changed

- **[lib]** Extend validity of TLS-ALPN-01 certificates to 365 days
- **[lib,cli]** Allows defining the reason for the certificate revocation

### Fixed

- **[dnsprovider]** mythicbeasts: fix token expiration
- **[dnsprovider]** rackspace: change zone ID to string

## [v4.5.3](https://github.com/go-acme/lego/releases/tag/v4.5.3) (2021-09-06)

### Fixed

- **[lib,cli]** fix: missing preferred chain param for renew request

## [v4.5.2](https://github.com/go-acme/lego/releases/tag/v4.5.2) (2021-09-01)

### Added

- **[dnsprovider]** Add DNS provider for all-inkl
- **[dnsprovider]** Add DNS provider for Epik
- **[dnsprovider]** Add DNS provider for freemyip.com
- **[dnsprovider]** Add DNS provider for g-core labs
- **[dnsprovider]** Add DNS provider for hosttech
- **[dnsprovider]** Add DNS Provider for IBM Cloud (SoftLayer)
- **[dnsprovider]** Add DNS provider for Internet.bs
- **[dnsprovider]** Add DNS provider for nicmanager

### Changed

- **[dnsprovider]** alidns: support ECS instance RAM role
- **[dnsprovider]** alidns: support sts token credential
- **[dnsprovider]** azure: zone name as environment variable
- **[dnsprovider]** ovh: follow cname
- **[lib,cli]** Add AlwaysDeactivateAuthorizations flag to ObtainRequest

### Fixed

- **[dnsprovider]** infomaniak: fix subzone support
- **[dnsprovider]** edgedns: fix Present and CleanUp logic
- **[dnsprovider]** lightsail: wrong Region env var name
- **[lib]** lib: fix backoff in SolverManager
- **[lib]** lib: use permanent error instead of context cancellation
- **[dnsprovider]** desec: bump to v0.6.0

## v4.5.1 (2021-09-01)

Cancelled due to a CI issue, replaced by v4.5.2.

## v4.5.0 (2021-09-30)

Cancelled due to a CI issue, replaced by v4.5.2.

## [v4.4.0](https://github.com/go-acme/lego/releases/tag/v4.4.0) (2021-06-08)

### Added

- **[dnsprovider]** Add DNS provider for Infoblox
- **[dnsprovider]** Add DNS provider for Porkbun
- **[dnsprovider]** Add DNS provider for Simply.com
- **[dnsprovider]** Add DNS provider for Sonic
- **[dnsprovider]** Add DNS provider for VinylDNS
- **[dnsprovider]** Add DNS provider for wedos

### Changed

- **[cli]** log: Use stderr instead of stdout.
- **[dnsprovider]** hostingde: autodetection of the zone name.
- **[dnsprovider]** scaleway: use official SDK
- **[dnsprovider]** powerdns: several improvements
- **[lib]** lib: improve wait.For returns.

### Fixed

- **[dnsprovider]** hurricane: add API rate limiter.
- **[dnsprovider]** hurricane: only treat first word of response body as response code
- **[dnsprovider]** exoscale: fix DNS provider debugging
- **[dnsprovider]** wedos: fix api call parameters
- **[dnsprovider]** nifcloud: Get zone info from dns01.FindZoneByFqdn
- **[cli,lib]** csr: Support the type `NEW CERTIFICATE REQUEST`

## [v4.3.1](https://github.com/go-acme/lego/releases/tag/v4.3.1) (2021-03-12)

### Fixed

- **[dnsprovider]** exoscale: fix dependency version.

## [v4.3.0](https://github.com/go-acme/lego/releases/tag/v4.3.0) (2021-03-10)

### Added

- **[dnsprovider]** Add DNS provider for Njalla
- **[dnsprovider]** Add DNS provider for Domeneshop
- **[dnsprovider]** Add DNS provider for Hurricane Electric
- **[dnsprovider]** designate: support for Openstack Application Credentials
- **[dnsprovider]** edgedns: support for .edgerc file

### Changed

- **[dnsprovider]** infomaniak: Make error message more meaningful
- **[dnsprovider]** cloudns: Improve reliability
- **[dnsprovider]** rfc2163: Removed support for MD5 algorithm. The default algorithm is now SHA1.

### Fixed

- **[dnsprovider]** desec: fix error with default TTL
- **[dnsprovider]** mythicbeasts: implement `ProviderTimeout`
- **[dnsprovider]** dnspod: improve search accuracy when a domain have more than 100 records
- **[lib]** Increase HTTP client timeouts
- **[lib]** preferred chain only match root name

## [v4.2.0](https://github.com/go-acme/lego/releases/tag/v4.2.0) (2021-01-24)

### Added

- **[dnsprovider]** Add DNS provider for Loopia
- **[dnsprovider]** Add DNS provider for Ionos.

### Changed

- **[dnsprovider]** acme-dns: update cpu/goacmedns to v0.1.1.
- **[dnsprovider]** inwx: Increase propagation timeout to 360s to improve robustness
- **[dnsprovider]** vultr: Update to govultr v2 API
- **[dnsprovider]** pdns: get exact zone instead of all zones

### Fixed

- **[dnsprovider]** vult, dnspod: fix default HTTP timeout.
- **[dnsprovider]** pdns: URL request creation.
- **[lib]** errors: Fix instance not being printed

## [v4.1.3](https://github.com/go-acme/lego/releases/tag/v4.1.3) (2020-11-25)

### Fixed

- **[dnsprovider]** azure: fix error handling.

## [v4.1.2](https://github.com/go-acme/lego/releases/tag/v4.1.2) (2020-11-21)

### Fixed

- **[lib]** fix: preferred chain support.

## [v4.1.1](https://github.com/go-acme/lego/releases/tag/v4.1.1) (2020-11-19)

### Fixed

- **[dnsprovider]** otc: select correct zone if multiple returned
- **[dnsprovider]** azure: fix target must be a non-nil pointer

## [v4.1.0](https://github.com/go-acme/lego/releases/tag/v4.1.0) (2020-11-06)

### Added

- **[dnsprovider]** Add DNS provider for Infomaniak
- **[dnsprovider]** joker: add support for SVC API
- **[dnsprovider]** gcloud: add an option to allow the use of private zones

### Changed

- **[dnsprovider]** rfc2136: ensure TSIG algorithm is fully qualified
- **[dnsprovider]** designate: Deprecate OS_TENANT_NAME as required field

### Fixed

- **[lib]** acme/api: use postAsGet instead of post for AccountService.Get
- **[lib]** fix: use http.Header.Set method instead of Add.

## [v4.0.1](https://github.com/go-acme/lego/releases/tag/v4.0.1) (2020-09-03)

### Fixed

- **[dnsprovider]** exoscale: change dependency version.

## [v4.0.0](https://github.com/go-acme/lego/releases/tag/v4.0.0) (2020-09-02)

### Added

- **[cli], [lib]** Support "alternate" certificate links for selecting different signing Chains

### Changed

- **[cli]** Replaces `ec384` by `ec256` as default key-type
- **[lib]** Changes `ObtainForCSR` method signature

### Removed

- **[dnsprovider]** Replaces FastDNS by EdgeDNS
- **[dnsprovider]** Removes old Linode provider
- **[lib]** Removes `AddPreCheck` function

## [v3.9.0](https://github.com/go-acme/lego/releases/tag/v3.9.0) (2020-09-01)

### Added

- **[dnsprovider]** Add Akamai Edgedns. Deprecate FastDNS
- **[dnsprovider]** Add DNS provider for HyperOne

### Changed

- **[dnsprovider]** designate: add support for Openstack clouds.yaml
- **[dnsprovider]** azure: allow selecting environments
- **[dnsprovider]** desec: applies API rate limits.

### Fixed

- **[dnsprovider]** namesilo: fix cleanup.

## [v3.8.0](https://github.com/go-acme/lego/releases/tag/v3.8.0) (2020-07-02)

### Added

- **[cli]** cli: add hook on the run command.
- **[dnsprovider]** inwx: Two-Factor-Authentication
- **[dnsprovider]** Add DNS provider for ArvanCloud

### Changed

- **[dnsprovider]** vultr: bumping govultr version
- **[dnsprovider]** desec: improve error logs.
- **[lib]** Ensures the return of a location during account updates
- **[dnsprovider]** route53: Document all AWS credential environment variables

### Fixed

- **[dnsprovider]** stackpath: fix subdomain support.
- **[dnsprovider]** arvandcloud: fix record name.
- **[dnsprovider]** fix: multi-va.
- **[dnsprovider]** constellix: fix search records API call.
- **[dnsprovider]** hetzner: fix record name.
- **[lib]** Registrar.ResolveAccountByKey: Fix malformed request

## [v3.7.0](https://github.com/go-acme/lego/releases/tag/v3.7.0) (2020-05-11)

### Added

- **[dnsprovider]** Add DNS provider for Netlify.
- **[dnsprovider]** Add DNS provider for deSEC.io
- **[dnsprovider]** Add DNS provider for LuaDNS
- **[dnsprovider]** Adding Hetzner DNS provider
- **[dnsprovider]** Add DNS provider for Mythic beasts DNSv2
- **[dnsprovider]** Add DNS provider for Yandex.

### Changed

- **[dnsprovider]** Upgrade DNSimple client to 0.60.0
- **[dnsprovider]** update aws sdk

### Fixed

- **[dnsprovider]** autodns: removes TXT records during CleanUp.
- **[dnsprovider]** Fix exoscale HTTP timeout
- **[cli]** fix: renew path information.
- **[cli]** Fix account storage location warning message

## [v3.6.0](https://github.com/go-acme/lego/releases/tag/v3.6.0) (2020-04-24)

### Added

- **[dnsprovider]** Add DNS provider for CloudDNS.
- **[dnsprovider]** alicloud: add support for domain with punycode
- **[dnsprovider]** cloudns: Add subuser support
- **[cli]** Information about renewed certificates are now passed to the renew hook

### Changed

- **[dnsprovider]** acmedns: Update cpu/goacmedns v0.0.1 -&gt; v0.0.2
- **[dnsprovider]** alicloud: update sdk dependency version to v1.61.112
- **[dnsprovider]** azure: Allow for the use of MSI
- **[dnsprovider]** constellix: improve challenge.
- **[dnsprovider]** godaddy: allow parallel solve.
- **[dnsprovider]** namedotcom: get the actual registered domain, so we can remove just that from the hostname to be created
- **[dnsprovider]** transip: updated the client to v6

### Fixed

- **[dnsprovider]** ns1: fix missing domain in log 
- **[dnsprovider]** rimuhosting: use HTTP client from config.

## [v3.5.0](https://github.com/go-acme/lego/releases/tag/v3.5.0) (2020-03-15)

### Added

- **[dnsprovider]** Add DNS provider for Dynu.
- **[dnsprovider]** Add DNS provider for reg.ru
- **[dnsprovider]** Add DNS provider for Zonomi and RimuHosting.
- **[cli]** Building binaries for arm 6 and 5
- **[cli]** Uses CGO_ENABLED=0
- **[cli]** Multi-arch Docker image.
- **[cli]** Adds `--name` flag to list command.

### Changed

- **[lib]** lib: Improve cleanup log messages.
- **[lib]** Wrap errors.

### Fixed

- **[dnsprovider]** azure: pass AZURE_CLIENT_SECRET_FILE to autorest.Authorizer
- **[dnsprovider]** gcloud: fixes issues when used with GKE Workload Identity
- **[dnsprovider]** oraclecloud: fix subdomain support

## [v3.4.0](https://github.com/go-acme/lego/releases/tag/v3.4.0) (2020-02-25)

### Added

- **[dnsprovider]** Add DNS provider for Constellix
- **[dnsprovider]** Add DNS provider for Servercow.
- **[dnsprovider]** Add DNS provider for Scaleway
- **[cli]** Add "LEGO_PATH" environment variable

### Changed

- **[dnsprovider]** route53: allow custom client to be provided
- **[dnsprovider]** namecheap: allow external domains
- **[dnsprovider]** namecheap: add sandbox support.
- **[dnsprovider]** ovh: Improve provider documentation
- **[dnsprovider]** route53: Improve provider documentation

### Fixed

- **[dnsprovider]** zoneee: fix subdomains.
- **[dnsprovider]** designate: Don't clean up managed records like SOA and NS
- **[dnsprovider]** dnspod: update lib.
- **[lib]** crypto: Treat CommonName as optional
- **[lib]** chore: update cenkalti/backoff to v4.

## [v3.3.0](https://github.com/go-acme/lego/releases/tag/v3.3.0) (2020-01-08)

### Added

- **[dnsprovider]** Add DNS provider for Checkdomain
- **[lib]** Add support to update account

### Changed

- **[dnsprovider]** gcloud: Auto-detection of the project ID.
- **[lib]** Successfully parse private key PEM blocks

### Fixed

- **[dnsprovider]** Update dnspod, because of API breaking changes.

## [v3.2.0](https://github.com/go-acme/lego/releases/tag/v3.2.0) (2019-11-10)

### Added

- **[dnsprovider]** Add support for autodns

### Changed

- **[dnsprovider]** httpreq: Allow use environment vars from a `_FILE` file
- **[lib]** Don't deactivate valid authorizations
- **[lib]** Expose more SOA fields found by dns01.FindZoneByFqdn

### Fixed

- **[dnsprovider]** use token as unique ID.

## [v3.1.0](https://github.com/go-acme/lego/releases/tag/v3.1.0) (2019-10-07)

### Added

- **[dnsprovider]** Add DNS provider for Liquid Web
- **[dnsprovider]** cloudflare: add support for API tokens
- **[cli]** feat: ease operation behind proxy servers

### Changed

- **[dnsprovider]** cloudflare: update client
- **[dnsprovider]** linodev4: propagation timeout configuration.

### Fixed

- **[dnsprovider]** ovh: fix int overflow.
- **[dnsprovider]** bindman: fix client version.

## [v3.0.2](https://github.com/go-acme/lego/releases/tag/v3.0.2) (2019-08-15)

### Fixed

- Invalid pseudo version (related to Cloudflare client).

## [v3.0.1](https://github.com/go-acme/lego/releases/tag/v3.0.1) (2019-08-14)

There was a problem when creating the tag v3.0.1, this tag has been invalidated.

## [v3.0.0](https://github.com/go-acme/lego/releases/tag/v3.0.0) (2019-08-05)

### Changed

- migrate to go module (new import github.com/go-acme/lego/v3/)
- update DNS clients

## [v2.7.2](https://github.com/go-acme/lego/releases/tag/v2.7.2) (2019-07-30)

### Fixed

- **[dnsprovider]** vultr: quote TXT record

## [v2.7.1](https://github.com/go-acme/lego/releases/tag/v2.7.1) (2019-07-22)

### Fixed

- **[dnsprovider]** vultr: invalid record type.

## [v2.7.0](https://github.com/go-acme/lego/releases/tag/v2.7.0) (2019-07-17)

### Added

- **[dnsprovider]** Add DNS provider for namesilo
- **[dnsprovider]** Add DNS provider for versio.nl

### Changed

- **[dnsprovider]** Update DNS providers libs.
- **[dnsprovider]** joker: support username and password.
- **[dnsprovider]** Vultr: Switch to official client

### Fixed

- **[dnsprovider]** otc: Prevent sending empty body.

## [v2.6.0](https://github.com/go-acme/lego/releases/tag/v2.6.0) (2019-05-27)

### Added

- **[dnsprovider]** Add support for Joker.com DMAPI
- **[dnsprovider]** Add support for Bindman DNS provider
- **[dnsprovider]** Add support for EasyDNS
- **[lib]** Get an existing certificate by URL

### Changed

- **[dnsprovider]** digitalocean: LEGO_EXPERIMENTAL_CNAME_SUPPORT support
- **[dnsprovider]** gcloud: Use fqdn to get zone Present/CleanUp
- **[dnsprovider]** exec: serial behavior
- **[dnsprovider]** manual: serial behavior.
- **[dnsprovider]** Strip newlines when reading environment variables from `_FILE` suffixed files.

### Fixed

- **[cli]** fix: cli disable-cp option.
- **[dnsprovider]** gcloud: fix zone visibility.

## [v2.5.0](https://github.com/go-acme/lego/releases/tag/v2.5.0) (2019-04-17)

### Added

- **[cli]** Adds renew hook
- **[dnsprovider]** Adds 'Since' to DNS providers documentation

### Changed

- **[dnsprovider]** gcloud: use public DNS zones
- **[dnsprovider]** route53: enhance documentation.

### Fixed

- **[dnsprovider]** cloudns: fix TTL and status validation
- **[dnsprovider]** sakuracloud: supports concurrent update
- **[dnsprovider]** Disable authz when solve fail.
- Add tzdata to the Docker image.

## [v2.4.0](https://github.com/go-acme/lego/releases/tag/v2.4.0) (2019-03-25)

- Migrate from xenolf/lego to go-acme/lego.

### Added

- **[dnsprovider]** Add DNS Provider for Domain Offensive (do.de)
- **[dnsprovider]** Adds information about '_FILE' suffix.

### Fixed

- **[cli,dnsprovider]** Add 'manual' provider to the output of dnshelp
- **[dnsprovider]** hostingde: Use provided ZoneName instead of domain
- **[dnsprovider]** pdns: fix wildcard with SANs

## [v2.3.0](https://github.com/go-acme/lego/releases/tag/v2.3.0) (2019-03-11)

### Added

- **[dnsprovider]** Add DNS Provider for ClouDNS.net
- **[dnsprovider]** Add DNS Provider for Oracle Cloud

### Changed

- **[cli]** Adds log when no renewal.
- **[dnsprovider,lib]** Add a mechanism to wrap a PreCheckFunc
- **[dnsprovider]** oraclecloud: better way to get private key.
- **[dnsprovider]** exoscale: update library

### Fixed

- **[dnsprovider]** OVH: Refresh zone after deleting challenge record
- **[dnsprovider]** oraclecloud: ttl config and timeout 
- **[dnsprovider]** hostingde: fix client fails if customer has no access to dns-groups
- **[dnsprovider]** vscale: getting sub-domain
- **[dnsprovider]** selectel: getting sub-domain
- **[dnsprovider]** vscale: fix TXT records clean up
- **[dnsprovider]** selectel: fix TXT records clean up

## [v2.2.0](https://github.com/go-acme/lego/releases/tag/v2.2.0) (2019-02-08)

### Added

- **[dnsprovider]** Add support for Openstack Designate as a DNS provider
- **[dnsprovider]** gcloud: Option to specify gcloud service account json by env as string
- **[experimental feature]** Resolve CNAME when creating dns-01 challenge. To enable: set `LEGO_EXPERIMENTAL_CNAME_SUPPORT` to `true`.
 
### Changed

- **[cli]** Applies Let’s Encrypt’s recommendation about renew. The option `--days` of the command `renew` has a new default value (`30`)
- **[lib]** Uses a jittered exponential backoff

### Fixed

- **[cli]** CLI and key type.
- **[dnsprovider]** httpreq: Endpoint with path.
- **[dnsprovider]** fastdns: Do not overwrite existing TXT records
- Log wildcard domain correctly in validation

## [v2.1.0](https://github.com/go-acme/lego/releases/tag/v2.1.0) (2019-01-24)

### Added

- **[dnsprovider]** Add support for zone.ee as a DNS provider.

### Changed

- **[dnsprovider]** nifcloud: Change DNS base url.
- **[dnsprovider]** gcloud: More detailed information about Google Cloud DNS.

### Fixed

- **[lib]** fix: OCSP, set HTTP client.
- **[dnsprovider]** alicloud: fix pagination.
- **[dnsprovider]** namecheap: fix panic.

## [v2.0.0](https://github.com/go-acme/lego/releases/tag/v2.0.0) (2019-01-09)

### Added

- **[cli,lib]** Option to disable the complete propagation Requirement
- **[lib,cli]** Support non-ascii domain name (punnycode)
- **[cli,lib]** Add configurable timeout when obtaining certificates
- **[cli]** Archive revoked certificates
- **[cli]** Add command to list certificates.
- **[cli]** support for renew with CSR
- **[cli]** add SAN on renew
- **[lib]** Adds `Remove` for challenges
- **[lib]** Add version to xenolf-acme in User-Agent.
- **[dnsprovider]** The ability for a DNS provider to solve the challenge sequentially
- **[dnsprovider]** Add DNS provider for "HTTP request".
- **[dnsprovider]** Add DNS Provider for Vscale
- **[dnsprovider]** Add DNS Provider for TransIP
- **[dnsprovider]** Add DNS Provider for inwx
- **[dnsprovider]** alidns: add support to handle more than 20 domains

### Changed

- **[lib]** Check all challenges in a predictable order
- **[lib]** Poll authz URL instead of challenge URL
- **[lib]** Check all nameservers in a predictable order
- **[lib]** Logs every iteration of waiting for the propagation
- **[cli]** `--http`: enable HTTP challenge **important**
- **[cli]** `--http.port`: previously named `--http`
- **[cli]** `--http.webroot`: previously named `--webroot`
- **[cli]** `--http.memcached-host`: previously named `--memcached-host`
- **[cli]** `--tls`: enable TLS challenge **important**
- **[cli]** `--tls.port`:  previously named `--tls`
- **[cli]** `--dns.resolvers`: previously named `--dns-resolvers`
- **[cli]** the option `--days` of the command `renew` has default value (`15`)
- **[dnsprovider]** gcloud: Use GCE_PROJECT for project always, if specified

### Removed

- **[lib]** Remove `SetHTTP01Address`
- **[lib]** Remove `SetTLSALPN01Address`
- **[lib]** Remove `Exclude`
- **[cli]** Remove `--exclude`, `-x` 

### Fixed

- **[lib]** Fixes revocation for subdomains and non-ascii domains
- **[lib]** Disable pending authorizations
- **[dnsprovider]** transip: concurrent access to the API.
- **[dnsprovider]** gcloud: fix for wildcard
- **[dnsprovider]** Azure: Do not overwrite existing TXT records
- **[dnsprovider]** fix: Cloudflare error.

## [v1.2.0](https://github.com/go-acme/lego/releases/tag/v1.2.0) (2018-11-04)

### Added

- **[dnsprovider]** Add DNS Provider for ConoHa DNS
- **[dnsprovider]** Add DNS Provider for MyDNS.jp
- **[dnsprovider]** Add DNS Provider for Selectel

### Fixed

- **[dnsprovider]** netcup: make unmarshalling of api-responses more lenient.

### Changed

- **[dnsprovider]** aurora: change DNS client
- **[dnsprovider]** azure: update auth to support instance metadata service
- **[dnsprovider]** dnsmadeeasy: log response body on error
- **[lib]** TLS-ALPN-01: Update idPeAcmeIdentifierV1, draft refs.
- **[lib]** Do not send a JWS body when POSTing challenges.
- **[lib]** Support POST-as-GET.

## [v1.1.0](https://github.com/go-acme/lego/releases/tag/v1.1.0) (2018-10-16)

### Added

- **[lib]** TLS-ALPN-01 Challenge
- **[cli]** Add filename parameter
- **[dnsprovider]** Allow to configure TTL, interval and timeout
- **[dnsprovider]** Add support for reading DNS provider setup from files
- **[dnsprovider]** Add DNS Provider for ACME-DNS
- **[dnsprovider]** Add DNS Provider for ALIYUN DNS
- **[dnsprovider]** Add DNS Provider for DreamHost
- **[dnsprovider]** Add DNS provider for hosting.de
- **[dnsprovider]** Add DNS Provider for IIJ
- **[dnsprovider]** Add DNS Provider for netcup
- **[dnsprovider]** Add DNS Provider for NIFCLOUD DNS
- **[dnsprovider]** Add DNS Provider for SAKURA Cloud
- **[dnsprovider]** Add DNS Provider for Stackpath
- **[dnsprovider]** Add DNS Provider for VegaDNS
- **[dnsprovider]** exec: add EXEC_MODE=RAW support.
- **[dnsprovider]** cloudflare: support for CF_API_KEY and CF_API_EMAIL

### Fixed

- **[lib]** Don't trust identifiers order.
- **[lib]** Fix missing issuer certificates from Let's Encrypt
- **[dnsprovider]** duckdns: fix TXT record update url
- **[dnsprovider]** duckdns: fix subsubdomain
- **[dnsprovider]** gcloud: update findTxtRecords to use Name=fqdn and Type=TXT
- **[dnsprovider]** lightsail: Fix Domain does not exist error
- **[dnsprovider]** ns1: use the authoritative zone and not the domain name
- **[dnsprovider]** ovh: check error to avoid panic due to nil client

### Changed

- **[lib]** Submit all dns records up front, then validate serially

## [v1.0.0](https://github.com/go-acme/lego/releases/tag/v1.0.0) (2018-05-30)

### Changed

- **[lib]** ACME v2 Support.
- **[dnsprovider]** Renamed `/providers/dns/googlecloud` to `/providers/dns/gcloud`.
- **[dnsprovider]** Modified Google Cloud provider `gcloud.NewDNSProviderServiceAccount` function to extract the project id directly from the service account file.
- **[dnsprovider]** Made errors more verbose for the Cloudflare provider.

## [v0.5.0](https://github.com/go-acme/lego/releases/tag/v0.5.0) (2018-05-29)

### Added

- **[dnsprovider]** Add DNS challenge provider `exec`
- **[dnsprovider]** Add DNS Provider for Akamai FastDNS
- **[dnsprovider]** Add DNS Provider for Bluecat DNS
- **[dnsprovider]** Add DNS Provider for CloudXNS
- **[dnsprovider]** Add DNS Provider for Duck DNS
- **[dnsprovider]** Add DNS Provider for Gandi Beta Platform (LiveDNS)
- **[dnsprovider]** Add DNS Provider for GleSYS API
- **[dnsprovider]** Add DNS Provider for GoDaddy
- **[dnsprovider]** Add DNS Provider for Lightsail
- **[dnsprovider]** Add DNS Provider for Name.com

### Fixed

- **[dnsprovider]** Azure: Added missing environment variable in the comments
- **[dnsprovider]** PowerDNS: Fix zone URL, add leading slash.
- **[dnsprovider]** DNSimple: Fix api
- **[cli]** Correct help text for `--dns-resolvers` default.
- **[cli]** renew/revoke - don't panic on wrong account.
- **[lib]** Fix zone detection for cross-zone cnames.
- **[lib]** Use proxies from environment when making outbound http connections.

### Changed

- **[lib]** Users of an effective top-level domain can use the DNS challenge.
- **[dnsprovider]** Azure: Refactor to work with new Azure SDK version.
- **[dnsprovider]** Cloudflare and Azure: Adding output of which envvars are missing.
- **[dnsprovider]** Dyn DNS: Slightly improve provider error reporting.
- **[dnsprovider]** Exoscale: update to latest egoscale version.
- **[dnsprovider]** Route53: Use NewSessionWithOptions instead of deprecated New.

## [0.4.1](https://github.com/go-acme/lego/releases/tag/0.4.1) (2017-09-26)

### Added

- lib: A new DNS provider for OTC.
- lib: The `AWS_HOSTED_ZONE_ID` environment variable for the Route53 DNS provider to directly specify the zone.
- lib: The `RFC2136_TIMEOUT` environment variable to make the timeout for the RFC2136 provider configurable.
- lib: The `GCE_SERVICE_ACCOUNT_FILE` environment variable to specify a service account file for the Google Cloud DNS provider.

### Fixed

- lib: Fixed an authentication issue with the latest Azure SDK.

## [0.4.0](https://github.com/go-acme/lego/releases/tag/0.4.0) (2017-07-13)

### Added

- CLI: The `--http-timeout` switch. This allows for an override of the default client HTTP timeout.
- lib: The `HTTPClient` field. This allows for an override of the default HTTP timeout for library HTTP requests.
- CLI: The `--dns-timeout` switch. This allows for an override of the default DNS timeout for library DNS requests.
- lib: The `DNSTimeout` switch. This allows for an override of the default client DNS timeout.
- lib: The `QueryRegistration` function on `acme.Client`. This performs a POST on the client registration's URI and gets the updated registration info.
- lib: The `DeleteRegistration` function on `acme.Client`. This deletes the registration as currently configured in the client.
- lib: The `ObtainCertificateForCSR` function on `acme.Client`. The function allows to request a certificate for an already existing CSR.
- CLI: The `--csr` switch. Allows to use already existing CSRs for certificate requests on the command line.
- CLI: The `--pem` flag. This will change the certificate output, so it outputs a .pem file concatanating the .key and .crt files together.
- CLI: The `--dns-resolvers` flag. Allows for users to override the default DNS servers used for recursive lookup.
- lib: Added a memcached provider for the HTTP challenge.
- CLI: The `--memcached-host` flag. This allows to use memcached for challenge storage.
- CLI: The `--must-staple` flag. This enables OCSP must staple in the generated CSR.
- lib: The library will now honor entries in your resolv.conf.
- lib: Added a field `IssuerCertificate` to the `CertificateResource` struct.
- lib: A new DNS provider for OVH.
- lib: A new DNS provider for DNSMadeEasy.
- lib: A new DNS provider for Linode.
- lib: A new DNS provider for AuroraDNS.
- lib: A new DNS provider for NS1.
- lib: A new DNS provider for Azure DNS.
- lib: A new DNS provider for Rackspace DNS.
- lib: A new DNS provider for Exoscale DNS.
- lib: A new DNS provider for DNSPod.

### Changed

- lib: Exported the `PreCheckDNS` field so library users can manage the DNS check in tests.
- lib: The library will now skip challenge solving if a valid Authz already exists.

### Removed

- lib: The library will no longer check for auto-renewed certificates. This has been removed from the spec and is not supported in Boulder.

### Fixed

- lib: Fix a problem with the Route53 provider where it was possible the verification was published to a private zone.
- lib: Loading an account from file should fail if an integral part is nil
- lib: Fix a potential issue where the Dyn provider could resolve to an incorrect zone.
- lib: If a registration encounteres a conflict, the old registration is now recovered.
- CLI: The account.json file no longer has the executable flag set.
- lib: Made the client registration more robust in case of a 403 HTTP response.
- lib: Fixed an issue with zone lookups when they have a CNAME in another zone.
- lib: Fixed the lookup for the authoritative zone for Google Cloud.
- lib: Fixed a race condition in the nonce store.
- lib: The Google Cloud provider now removes old entries before trying to add new ones.
- lib: Fixed a condition where we could stall due to an early error condition.
- lib: Fixed an issue where Authz object could end up in an active state after an error condition.

## [0.3.1](https://github.com/go-acme/lego/releases/tag/0.3.1) (2016-04-19)

### Added

- lib: A new DNS provider for Vultr.

### Fixed

- lib: DNS Provider for DigitalOcean could not handle subdomains properly.
- lib: handleHTTPError should only try to JSON decode error messages with the right content type.
- lib: The propagation checker for the DNS challenge would not retry on send errors.

## [0.3.0](https://github.com/go-acme/lego/releases/tag/0.3.0) (2016-03-19)

### Added

- CLI: The `--dns` switch. To include the DNS challenge for consideration. When using this switch, all other solvers are disabled. Supported are the following solvers: cloudflare, digitalocean, dnsimple, dyn, gandi, googlecloud, namecheap, route53, rfc2136 and manual.
- CLI: The `--accept-tos`  switch. Indicates your acceptance of the Let's Encrypt terms of service without prompting you.
- CLI: The `--webroot` switch. The HTTP-01 challenge may now be completed by dropping a file into a webroot. When using this switch, all other solvers are disabled.
- CLI: The `--key-type` switch. This replaces the `--rsa-key-size` switch and supports the following key types: EC256, EC384, RSA2048, RSA4096 and RSA8192.
- CLI: The `--dnshelp` switch. This displays a more in-depth help topic for DNS solvers.
- CLI: The `--no-bundle` sub switch for the `run` and `renew` commands. When this switch is set, the CLI will not bundle the issuer certificate with your certificate.
- lib: A new type for challenge identifiers `Challenge`
- lib: A new interface for custom challenge providers `acme.ChallengeProvider`
- lib: A new interface for DNS-01 providers to allow for custom timeouts for the validation function `acme.ChallengeProviderTimeout`
- lib: SetChallengeProvider function. Pass a challenge identifier and a Provider to replace the default behaviour of a challenge.
- lib: The DNS-01 challenge has been implemented with modular solvers using the `ChallengeProvider` interface. Included solvers are: cloudflare, digitalocean, dnsimple, gandi, namecheap, route53, rfc2136 and manual.
- lib: The `acme.KeyType` type was added and is used for the configuration of crypto parameters for RSA and EC keys. Valid KeyTypes are: EC256, EC384, RSA2048, RSA4096 and RSA8192.

### Changed

- lib: ExcludeChallenges now expects to be passed an array of `Challenge` types.
- lib: HTTP-01 now supports custom solvers using the `ChallengeProvider` interface.
- lib: TLS-SNI-01 now supports custom solvers using the `ChallengeProvider` interface.
- lib: The `GetPrivateKey` function in the `acme.User` interface is now expected to return a `crypto.PrivateKey` instead of an `rsa.PrivateKey` for EC compat.
- lib: The `acme.NewClient` function now expects an `acme.KeyType` instead of the keyBits parameter.
 
### Removed

- CLI: The `rsa-key-size` switch was removed in favor of `key-type` to support EC keys.

### Fixed

- lib: Fixed a race condition in HTTP-01
- lib: Fixed an issue where status codes on ACME challenge responses could lead to no action being taken.
- lib: Fixed a regression when calling the Renew function with a SAN certificate.

## [0.2.0](https://github.com/go-acme/lego/releases/tag/0.2.0) (2016-01-09)

### Added

- CLI: The `--exclude` or `-x` switch. To exclude a challenge from being solved.
- CLI: The `--http` switch. To set the listen address and port of HTTP based challenges. Supports `host:port` and `:port` for any interface.
- CLI: The `--tls` switch. To set the listen address and port of TLS based challenges. Supports `host:port` and `:port` for any interface.
- CLI: The `--reuse-key` switch for the `renew` operation. This lets you reuse an existing private key for renewals.
- lib: ExcludeChallenges function. Pass an array of challenge identifiers to exclude them from solving.
- lib: SetHTTPAddress function. Pass a port to set the listen port for HTTP based challenges.
- lib: SetTLSAddress function. Pass a port to set the listen port of TLS based challenges.
- lib: acme.UserAgent variable. Use this to customize the user agent on all requests sent by lego.

### Changed

- lib: NewClient does no longer accept the optPort parameter
- lib: ObtainCertificate now returns a SAN certificate if you pass more than one domain.
- lib: GetOCSPForCert now returns the parsed OCSP response instead of just the status.
- lib: ObtainCertificate has a new parameter `privKey crypto.PrivateKey` which lets you reuse an existing private key for new certificates.
- lib: RenewCertificate now expects the PrivateKey property of the CertificateResource to be set only if you want to reuse the key.

### Removed

- CLI: The `--port` switch was removed.
- lib: RenewCertificate does no longer offer to also revoke your old certificate.

### Fixed

- CLI: Fix logic using the `--days` parameter for renew

## [0.1.1](https://github.com/go-acme/lego/releases/tag/0.1.1) (2015-12-18)

### Added

- CLI: Added a way to automate renewal through a cronjob using the --days parameter to renew

### Changed

- lib: Improved log output on challenge failures.

### Fixed

- CLI: The short parameter for domains would not get accepted
- CLI: The cli did not return proper exit codes on error library errors.
- lib: RenewCertificate did not properly renew SAN certificates.

### Security

- lib: Fix possible DOS on GetOCSPForCert

## [0.1.0](https://github.com/go-acme/lego/releases/tag/0.1.0) (2015-12-03)

- Initial release
