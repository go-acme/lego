package cmd

// CODE GENERATED AUTOMATICALLY
// THIS FILE MUST NOT BE EDITED BY HAND

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/go-acme/lego/log"
)

func allDNSCodes() string {
	providers := []string{
		"manual",
		"acme-dns",
		"alidns",
		"auroradns",
		"azure",
		"bindman",
		"bluecat",
		"cloudflare",
		"cloudns",
		"cloudxns",
		"conoha",
		"designate",
		"digitalocean",
		"dnsimple",
		"dnsmadeeasy",
		"dnspod",
		"dode",
		"dreamhost",
		"duckdns",
		"dyn",
		"easydns",
		"exec",
		"exoscale",
		"fastdns",
		"gandi",
		"gandiv5",
		"gcloud",
		"glesys",
		"godaddy",
		"hostingde",
		"httpreq",
		"iij",
		"inwx",
		"joker",
		"lightsail",
		"linode",
		"linodev4",
		"mydnsjp",
		"namecheap",
		"namedotcom",
		"netcup",
		"nifcloud",
		"ns1",
		"oraclecloud",
		"otc",
		"ovh",
		"pdns",
		"rackspace",
		"rfc2136",
		"route53",
		"sakuracloud",
		"selectel",
		"stackpath",
		"transip",
		"vegadns",
		"vscale",
		"vultr",
		"zoneee",
	}
	sort.Strings(providers)
	return strings.Join(providers, ", ")
}

