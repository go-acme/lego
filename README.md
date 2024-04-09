<div align="center">
  <img alt="lego logo" src="./docs/static/images/lego-logo.min.svg">
  <p>Automatic Certificates and HTTPS for everyone.</p>
</div>

# Lego

Let's Encrypt client and ACME library written in Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/go-acme/lego/v4.svg)](https://pkg.go.dev/github.com/go-acme/lego/v4)
[![Build Status](https://github.com//go-acme/lego/workflows/Main/badge.svg?branch=master)](https://github.com//go-acme/lego/actions)
[![Docker Pulls](https://img.shields.io/docker/pulls/goacme/lego.svg)](https://hub.docker.com/r/goacme/lego/)

## Features

- ACME v2 [RFC 8555](https://www.rfc-editor.org/rfc/rfc8555.html)
  - Support [RFC 8737](https://www.rfc-editor.org/rfc/rfc8737.html): TLS Application‑Layer Protocol Negotiation (ALPN) Challenge Extension
  - Support [RFC 8738](https://www.rfc-editor.org/rfc/rfc8738.html): certificates for IP addresses
  - Support [draft-ietf-acme-ari-03](https://datatracker.ietf.org/doc/draft-ietf-acme-ari/): Renewal Information (ARI) Extension
- Register with CA
- Obtain certificates, both from scratch or with an existing CSR
- Renew certificates
- Revoke certificates
- Robust implementation of all ACME challenges
  - HTTP (http-01)
  - DNS (dns-01)
  - TLS (tls-alpn-01)
- SAN certificate support
- [CNAME support](https://letsencrypt.org/2019/10/09/onboarding-your-customers-with-lets-encrypt-and-acme.html) by default
- Comes with multiple optional [DNS providers](https://go-acme.github.io/lego/dns)
- [Custom challenge solvers](https://go-acme.github.io/lego/usage/library/writing-a-challenge-solver/)
- Certificate bundling
- OCSP helper function

## Installation

How to [install](https://go-acme.github.io/lego/installation/).

## Usage

- as a [CLI](https://go-acme.github.io/lego/usage/cli)
- as a [library](https://go-acme.github.io/lego/usage/library)

## Documentation

Documentation is hosted live at https://go-acme.github.io/lego/.

## DNS providers

Detailed documentation is available [here](https://go-acme.github.io/lego/dns).

<!-- START DNS PROVIDERS LIST -->

|                                                                                 |                                                                                 |                                                                                 |                                                                                 |
|---------------------------------------------------------------------------------|---------------------------------------------------------------------------------|---------------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| [Akamai EdgeDNS](https://go-acme.github.io/lego/dns/edgedns/)                   | [Alibaba Cloud DNS](https://go-acme.github.io/lego/dns/alidns/)                 | [all-inkl](https://go-acme.github.io/lego/dns/allinkl/)                         | [Amazon Lightsail](https://go-acme.github.io/lego/dns/lightsail/)               |
| [Amazon Route 53](https://go-acme.github.io/lego/dns/route53/)                  | [ArvanCloud](https://go-acme.github.io/lego/dns/arvancloud/)                    | [Aurora DNS](https://go-acme.github.io/lego/dns/auroradns/)                     | [Autodns](https://go-acme.github.io/lego/dns/autodns/)                          |
| [Azure (deprecated)](https://go-acme.github.io/lego/dns/azure/)                 | [Azure DNS](https://go-acme.github.io/lego/dns/azuredns/)                       | [Bindman](https://go-acme.github.io/lego/dns/bindman/)                          | [Bluecat](https://go-acme.github.io/lego/dns/bluecat/)                          |
| [Brandit](https://go-acme.github.io/lego/dns/brandit/)                          | [Bunny](https://go-acme.github.io/lego/dns/bunny/)                              | [Checkdomain](https://go-acme.github.io/lego/dns/checkdomain/)                  | [Civo](https://go-acme.github.io/lego/dns/civo/)                                |
| [Cloud.ru](https://go-acme.github.io/lego/dns/cloudru/)                         | [CloudDNS](https://go-acme.github.io/lego/dns/clouddns/)                        | [Cloudflare](https://go-acme.github.io/lego/dns/cloudflare/)                    | [ClouDNS](https://go-acme.github.io/lego/dns/cloudns/)                          |
| [CloudXNS](https://go-acme.github.io/lego/dns/cloudxns/)                        | [ConoHa](https://go-acme.github.io/lego/dns/conoha/)                            | [Constellix](https://go-acme.github.io/lego/dns/constellix/)                    | [CPanel/WHM](https://go-acme.github.io/lego/dns/cpanel/)                        |
| [Derak Cloud](https://go-acme.github.io/lego/dns/derak/)                        | [deSEC.io](https://go-acme.github.io/lego/dns/desec/)                           | [Designate DNSaaS for Openstack](https://go-acme.github.io/lego/dns/designate/) | [Digital Ocean](https://go-acme.github.io/lego/dns/digitalocean/)               |
| [DNS Made Easy](https://go-acme.github.io/lego/dns/dnsmadeeasy/)                | [dnsHome.de](https://go-acme.github.io/lego/dns/dnshomede/)                     | [DNSimple](https://go-acme.github.io/lego/dns/dnsimple/)                        | [DNSPod (deprecated)](https://go-acme.github.io/lego/dns/dnspod/)               |
| [Domain Offensive (do.de)](https://go-acme.github.io/lego/dns/dode/)            | [Domeneshop](https://go-acme.github.io/lego/dns/domeneshop/)                    | [DreamHost](https://go-acme.github.io/lego/dns/dreamhost/)                      | [Duck DNS](https://go-acme.github.io/lego/dns/duckdns/)                         |
| [Dyn](https://go-acme.github.io/lego/dns/dyn/)                                  | [Dynu](https://go-acme.github.io/lego/dns/dynu/)                                | [EasyDNS](https://go-acme.github.io/lego/dns/easydns/)                          | [Efficient IP](https://go-acme.github.io/lego/dns/efficientip/)                 |
| [Epik](https://go-acme.github.io/lego/dns/epik/)                                | [Exoscale](https://go-acme.github.io/lego/dns/exoscale/)                        | [External program](https://go-acme.github.io/lego/dns/exec/)                    | [freemyip.com](https://go-acme.github.io/lego/dns/freemyip/)                    |
| [G-Core](https://go-acme.github.io/lego/dns/gcore/)                             | [Gandi Live DNS (v5)](https://go-acme.github.io/lego/dns/gandiv5/)              | [Gandi](https://go-acme.github.io/lego/dns/gandi/)                              | [Glesys](https://go-acme.github.io/lego/dns/glesys/)                            |
| [Go Daddy](https://go-acme.github.io/lego/dns/godaddy/)                         | [Google Cloud](https://go-acme.github.io/lego/dns/gcloud/)                      | [Google Domains](https://go-acme.github.io/lego/dns/googledomains/)             | [Hetzner](https://go-acme.github.io/lego/dns/hetzner/)                          |
| [Hosting.de](https://go-acme.github.io/lego/dns/hostingde/)                     | [Hosttech](https://go-acme.github.io/lego/dns/hosttech/)                        | [HTTP request](https://go-acme.github.io/lego/dns/httpreq/)                     | [http.net](https://go-acme.github.io/lego/dns/httpnet/)                         |
| [Hurricane Electric DNS](https://go-acme.github.io/lego/dns/hurricane/)         | [HyperOne](https://go-acme.github.io/lego/dns/hyperone/)                        | [IBM Cloud (SoftLayer)](https://go-acme.github.io/lego/dns/ibmcloud/)           | [IIJ DNS Platform Service](https://go-acme.github.io/lego/dns/iijdpf/)          |
| [Infoblox](https://go-acme.github.io/lego/dns/infoblox/)                        | [Infomaniak](https://go-acme.github.io/lego/dns/infomaniak/)                    | [Internet Initiative Japan](https://go-acme.github.io/lego/dns/iij/)            | [Internet.bs](https://go-acme.github.io/lego/dns/internetbs/)                   |
| [INWX](https://go-acme.github.io/lego/dns/inwx/)                                | [Ionos](https://go-acme.github.io/lego/dns/ionos/)                              | [IPv64](https://go-acme.github.io/lego/dns/ipv64/)                              | [iwantmyname](https://go-acme.github.io/lego/dns/iwantmyname/)                  |
| [Joker](https://go-acme.github.io/lego/dns/joker/)                              | [Joohoi's ACME-DNS](https://go-acme.github.io/lego/dns/acme-dns/)               | [Liara](https://go-acme.github.io/lego/dns/liara/)                              | [Linode (v4)](https://go-acme.github.io/lego/dns/linode/)                       |
| [Liquid Web](https://go-acme.github.io/lego/dns/liquidweb/)                     | [Loopia](https://go-acme.github.io/lego/dns/loopia/)                            | [LuaDNS](https://go-acme.github.io/lego/dns/luadns/)                            | [Mail-in-a-Box](https://go-acme.github.io/lego/dns/mailinabox/)                 |
| [Manual](https://go-acme.github.io/lego/dns/manual/)                            | [Metaname](https://go-acme.github.io/lego/dns/metaname/)                        | [MyDNS.jp](https://go-acme.github.io/lego/dns/mydnsjp/)                         | [MythicBeasts](https://go-acme.github.io/lego/dns/mythicbeasts/)                |
| [Name.com](https://go-acme.github.io/lego/dns/namedotcom/)                      | [Namecheap](https://go-acme.github.io/lego/dns/namecheap/)                      | [Namesilo](https://go-acme.github.io/lego/dns/namesilo/)                        | [NearlyFreeSpeech.NET](https://go-acme.github.io/lego/dns/nearlyfreespeech/)    |
| [Netcup](https://go-acme.github.io/lego/dns/netcup/)                            | [Netlify](https://go-acme.github.io/lego/dns/netlify/)                          | [Nicmanager](https://go-acme.github.io/lego/dns/nicmanager/)                    | [NIFCloud](https://go-acme.github.io/lego/dns/nifcloud/)                        |
| [Njalla](https://go-acme.github.io/lego/dns/njalla/)                            | [Nodion](https://go-acme.github.io/lego/dns/nodion/)                            | [NS1](https://go-acme.github.io/lego/dns/ns1/)                                  | [Open Telekom Cloud](https://go-acme.github.io/lego/dns/otc/)                   |
| [Oracle Cloud](https://go-acme.github.io/lego/dns/oraclecloud/)                 | [OVH](https://go-acme.github.io/lego/dns/ovh/)                                  | [plesk.com](https://go-acme.github.io/lego/dns/plesk/)                          | [Porkbun](https://go-acme.github.io/lego/dns/porkbun/)                          |
| [PowerDNS](https://go-acme.github.io/lego/dns/pdns/)                            | [Rackspace](https://go-acme.github.io/lego/dns/rackspace/)                      | [RcodeZero](https://go-acme.github.io/lego/dns/rcodezero/)                      | [reg.ru](https://go-acme.github.io/lego/dns/regru/)                             |
| [RFC2136](https://go-acme.github.io/lego/dns/rfc2136/)                          | [RimuHosting](https://go-acme.github.io/lego/dns/rimuhosting/)                  | [Sakura Cloud](https://go-acme.github.io/lego/dns/sakuracloud/)                 | [Scaleway](https://go-acme.github.io/lego/dns/scaleway/)                        |
| [Selectel v2](https://go-acme.github.io/lego/dns/selectelv2/)                   | [Selectel](https://go-acme.github.io/lego/dns/selectel/)                        | [Servercow](https://go-acme.github.io/lego/dns/servercow/)                      | [Shellrent](https://go-acme.github.io/lego/dns/shellrent/)                      |
| [Simply.com](https://go-acme.github.io/lego/dns/simply/)                        | [Sonic](https://go-acme.github.io/lego/dns/sonic/)                              | [Stackpath](https://go-acme.github.io/lego/dns/stackpath/)                      | [Tencent Cloud DNS](https://go-acme.github.io/lego/dns/tencentcloud/)           |
| [TransIP](https://go-acme.github.io/lego/dns/transip/)                          | [UKFast SafeDNS](https://go-acme.github.io/lego/dns/safedns/)                   | [Ultradns](https://go-acme.github.io/lego/dns/ultradns/)                        | [Variomedia](https://go-acme.github.io/lego/dns/variomedia/)                    |
| [VegaDNS](https://go-acme.github.io/lego/dns/vegadns/)                          | [Vercel](https://go-acme.github.io/lego/dns/vercel/)                            | [Versio.[nl/eu/uk]](https://go-acme.github.io/lego/dns/versio/)                 | [VinylDNS](https://go-acme.github.io/lego/dns/vinyldns/)                        |
| [VK Cloud](https://go-acme.github.io/lego/dns/vkcloud/)                         | [Vscale](https://go-acme.github.io/lego/dns/vscale/)                            | [Vultr](https://go-acme.github.io/lego/dns/vultr/)                              | [Webnames](https://go-acme.github.io/lego/dns/webnames/)                        |
| [Websupport](https://go-acme.github.io/lego/dns/websupport/)                    | [WEDOS](https://go-acme.github.io/lego/dns/wedos/)                              | [Yandex 360](https://go-acme.github.io/lego/dns/yandex360/)                     | [Yandex Cloud](https://go-acme.github.io/lego/dns/yandexcloud/)                 |
| [Yandex PDD](https://go-acme.github.io/lego/dns/yandex/)                        | [Zone.ee](https://go-acme.github.io/lego/dns/zoneee/)                           | [Zonomi](https://go-acme.github.io/lego/dns/zonomi/)                            |                                                                                 |

<!-- END DNS PROVIDERS LIST -->

If your DNS provider is not supported, please open an [issue](https://github.com/go-acme/lego/issues/new?assignees=&labels=enhancement%2C+new-provider&template=new_dns_provider.md).
