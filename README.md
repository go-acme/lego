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
  - Support [draft-aaron-acme-profiles-00](https://datatracker.ietf.org/doc/draft-aaron-acme-profiles/): Profiles Extension
- Comes with about [150 DNS providers](https://go-acme.github.io/lego/dns)
- Register with CA
- Obtain certificates, both from scratch or with an existing CSR
- Renew certificates
- Revoke certificates
- Robust implementation of ACME challenges:
  - HTTP (http-01)
  - DNS (dns-01)
  - TLS (tls-alpn-01)
- SAN certificate support
- [CNAME support](https://letsencrypt.org/2019/10/09/onboarding-your-customers-with-lets-encrypt-and-acme.html) by default
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

<table><tr>
  <td><a href="https://go-acme.github.io/lego/dns/edgedns/">Akamai EdgeDNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/alidns/">Alibaba Cloud DNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/allinkl/">all-inkl</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/lightsail/">Amazon Lightsail</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/route53/">Amazon Route 53</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/arvancloud/">ArvanCloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/auroradns/">Aurora DNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/autodns/">Autodns</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/azure/">Azure (deprecated)</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/azuredns/">Azure DNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/bindman/">Bindman</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/bluecat/">Bluecat</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/bookmyname/">BookMyName</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/brandit/">Brandit (deprecated)</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/bunny/">Bunny</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/checkdomain/">Checkdomain</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/civo/">Civo</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/cloudru/">Cloud.ru</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/clouddns/">CloudDNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/cloudflare/">Cloudflare</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/cloudns/">ClouDNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/cloudxns/">CloudXNS (Deprecated)</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/conoha/">ConoHa</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/constellix/">Constellix</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/corenetworks/">Core-Networks</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/cpanel/">CPanel/WHM</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/derak/">Derak Cloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/desec/">deSEC.io</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/designate/">Designate DNSaaS for Openstack</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/digitalocean/">Digital Ocean</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/directadmin/">DirectAdmin</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/dnsmadeeasy/">DNS Made Easy</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/dnshomede/">dnsHome.de</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/dnsimple/">DNSimple</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/dnspod/">DNSPod (deprecated)</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/dode/">Domain Offensive (do.de)</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/domeneshop/">Domeneshop</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/dreamhost/">DreamHost</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/duckdns/">Duck DNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/dyn/">Dyn</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/dynu/">Dynu</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/easydns/">EasyDNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/efficientip/">Efficient IP</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/epik/">Epik</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/exoscale/">Exoscale</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/exec/">External program</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/freemyip/">freemyip.com</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/gcore/">G-Core</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/gandi/">Gandi</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/gandiv5/">Gandi Live DNS (v5)</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/glesys/">Glesys</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/godaddy/">Go Daddy</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/gcloud/">Google Cloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/googledomains/">Google Domains</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/hetzner/">Hetzner</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/hostingde/">Hosting.de</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/hosttech/">Hosttech</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/httpreq/">HTTP request</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/httpnet/">http.net</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/huaweicloud/">Huawei Cloud</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/hurricane/">Hurricane Electric DNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/hyperone/">HyperOne</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/ibmcloud/">IBM Cloud (SoftLayer)</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/iijdpf/">IIJ DNS Platform Service</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/infoblox/">Infoblox</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/infomaniak/">Infomaniak</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/iij/">Internet Initiative Japan</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/internetbs/">Internet.bs</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/inwx/">INWX</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/ionos/">Ionos</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/ipv64/">IPv64</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/iwantmyname/">iwantmyname</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/joker/">Joker</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/acme-dns/">Joohoi&#39;s ACME-DNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/liara/">Liara</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/limacity/">Lima-City</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/linode/">Linode (v4)</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/liquidweb/">Liquid Web</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/loopia/">Loopia</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/luadns/">LuaDNS</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/mailinabox/">Mail-in-a-Box</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/manageengine/">ManageEngine CloudDNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/manual/">Manual</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/metaname/">Metaname</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/metaregistrar/">Metaregistrar</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/mijnhost/">mijn.host</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/mittwald/">Mittwald</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/myaddr/">myaddr.{tools,dev,io}</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/mydnsjp/">MyDNS.jp</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/mythicbeasts/">MythicBeasts</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/namedotcom/">Name.com</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/namecheap/">Namecheap</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/namesilo/">Namesilo</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/nearlyfreespeech/">NearlyFreeSpeech.NET</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/netcup/">Netcup</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/netlify/">Netlify</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/nicmanager/">Nicmanager</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/nifcloud/">NIFCloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/njalla/">Njalla</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/nodion/">Nodion</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/ns1/">NS1</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/otc/">Open Telekom Cloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/oraclecloud/">Oracle Cloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/ovh/">OVH</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/plesk/">plesk.com</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/porkbun/">Porkbun</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/pdns/">PowerDNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/rackspace/">Rackspace</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/rainyun/">Rain Yun/雨云</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/rcodezero/">RcodeZero</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/regru/">reg.ru</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/regfish/">Regfish</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/rfc2136/">RFC2136</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/rimuhosting/">RimuHosting</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/sakuracloud/">Sakura Cloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/scaleway/">Scaleway</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/selectel/">Selectel</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/selectelv2/">Selectel v2</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/selfhostde/">SelfHost.(de|eu)</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/servercow/">Servercow</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/shellrent/">Shellrent</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/simply/">Simply.com</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/sonic/">Sonic</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/spaceship/">Spaceship</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/stackpath/">Stackpath</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/technitium/">Technitium</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/tencentcloud/">Tencent Cloud DNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/timewebcloud/">Timeweb Cloud</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/transip/">TransIP</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/safedns/">UKFast SafeDNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/ultradns/">Ultradns</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/variomedia/">Variomedia</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/vegadns/">VegaDNS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/vercel/">Vercel</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/versio/">Versio.[nl|eu|uk]</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/vinyldns/">VinylDNS</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/vkcloud/">VK Cloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/volcengine/">Volcano Engine/火山引擎</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/vscale/">Vscale</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/vultr/">Vultr</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/webnames/">Webnames</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/websupport/">Websupport</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/wedos/">WEDOS</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/westcn/">West.cn/西部数码</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/yandex360/">Yandex 360</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/yandexcloud/">Yandex Cloud</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/yandex/">Yandex PDD</a></td>
  <td><a href="https://go-acme.github.io/lego/dns/zoneee/">Zone.ee</a></td>
</tr><tr>
  <td><a href="https://go-acme.github.io/lego/dns/zonomi/">Zonomi</a></td>
  <td></td>
  <td></td>
  <td></td>
</tr></table>

<!-- END DNS PROVIDERS LIST -->

If your DNS provider is not supported, please open an [issue](https://github.com/go-acme/lego/issues/new?assignees=&labels=enhancement%2C+new-provider&template=new_dns_provider.md).