func displayDNSHelp(name string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	switch name {

	case "acme-dns":
		// generated from: providers/dns/acmedns/acmedns.toml
		fmt.Fprintln(w, `Configuration for Joohoi's ACME-DNS.`)
		fmt.Fprintln(w, `Code:	'acme-dns'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "ACME_DNS_API_BASE":	The ACME-DNS API address`)
		fmt.Fprintln(w, `	- "ACME_DNS_STORAGE_PATH":	The ACME-DNS JSON account data file. A per-domain account will be registered/persisted to this file and used for TXT updates.`)
		fmt.Fprintln(w)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/acme-dns`)

	case "alidns":
		// generated from: providers/dns/alidns/alidns.toml
		fmt.Fprintln(w, `Configuration for Alibaba Cloud DNS.`)
		fmt.Fprintln(w, `Code:	'alidns'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "ALICLOUD_ACCESS_KEY":	Access key ID`)
		fmt.Fprintln(w, `	- "ALICLOUD_SECRET_KEY":	Access Key secret`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "ALICLOUD_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "ALICLOUD_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "ALICLOUD_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "ALICLOUD_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/alidns`)

	case "auroradns":
		// generated from: providers/dns/auroradns/auroradns.toml
		fmt.Fprintln(w, `Configuration for Aurora DNS.`)
		fmt.Fprintln(w, `Code:	'auroradns'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "AURORA_ENDPOINT":	API endpoint URL`)
		fmt.Fprintln(w, `	- "AURORA_KEY":	User API key`)
		fmt.Fprintln(w, `	- "AURORA_USER_ID":	User ID`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "AURORA_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "AURORA_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "AURORA_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/auroradns`)

	case "azure":
		// generated from: providers/dns/azure/azure.toml
		fmt.Fprintln(w, `Configuration for Azure.`)
		fmt.Fprintln(w, `Code:	'azure'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "AZURE_CLIENT_ID":	Client ID`)
		fmt.Fprintln(w, `	- "AZURE_CLIENT_SECRET":	Client secret`)
		fmt.Fprintln(w, `	- "AZURE_RESOURCE_GROUP":	Resource group`)
		fmt.Fprintln(w, `	- "AZURE_SUBSCRIPTION_ID":	Subscription ID`)
		fmt.Fprintln(w, `	- "AZURE_TENANT_ID":	Tenant ID`)
		fmt.Fprintln(w, `	- "instance metadata service":	If the credentials are **not** set via the environment, then it will attempt to get a bearer token via the [instance metadata service](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service).`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "AZURE_METADATA_ENDPOINT":	Metadata Service endpoint URL`)
		fmt.Fprintln(w, `	- "AZURE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "AZURE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "AZURE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/azure`)

	case "bindman":
		// generated from: providers/dns/bindman/bindman.toml
		fmt.Fprintln(w, `Configuration for Bindman.`)
		fmt.Fprintln(w, `Code:	'bindman'`)
		fmt.Fprintln(w, `Since:	'v2.6.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "BINDMAN_MANAGER_ADDRESS":	The server URL, should have scheme, hostname, and port (if required) of the Bindman-DNS Manager server`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "BINDMAN_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "BINDMAN_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "BINDMAN_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/bindman`)

	case "bluecat":
		// generated from: providers/dns/bluecat/bluecat.toml
		fmt.Fprintln(w, `Configuration for Bluecat.`)
		fmt.Fprintln(w, `Code:	'bluecat'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "BLUECAT_CONFIG_NAME":	Configuration name`)
		fmt.Fprintln(w, `	- "BLUECAT_DNS_VIEW":	External DNS View Name`)
		fmt.Fprintln(w, `	- "BLUECAT_PASSWORD":	API password`)
		fmt.Fprintln(w, `	- "BLUECAT_SERVER_URL":	The server URL, should have scheme, hostname, and port (if required) of the authoritative Bluecat BAM serve`)
		fmt.Fprintln(w, `	- "BLUECAT_USER_NAME":	API username`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "BLUECAT_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "BLUECAT_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "BLUECAT_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "BLUECAT_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/bluecat`)

	case "cloudflare":
		// generated from: providers/dns/cloudflare/cloudflare.toml
		fmt.Fprintln(w, `Configuration for Cloudflare.`)
		fmt.Fprintln(w, `Code:	'cloudflare'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "CLOUDFLARE_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "CLOUDFLARE_EMAIL":	Account email`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "CLOUDFLARE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "CLOUDFLARE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "CLOUDFLARE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "CLOUDFLARE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/cloudflare`)

	case "cloudns":
		// generated from: providers/dns/cloudns/cloudns.toml
		fmt.Fprintln(w, `Configuration for ClouDNS.`)
		fmt.Fprintln(w, `Code:	'cloudns'`)
		fmt.Fprintln(w, `Since:	'v2.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "CLOUDNS_AUTH_ID":	The API user ID`)
		fmt.Fprintln(w, `	- "CLOUDNS_AUTH_PASSWORD":	The password for API user ID`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "CLOUDNS_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "CLOUDNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "CLOUDNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "CLOUDNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/cloudns`)

	case "cloudxns":
		// generated from: providers/dns/cloudxns/cloudxns.toml
		fmt.Fprintln(w, `Configuration for CloudXNS.`)
		fmt.Fprintln(w, `Code:	'cloudxns'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "CLOUDXNS_API_KEY":	The API key`)
		fmt.Fprintln(w, `	- "CLOUDXNS_SECRET_KEY":	THe API secret key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "CLOUDXNS_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "CLOUDXNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "CLOUDXNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "CLOUDXNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/cloudxns`)

	case "conoha":
		// generated from: providers/dns/conoha/conoha.toml
		fmt.Fprintln(w, `Configuration for ConoHa.`)
		fmt.Fprintln(w, `Code:	'conoha'`)
		fmt.Fprintln(w, `Since:	'v1.2.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "CONOHA_API_PASSWORD":	The API password`)
		fmt.Fprintln(w, `	- "CONOHA_API_USERNAME":	The API username`)
		fmt.Fprintln(w, `	- "CONOHA_TENANT_ID":	Tenant ID`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "CONOHA_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "CONOHA_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "CONOHA_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "CONOHA_REGION":	The region`)
		fmt.Fprintln(w, `	- "CONOHA_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/conoha`)

	case "designate":
		// generated from: providers/dns/designate/designate.toml
		fmt.Fprintln(w, `Configuration for Designate DNSaaS for Openstack.`)
		fmt.Fprintln(w, `Code:	'designate'`)
		fmt.Fprintln(w, `Since:	'v2.2.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "OS_AUTH_URL":	Identity endpoint URL`)
		fmt.Fprintln(w, `	- "OS_PASSWORD":	Password`)
		fmt.Fprintln(w, `	- "OS_REGION_NAME":	Region name`)
		fmt.Fprintln(w, `	- "OS_TENANT_NAME":	Tenant name`)
		fmt.Fprintln(w, `	- "OS_USERNAME":	Username`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DESIGNATE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DESIGNATE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DESIGNATE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/designate`)

	case "digitalocean":
		// generated from: providers/dns/digitalocean/digitalocean.toml
		fmt.Fprintln(w, `Configuration for Digital Ocean.`)
		fmt.Fprintln(w, `Code:	'digitalocean'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "DO_AUTH_TOKEN":	Authentication token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DO_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "DO_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DO_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DO_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/digitalocean`)

	case "dnsimple":
		// generated from: providers/dns/dnsimple/dnsimple.toml
		fmt.Fprintln(w, `Configuration for DNSimple.`)
		fmt.Fprintln(w, `Code:	'dnsimple'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "DNSIMPLE_BASE_URL":	API endpoint URL`)
		fmt.Fprintln(w, `	- "DNSIMPLE_OAUTH_TOKEN":	OAuth token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DNSIMPLE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DNSIMPLE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DNSIMPLE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/dnsimple`)

	case "dnsmadeeasy":
		// generated from: providers/dns/dnsmadeeasy/dnsmadeeasy.toml
		fmt.Fprintln(w, `Configuration for DNS Made Easy.`)
		fmt.Fprintln(w, `Code:	'dnsmadeeasy'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "DNSMADEEASY_API_KEY":	The API key`)
		fmt.Fprintln(w, `	- "DNSMADEEASY_API_SECRET":	The API Secret key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DNSMADEEASY_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "DNSMADEEASY_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DNSMADEEASY_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DNSMADEEASY_SANDBOX":	Activate the sandbox (boolean)`)
		fmt.Fprintln(w, `	- "DNSMADEEASY_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/dnsmadeeasy`)

	case "dnspod":
		// generated from: providers/dns/dnspod/dnspod.toml
		fmt.Fprintln(w, `Configuration for DNSPod.`)
		fmt.Fprintln(w, `Code:	'dnspod'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "DNSPOD_API_KEY":	The user token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DNSPOD_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "DNSPOD_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DNSPOD_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DNSPOD_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/dnspod`)

	case "dode":
		// generated from: providers/dns/dode/dode.toml
		fmt.Fprintln(w, `Configuration for Domain Offensive (do.de).`)
		fmt.Fprintln(w, `Code:	'dode'`)
		fmt.Fprintln(w, `Since:	'v2.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "DODE_TOKEN":	API token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DODE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "DODE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DODE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DODE_SEQUENCE_INTERVAL":	Interval between iteration`)
		fmt.Fprintln(w, `	- "DODE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/dode`)

	case "dreamhost":
		// generated from: providers/dns/dreamhost/dreamhost.toml
		fmt.Fprintln(w, `Configuration for DreamHost.`)
		fmt.Fprintln(w, `Code:	'dreamhost'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "DREAMHOST_API_KEY":	The API key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DREAMHOST_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "DREAMHOST_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DREAMHOST_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DREAMHOST_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/dreamhost`)

	case "duckdns":
		// generated from: providers/dns/duckdns/duckdns.toml
		fmt.Fprintln(w, `Configuration for Duck DNS.`)
		fmt.Fprintln(w, `Code:	'duckdns'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "DUCKDNS_TOKEN":	Account token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DUCKDNS_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "DUCKDNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DUCKDNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DUCKDNS_SEQUENCE_INTERVAL":	Interval between iteration`)
		fmt.Fprintln(w, `	- "DUCKDNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/duckdns`)

	case "dyn":
		// generated from: providers/dns/dyn/dyn.toml
		fmt.Fprintln(w, `Configuration for Dyn.`)
		fmt.Fprintln(w, `Code:	'dyn'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "DYN_CUSTOMER_NAME":	Customer name`)
		fmt.Fprintln(w, `	- "DYN_PASSWORD":	Paswword`)
		fmt.Fprintln(w, `	- "DYN_USER_NAME":	User name`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "DYN_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "DYN_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "DYN_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "DYN_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/dyn`)

	case "easydns":
		// generated from: providers/dns/easydns/easydns.toml
		fmt.Fprintln(w, `Configuration for EasyDNS.`)
		fmt.Fprintln(w, `Code:	'easydns'`)
		fmt.Fprintln(w, `Since:	'v2.6.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "EASYDNS_KEY":	API Key`)
		fmt.Fprintln(w, `	- "EASYDNS_TOKEN":	API Token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "EASYDNS_ENDPOINT":	The endpoint URL of the API Server`)
		fmt.Fprintln(w, `	- "EASYDNS_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "EASYDNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "EASYDNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "EASYDNS_SEQUENCE_INTERVAL":	Time between sequential requests`)
		fmt.Fprintln(w, `	- "EASYDNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/easydns`)

	case "exec":
		// generated from: providers/dns/exec/exec.toml
		fmt.Fprintln(w, `Configuration for External program.`)
		fmt.Fprintln(w, `Code:	'exec'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/exec`)

	case "exoscale":
		// generated from: providers/dns/exoscale/exoscale.toml
		fmt.Fprintln(w, `Configuration for Exoscale.`)
		fmt.Fprintln(w, `Code:	'exoscale'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "EXOSCALE_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "EXOSCALE_API_SECRET":	API secret`)
		fmt.Fprintln(w, `	- "EXOSCALE_ENDPOINT":	API endpoint URL`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "EXOSCALE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "EXOSCALE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "EXOSCALE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "EXOSCALE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/exoscale`)

	case "fastdns":
		// generated from: providers/dns/fastdns/fastdns.toml
		fmt.Fprintln(w, `Configuration for FastDNS.`)
		fmt.Fprintln(w, `Code:	'fastdns'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "AKAMAI_ACCESS_TOKEN":	Access token`)
		fmt.Fprintln(w, `	- "AKAMAI_CLIENT_SECRET":	Client secret`)
		fmt.Fprintln(w, `	- "AKAMAI_CLIENT_TOKEN":	Client token`)
		fmt.Fprintln(w, `	- "AKAMAI_HOST":	API host`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "AKAMAI_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "AKAMAI_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "AKAMAI_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/fastdns`)

	case "gandi":
		// generated from: providers/dns/gandi/gandi.toml
		fmt.Fprintln(w, `Configuration for Gandi.`)
		fmt.Fprintln(w, `Code:	'gandi'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "GANDI_API_KEY":	API key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "GANDI_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "GANDI_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "GANDI_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "GANDI_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/gandi`)

	case "gandiv5":
		// generated from: providers/dns/gandiv5/gandiv5.toml
		fmt.Fprintln(w, `Configuration for Gandi Live DNS (v5).`)
		fmt.Fprintln(w, `Code:	'gandiv5'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "GANDIV5_API_KEY":	API key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "GANDIV5_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "GANDIV5_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "GANDIV5_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "GANDIV5_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/gandiv5`)

	case "gcloud":
		// generated from: providers/dns/gcloud/gcloud.toml
		fmt.Fprintln(w, `Configuration for Google Cloud.`)
		fmt.Fprintln(w, `Code:	'gcloud'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "Application Default Credentials":	[Documentation](https://cloud.google.com/docs/authentication/production#providing_credentials_to_your_application)`)
		fmt.Fprintln(w, `	- "GCE_PROJECT":	Project name`)
		fmt.Fprintln(w, `	- "GCE_SERVICE_ACCOUNT":	Account`)
		fmt.Fprintln(w, `	- "GCE_SERVICE_ACCOUNT_FILE":	Account file path`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "GCE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "GCE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "GCE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/gcloud`)

	case "glesys":
		// generated from: providers/dns/glesys/glesys.toml
		fmt.Fprintln(w, `Configuration for Glesys.`)
		fmt.Fprintln(w, `Code:	'glesys'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "GLESYS_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "GLESYS_API_USER":	API user`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "GLESYS_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "GLESYS_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "GLESYS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "GLESYS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/glesys`)

	case "godaddy":
		// generated from: providers/dns/godaddy/godaddy.toml
		fmt.Fprintln(w, `Configuration for Go Daddy.`)
		fmt.Fprintln(w, `Code:	'godaddy'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "GODADDY_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "GODADDY_API_SECRET":	API secret`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "GODADDY_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "GODADDY_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "GODADDY_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "GODADDY_SEQUENCE_INTERVAL":	Interval between iteration`)
		fmt.Fprintln(w, `	- "GODADDY_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/godaddy`)

	case "hostingde":
		// generated from: providers/dns/hostingde/hostingde.toml
		fmt.Fprintln(w, `Configuration for Hosting.de.`)
		fmt.Fprintln(w, `Code:	'hostingde'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "HOSTINGDE_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "HOSTINGDE_ZONE_NAME":	Zone name in ACE format`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "HOSTINGDE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "HOSTINGDE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "HOSTINGDE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "HOSTINGDE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/hostingde`)

	case "httpreq":
		// generated from: providers/dns/httpreq/httpreq.toml
		fmt.Fprintln(w, `Configuration for HTTP request.`)
		fmt.Fprintln(w, `Code:	'httpreq'`)
		fmt.Fprintln(w, `Since:	'v2.0.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "HTTPREQ_ENDPOINT":	The URL of the server`)
		fmt.Fprintln(w, `	- "HTTPREQ_MODE":	'RAW', none`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "HTTPREQ_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "HTTPREQ_PASSWORD":	Basic authentication password`)
		fmt.Fprintln(w, `	- "HTTPREQ_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "HTTPREQ_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "HTTPREQ_USERNAME":	Basic authentication username`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/httpreq`)

	case "iij":
		// generated from: providers/dns/iij/iij.toml
		fmt.Fprintln(w, `Configuration for Internet Initiative Japan.`)
		fmt.Fprintln(w, `Code:	'iij'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "IIJ_API_ACCESS_KEY":	API access key`)
		fmt.Fprintln(w, `	- "IIJ_API_SECRET_KEY":	API secret key`)
		fmt.Fprintln(w, `	- "IIJ_DO_SERVICE_CODE":	DO service code`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "IIJ_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "IIJ_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "IIJ_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/iij`)

	case "inwx":
		// generated from: providers/dns/inwx/inwx.toml
		fmt.Fprintln(w, `Configuration for INWX.`)
		fmt.Fprintln(w, `Code:	'inwx'`)
		fmt.Fprintln(w, `Since:	'v2.0.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "INWX_PASSWORD":	Password`)
		fmt.Fprintln(w, `	- "INWX_USERNAME":	Username`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "INWX_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "INWX_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "INWX_SANDBOX":	Activate the sandbox (boolean)`)
		fmt.Fprintln(w, `	- "INWX_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/inwx`)

	case "joker":
		// generated from: providers/dns/joker/joker.toml
		fmt.Fprintln(w, `Configuration for Joker.`)
		fmt.Fprintln(w, `Code:	'joker'`)
		fmt.Fprintln(w, `Since:	'v2.6.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "JOKER_API_KEY":	API key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "JOKER_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "JOKER_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "JOKER_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "JOKER_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/joker`)

	case "lightsail":
		// generated from: providers/dns/lightsail/lightsail.toml
		fmt.Fprintln(w, `Configuration for Amazon Lightsail.`)
		fmt.Fprintln(w, `Code:	'lightsail'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "AWS_ACCESS_KEY_ID":	Access key ID`)
		fmt.Fprintln(w, `	- "AWS_SECRET_ACCESS_KEY":	Secret access key`)
		fmt.Fprintln(w, `	- "DNS_ZONE":	DNS zone`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "LIGHTSAIL_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "LIGHTSAIL_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/lightsail`)

	case "linode":
		// generated from: providers/dns/linode/linode.toml
		fmt.Fprintln(w, `Configuration for Linode (deprecated).`)
		fmt.Fprintln(w, `Code:	'linode'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "LINODE_API_KEY":	API key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "LINODE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "LINODE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "LINODE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/linode`)

	case "linodev4":
		// generated from: providers/dns/linodev4/linodev4.toml
		fmt.Fprintln(w, `Configuration for Linode (v4).`)
		fmt.Fprintln(w, `Code:	'linodev4'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "LINODE_TOKEN":	API token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "LINODE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "LINODE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "LINODE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/linodev4`)

	case "mydnsjp":
		// generated from: providers/dns/mydnsjp/mydnsjp.toml
		fmt.Fprintln(w, `Configuration for MyDNS.jp.`)
		fmt.Fprintln(w, `Code:	'mydnsjp'`)
		fmt.Fprintln(w, `Since:	'v1.2.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "MYDNSJP_MASTER_ID":	Master ID`)
		fmt.Fprintln(w, `	- "MYDNSJP_PASSWORD":	Password`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "MYDNSJP_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "MYDNSJP_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "MYDNSJP_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "MYDNSJP_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/mydnsjp`)

	case "namecheap":
		// generated from: providers/dns/namecheap/namecheap.toml
		fmt.Fprintln(w, `Configuration for Namecheap.`)
		fmt.Fprintln(w, `Code:	'namecheap'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "NAMECHEAP_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "NAMECHEAP_API_USER":	API user`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "NAMECHEAP_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "NAMECHEAP_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "NAMECHEAP_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "NAMECHEAP_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/namecheap`)

	case "namedotcom":
		// generated from: providers/dns/namedotcom/namedotcom.toml
		fmt.Fprintln(w, `Configuration for Name.com.`)
		fmt.Fprintln(w, `Code:	'namedotcom'`)
		fmt.Fprintln(w, `Since:	'v0.5.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "NAMECOM_API_TOKEN":	API token`)
		fmt.Fprintln(w, `	- "NAMECOM_USERNAME":	Username`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "NAMECOM_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "NAMECOM_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "NAMECOM_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "NAMECOM_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/namedotcom`)

	case "netcup":
		// generated from: providers/dns/netcup/netcup.toml
		fmt.Fprintln(w, `Configuration for Netcup.`)
		fmt.Fprintln(w, `Code:	'netcup'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "NETCUP_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "NETCUP_API_PASSWORD":	API password`)
		fmt.Fprintln(w, `	- "NETCUP_CUSTOMER_NUMBER":	Customer number`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "NETCUP_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "NETCUP_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "NETCUP_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "NETCUP_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/netcup`)

	case "nifcloud":
		// generated from: providers/dns/nifcloud/nifcloud.toml
		fmt.Fprintln(w, `Configuration for NIFCloud.`)
		fmt.Fprintln(w, `Code:	'nifcloud'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "NIFCLOUD_ACCESS_KEY_ID":	Access key`)
		fmt.Fprintln(w, `	- "NIFCLOUD_SECRET_ACCESS_KEY":	Secret access key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "NIFCLOUD_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "NIFCLOUD_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "NIFCLOUD_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "NIFCLOUD_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/nifcloud`)

	case "ns1":
		// generated from: providers/dns/ns1/ns1.toml
		fmt.Fprintln(w, `Configuration for NS1.`)
		fmt.Fprintln(w, `Code:	'ns1'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "NS1_API_KEY":	API key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "NS1_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "NS1_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "NS1_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "NS1_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/ns1`)

	case "oraclecloud":
		// generated from: providers/dns/oraclecloud/oraclecloud.toml
		fmt.Fprintln(w, `Configuration for Oracle Cloud.`)
		fmt.Fprintln(w, `Code:	'oraclecloud'`)
		fmt.Fprintln(w, `Since:	'v2.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "OCI_COMPARTMENT_OCID":	Compartment OCID`)
		fmt.Fprintln(w, `	- "OCI_PRIVKEY_FILE":	Private key file`)
		fmt.Fprintln(w, `	- "OCI_PRIVKEY_PASS":	Private key password`)
		fmt.Fprintln(w, `	- "OCI_PUBKEY_FINGERPRINT":	Public key fingerprint`)
		fmt.Fprintln(w, `	- "OCI_REGION":	Region`)
		fmt.Fprintln(w, `	- "OCI_TENANCY_OCID":	Tenanct OCID`)
		fmt.Fprintln(w, `	- "OCI_USER_OCID":	User OCID`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "OCI_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "OCI_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "OCI_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/oraclecloud`)

	case "otc":
		// generated from: providers/dns/otc/otc.toml
		fmt.Fprintln(w, `Configuration for Open Telekom Cloud.`)
		fmt.Fprintln(w, `Code:	'otc'`)
		fmt.Fprintln(w, `Since:	'v0.4.1'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "OTC_DOMAIN_NAME":	Domain name`)
		fmt.Fprintln(w, `	- "OTC_IDENTITY_ENDPOINT":	Identity endpoint URL`)
		fmt.Fprintln(w, `	- "OTC_PASSWORD":	Password`)
		fmt.Fprintln(w, `	- "OTC_PROJECT_NAME":	Project name`)
		fmt.Fprintln(w, `	- "OTC_USER_NAME":	User name`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "OTC_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "OTC_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "OTC_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "OTC_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/otc`)

	case "ovh":
		// generated from: providers/dns/ovh/ovh.toml
		fmt.Fprintln(w, `Configuration for OVH.`)
		fmt.Fprintln(w, `Code:	'ovh'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "OVH_APPLICATION_KEY":	Application key`)
		fmt.Fprintln(w, `	- "OVH_APPLICATION_SECRET":	Application secret`)
		fmt.Fprintln(w, `	- "OVH_CONSUMER_KEY":	Consumer key`)
		fmt.Fprintln(w, `	- "OVH_ENDPOINT":	Endpoint URL (ovh-eu or ovh-ca)`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "OVH_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "OVH_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "OVH_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "OVH_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/ovh`)

	case "pdns":
		// generated from: providers/dns/pdns/pdns.toml
		fmt.Fprintln(w, `Configuration for PowerDNS.`)
		fmt.Fprintln(w, `Code:	'pdns'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "PDNS_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "PDNS_API_URL":	API url`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "PDNS_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "PDNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "PDNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "PDNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/pdns`)

	case "rackspace":
		// generated from: providers/dns/rackspace/rackspace.toml
		fmt.Fprintln(w, `Configuration for Rackspace.`)
		fmt.Fprintln(w, `Code:	'rackspace'`)
		fmt.Fprintln(w, `Since:	'v0.4.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "RACKSPACE_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "RACKSPACE_USER":	API user`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "RACKSPACE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "RACKSPACE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "RACKSPACE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "RACKSPACE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/rackspace`)

	case "rfc2136":
		// generated from: providers/dns/rfc2136/rfc2136.toml
		fmt.Fprintln(w, `Configuration for RFC2136.`)
		fmt.Fprintln(w, `Code:	'rfc2136'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "RFC2136_NAMESERVER":	Network address in the form "host" or "host:port"`)
		fmt.Fprintln(w, `	- "RFC2136_TSIG_ALGORITHM":	TSIG algorythm. See [miekg/dns#tsig.go](https://github.com/miekg/dns/blob/master/tsig.go) for supported values. To disable TSIG authentication, leave the 'RFC2136_TSIG*' variables unset.`)
		fmt.Fprintln(w, `	- "RFC2136_TSIG_KEY":	Name of the secret key as defined in DNS server configuration. To disable TSIG authentication, leave the 'RFC2136_TSIG*' variables unset.`)
		fmt.Fprintln(w, `	- "RFC2136_TSIG_SECRET":	Secret key payload. To disable TSIG authentication, leave the' RFC2136_TSIG*' variables unset.`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "RFC2136_DNS_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "RFC2136_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "RFC2136_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "RFC2136_SEQUENCE_INTERVAL":	Interval between iteration`)
		fmt.Fprintln(w, `	- "RFC2136_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/rfc2136`)

	case "route53":
		// generated from: providers/dns/route53/route53.toml
		fmt.Fprintln(w, `Configuration for Amazon Route 53.`)
		fmt.Fprintln(w, `Code:	'route53'`)
		fmt.Fprintln(w, `Since:	'v0.3.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "AWS_ACCESS_KEY_ID":	Managed by the AWS client`)
		fmt.Fprintln(w, `	- "AWS_HOSTED_ZONE_ID":	Override the hosted zone ID`)
		fmt.Fprintln(w, `	- "AWS_REGION":	Managed by the AWS client`)
		fmt.Fprintln(w, `	- "AWS_SECRET_ACCESS_KEY":	Managed by the AWS client`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "AWS_MAX_RETRIES":	The number of maximum returns the service will use to make an individual API request`)
		fmt.Fprintln(w, `	- "AWS_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "AWS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "AWS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/route53`)

	case "sakuracloud":
		// generated from: providers/dns/sakuracloud/sakuracloud.toml
		fmt.Fprintln(w, `Configuration for Sakura Cloud.`)
		fmt.Fprintln(w, `Code:	'sakuracloud'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "SAKURACLOUD_ACCESS_TOKEN":	Access token`)
		fmt.Fprintln(w, `	- "SAKURACLOUD_ACCESS_TOKEN_SECRET":	Access token secret`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "SAKURACLOUD_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "SAKURACLOUD_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "SAKURACLOUD_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "SAKURACLOUD_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/sakuracloud`)

	case "selectel":
		// generated from: providers/dns/selectel/selectel.toml
		fmt.Fprintln(w, `Configuration for Selectel.`)
		fmt.Fprintln(w, `Code:	'selectel'`)
		fmt.Fprintln(w, `Since:	'v1.2.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "SELECTEL_API_TOKEN":	API token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "SELECTEL_BASE_URL":	API endpoint URL`)
		fmt.Fprintln(w, `	- "SELECTEL_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "SELECTEL_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "SELECTEL_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "SELECTEL_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/selectel`)

	case "stackpath":
		// generated from: providers/dns/stackpath/stackpath.toml
		fmt.Fprintln(w, `Configuration for Stackpath.`)
		fmt.Fprintln(w, `Code:	'stackpath'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "STACKPATH_CLIENT_ID":	Client ID`)
		fmt.Fprintln(w, `	- "STACKPATH_CLIENT_SECRET":	Client secret`)
		fmt.Fprintln(w, `	- "STACKPATH_STACK_ID":	Stack ID`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "STACKPATH_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "STACKPATH_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "STACKPATH_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/stackpath`)

	case "transip":
		// generated from: providers/dns/transip/transip.toml
		fmt.Fprintln(w, `Configuration for TransIP.`)
		fmt.Fprintln(w, `Code:	'transip'`)
		fmt.Fprintln(w, `Since:	'v2.0.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "TRANSIP_ACCOUNT_NAME":	Account name`)
		fmt.Fprintln(w, `	- "TRANSIP_PRIVATE_KEY_PATH":	Private key path`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "TRANSIP_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "TRANSIP_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "TRANSIP_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/transip`)

	case "vegadns":
		// generated from: providers/dns/vegadns/vegadns.toml
		fmt.Fprintln(w, `Configuration for VegaDNS.`)
		fmt.Fprintln(w, `Code:	'vegadns'`)
		fmt.Fprintln(w, `Since:	'v1.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "SECRET_VEGADNS_KEY":	API key`)
		fmt.Fprintln(w, `	- "SECRET_VEGADNS_SECRET":	API secret`)
		fmt.Fprintln(w, `	- "VEGADNS_URL":	API endpoint URL`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "VEGADNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "VEGADNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "VEGADNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/vegadns`)

	case "vscale":
		// generated from: providers/dns/vscale/vscale.toml
		fmt.Fprintln(w, `Configuration for Vscale.`)
		fmt.Fprintln(w, `Code:	'vscale'`)
		fmt.Fprintln(w, `Since:	'v2.0.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "VSCALE_API_TOKEN":	API token`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "VSCALE_BASE_URL":	API enddpoint URL`)
		fmt.Fprintln(w, `	- "VSCALE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "VSCALE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "VSCALE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "VSCALE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/vscale`)

	case "vultr":
		// generated from: providers/dns/vultr/vultr.toml
		fmt.Fprintln(w, `Configuration for Vultr.`)
		fmt.Fprintln(w, `Code:	'vultr'`)
		fmt.Fprintln(w, `Since:	'v0.3.1'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "VULTR_API_KEY":	API key`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "VULTR_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "VULTR_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "VULTR_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "VULTR_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/vultr`)

	case "zoneee":
		// generated from: providers/dns/zoneee/zoneee.toml
		fmt.Fprintln(w, `Configuration for Zone.ee.`)
		fmt.Fprintln(w, `Code:	'zoneee'`)
		fmt.Fprintln(w, `Since:	'v2.1.0'`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Credentials:`)
		fmt.Fprintln(w, `	- "ZONEEE_API_KEY":	API key`)
		fmt.Fprintln(w, `	- "ZONEEE_API_USER":	API user`)
		fmt.Fprintln(w)

		fmt.Fprintln(w, `Additional Configuration:`)
		fmt.Fprintln(w, `	- "ZONEEE_ENDPOINT":	API endpoint URL`)
		fmt.Fprintln(w, `	- "ZONEEE_HTTP_TIMEOUT":	API request timeout`)
		fmt.Fprintln(w, `	- "ZONEEE_POLLING_INTERVAL":	Time between DNS propagation check`)
		fmt.Fprintln(w, `	- "ZONEEE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		fmt.Fprintln(w, `	- "ZONEEE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		fmt.Fprintln(w)
		fmt.Fprintln(w, `More information: https://go-acme.github.io/lego/dns/zoneee`)

	case "manual":
		fmt.Fprintln(w, `Solving the DNS-01 challenge using CLI prompt.`)
	default:
		log.Fatalf("%q is not yet supported.", name)
	}
	w.Flush()
}
