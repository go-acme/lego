package cmd

// CODE GENERATED AUTOMATICALLY
// THIS FILE MUST NOT BE EDITED BY HAND

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

func allDNSCodes() string {
	providers := []string{
		"manual",
		"acme-dns",
		"alidns",
		"auroradns",
		"autodns",
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
		"liquidweb",
		"mydnsjp",
		"namecheap",
		"namedotcom",
		"namesilo",
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
		"versio",
		"vscale",
		"vultr",
		"zoneee",
	}
	sort.Strings(providers)
	return strings.Join(providers, ", ")
}

func displayDNSHelp(name string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	ew := &errWriter{w: w}

	switch name {
	case "acme-dns":
		// generated from: providers/dns/acmedns/acmedns.toml
		ew.writeln(`Configuration for Joohoi's ACME-DNS.`)
		ew.writeln(`Code:	'acme-dns'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "ACME_DNS_API_BASE":	The ACME-DNS API address`)
		ew.writeln(`	- "ACME_DNS_STORAGE_PATH":	The ACME-DNS JSON account data file. A per-domain account will be registered/persisted to this file and used for TXT updates.`)
		ew.writeln()

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/acme-dns`)

	case "alidns":
		// generated from: providers/dns/alidns/alidns.toml
		ew.writeln(`Configuration for Alibaba Cloud DNS.`)
		ew.writeln(`Code:	'alidns'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "ALICLOUD_ACCESS_KEY":	Access key ID`)
		ew.writeln(`	- "ALICLOUD_SECRET_KEY":	Access Key secret`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "ALICLOUD_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "ALICLOUD_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "ALICLOUD_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "ALICLOUD_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/alidns`)

	case "auroradns":
		// generated from: providers/dns/auroradns/auroradns.toml
		ew.writeln(`Configuration for Aurora DNS.`)
		ew.writeln(`Code:	'auroradns'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "AURORA_ENDPOINT":	API endpoint URL`)
		ew.writeln(`	- "AURORA_KEY":	User API key`)
		ew.writeln(`	- "AURORA_USER_ID":	User ID`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "AURORA_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "AURORA_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "AURORA_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/auroradns`)

	case "autodns":
		// generated from: providers/dns/autodns/autodns.toml
		ew.writeln(`Configuration for Autodns.`)
		ew.writeln(`Code:	'autodns'`)
		ew.writeln(`Since:	'v3.2.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "AUTODNS_API_PASSWORD":	User Password`)
		ew.writeln(`	- "AUTODNS_API_USER":	Username`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "AUTODNS_CONTEXT":	API context (4 for production, 1 for testing. Defaults to 4)`)
		ew.writeln(`	- "AUTODNS_ENDPOINT":	API endpoint URL, defaults to https://api.autodns.com/v1/`)
		ew.writeln(`	- "AUTODNS_HTTP_TIMEOUT":	API request timeout, defaults to 30 seconds`)
		ew.writeln(`	- "AUTODNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "AUTODNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "AUTODNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/autodns`)

	case "azure":
		// generated from: providers/dns/azure/azure.toml
		ew.writeln(`Configuration for Azure.`)
		ew.writeln(`Code:	'azure'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "AZURE_CLIENT_ID":	Client ID`)
		ew.writeln(`	- "AZURE_CLIENT_SECRET":	Client secret`)
		ew.writeln(`	- "AZURE_RESOURCE_GROUP":	Resource group`)
		ew.writeln(`	- "AZURE_SUBSCRIPTION_ID":	Subscription ID`)
		ew.writeln(`	- "AZURE_TENANT_ID":	Tenant ID`)
		ew.writeln(`	- "instance metadata service":	If the credentials are **not** set via the environment, then it will attempt to get a bearer token via the [instance metadata service](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service).`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "AZURE_METADATA_ENDPOINT":	Metadata Service endpoint URL`)
		ew.writeln(`	- "AZURE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "AZURE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "AZURE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/azure`)

	case "bindman":
		// generated from: providers/dns/bindman/bindman.toml
		ew.writeln(`Configuration for Bindman.`)
		ew.writeln(`Code:	'bindman'`)
		ew.writeln(`Since:	'v2.6.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "BINDMAN_MANAGER_ADDRESS":	The server URL, should have scheme, hostname, and port (if required) of the Bindman-DNS Manager server`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "BINDMAN_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "BINDMAN_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "BINDMAN_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/bindman`)

	case "bluecat":
		// generated from: providers/dns/bluecat/bluecat.toml
		ew.writeln(`Configuration for Bluecat.`)
		ew.writeln(`Code:	'bluecat'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "BLUECAT_CONFIG_NAME":	Configuration name`)
		ew.writeln(`	- "BLUECAT_DNS_VIEW":	External DNS View Name`)
		ew.writeln(`	- "BLUECAT_PASSWORD":	API password`)
		ew.writeln(`	- "BLUECAT_SERVER_URL":	The server URL, should have scheme, hostname, and port (if required) of the authoritative Bluecat BAM serve`)
		ew.writeln(`	- "BLUECAT_USER_NAME":	API username`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "BLUECAT_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "BLUECAT_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "BLUECAT_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "BLUECAT_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/bluecat`)

	case "cloudflare":
		// generated from: providers/dns/cloudflare/cloudflare.toml
		ew.writeln(`Configuration for Cloudflare.`)
		ew.writeln(`Code:	'cloudflare'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "CF_API_EMAIL":	Account email`)
		ew.writeln(`	- "CF_API_KEY":	API key`)
		ew.writeln(`	- "CF_DNS_API_TOKEN":	API token with DNS:Edit permission (since v3.1.0)`)
		ew.writeln(`	- "CF_ZONE_API_TOKEN":	API token with Zone:Read permission (since v3.1.0)`)
		ew.writeln(`	- "CLOUDFLARE_API_KEY":	Alias to CF_API_KEY`)
		ew.writeln(`	- "CLOUDFLARE_DNS_API_TOKEN":	Alias to CF_DNS_API_TOKEN`)
		ew.writeln(`	- "CLOUDFLARE_EMAIL":	Alias to CF_API_EMAIL`)
		ew.writeln(`	- "CLOUDFLARE_ZONE_API_TOKEN":	Alias to CF_ZONE_API_TOKEN`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "CLOUDFLARE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "CLOUDFLARE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "CLOUDFLARE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "CLOUDFLARE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/cloudflare`)

	case "cloudns":
		// generated from: providers/dns/cloudns/cloudns.toml
		ew.writeln(`Configuration for ClouDNS.`)
		ew.writeln(`Code:	'cloudns'`)
		ew.writeln(`Since:	'v2.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "CLOUDNS_AUTH_ID":	The API user ID`)
		ew.writeln(`	- "CLOUDNS_AUTH_PASSWORD":	The password for API user ID`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "CLOUDNS_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "CLOUDNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "CLOUDNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "CLOUDNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/cloudns`)

	case "cloudxns":
		// generated from: providers/dns/cloudxns/cloudxns.toml
		ew.writeln(`Configuration for CloudXNS.`)
		ew.writeln(`Code:	'cloudxns'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "CLOUDXNS_API_KEY":	The API key`)
		ew.writeln(`	- "CLOUDXNS_SECRET_KEY":	THe API secret key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "CLOUDXNS_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "CLOUDXNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "CLOUDXNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "CLOUDXNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/cloudxns`)

	case "conoha":
		// generated from: providers/dns/conoha/conoha.toml
		ew.writeln(`Configuration for ConoHa.`)
		ew.writeln(`Code:	'conoha'`)
		ew.writeln(`Since:	'v1.2.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "CONOHA_API_PASSWORD":	The API password`)
		ew.writeln(`	- "CONOHA_API_USERNAME":	The API username`)
		ew.writeln(`	- "CONOHA_TENANT_ID":	Tenant ID`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "CONOHA_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "CONOHA_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "CONOHA_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "CONOHA_REGION":	The region`)
		ew.writeln(`	- "CONOHA_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/conoha`)

	case "designate":
		// generated from: providers/dns/designate/designate.toml
		ew.writeln(`Configuration for Designate DNSaaS for Openstack.`)
		ew.writeln(`Code:	'designate'`)
		ew.writeln(`Since:	'v2.2.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "OS_AUTH_URL":	Identity endpoint URL`)
		ew.writeln(`	- "OS_PASSWORD":	Password`)
		ew.writeln(`	- "OS_PROJECT_NAME":	Project name`)
		ew.writeln(`	- "OS_REGION_NAME":	Region name`)
		ew.writeln(`	- "OS_TENANT_NAME":	Tenant name (deprecated see OS_PROJECT_NAME and OS_PROJECT_ID)`)
		ew.writeln(`	- "OS_USERNAME":	Username`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DESIGNATE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DESIGNATE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DESIGNATE_TTL":	The TTL of the TXT record used for the DNS challenge`)
		ew.writeln(`	- "OS_PROJECT_ID":	Project ID`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/designate`)

	case "digitalocean":
		// generated from: providers/dns/digitalocean/digitalocean.toml
		ew.writeln(`Configuration for Digital Ocean.`)
		ew.writeln(`Code:	'digitalocean'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "DO_AUTH_TOKEN":	Authentication token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DO_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "DO_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DO_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DO_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/digitalocean`)

	case "dnsimple":
		// generated from: providers/dns/dnsimple/dnsimple.toml
		ew.writeln(`Configuration for DNSimple.`)
		ew.writeln(`Code:	'dnsimple'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "DNSIMPLE_BASE_URL":	API endpoint URL`)
		ew.writeln(`	- "DNSIMPLE_OAUTH_TOKEN":	OAuth token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DNSIMPLE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DNSIMPLE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DNSIMPLE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/dnsimple`)

	case "dnsmadeeasy":
		// generated from: providers/dns/dnsmadeeasy/dnsmadeeasy.toml
		ew.writeln(`Configuration for DNS Made Easy.`)
		ew.writeln(`Code:	'dnsmadeeasy'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "DNSMADEEASY_API_KEY":	The API key`)
		ew.writeln(`	- "DNSMADEEASY_API_SECRET":	The API Secret key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DNSMADEEASY_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "DNSMADEEASY_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DNSMADEEASY_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DNSMADEEASY_SANDBOX":	Activate the sandbox (boolean)`)
		ew.writeln(`	- "DNSMADEEASY_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/dnsmadeeasy`)

	case "dnspod":
		// generated from: providers/dns/dnspod/dnspod.toml
		ew.writeln(`Configuration for DNSPod.`)
		ew.writeln(`Code:	'dnspod'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "DNSPOD_API_KEY":	The user token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DNSPOD_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "DNSPOD_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DNSPOD_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DNSPOD_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/dnspod`)

	case "dode":
		// generated from: providers/dns/dode/dode.toml
		ew.writeln(`Configuration for Domain Offensive (do.de).`)
		ew.writeln(`Code:	'dode'`)
		ew.writeln(`Since:	'v2.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "DODE_TOKEN":	API token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DODE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "DODE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DODE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DODE_SEQUENCE_INTERVAL":	Interval between iteration`)
		ew.writeln(`	- "DODE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/dode`)

	case "dreamhost":
		// generated from: providers/dns/dreamhost/dreamhost.toml
		ew.writeln(`Configuration for DreamHost.`)
		ew.writeln(`Code:	'dreamhost'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "DREAMHOST_API_KEY":	The API key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DREAMHOST_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "DREAMHOST_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DREAMHOST_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DREAMHOST_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/dreamhost`)

	case "duckdns":
		// generated from: providers/dns/duckdns/duckdns.toml
		ew.writeln(`Configuration for Duck DNS.`)
		ew.writeln(`Code:	'duckdns'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "DUCKDNS_TOKEN":	Account token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DUCKDNS_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "DUCKDNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DUCKDNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DUCKDNS_SEQUENCE_INTERVAL":	Interval between iteration`)
		ew.writeln(`	- "DUCKDNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/duckdns`)

	case "dyn":
		// generated from: providers/dns/dyn/dyn.toml
		ew.writeln(`Configuration for Dyn.`)
		ew.writeln(`Code:	'dyn'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "DYN_CUSTOMER_NAME":	Customer name`)
		ew.writeln(`	- "DYN_PASSWORD":	Paswword`)
		ew.writeln(`	- "DYN_USER_NAME":	User name`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "DYN_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "DYN_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "DYN_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "DYN_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/dyn`)

	case "easydns":
		// generated from: providers/dns/easydns/easydns.toml
		ew.writeln(`Configuration for EasyDNS.`)
		ew.writeln(`Code:	'easydns'`)
		ew.writeln(`Since:	'v2.6.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "EASYDNS_KEY":	API Key`)
		ew.writeln(`	- "EASYDNS_TOKEN":	API Token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "EASYDNS_ENDPOINT":	The endpoint URL of the API Server`)
		ew.writeln(`	- "EASYDNS_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "EASYDNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "EASYDNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "EASYDNS_SEQUENCE_INTERVAL":	Time between sequential requests`)
		ew.writeln(`	- "EASYDNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/easydns`)

	case "exec":
		// generated from: providers/dns/exec/exec.toml
		ew.writeln(`Configuration for External program.`)
		ew.writeln(`Code:	'exec'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/exec`)

	case "exoscale":
		// generated from: providers/dns/exoscale/exoscale.toml
		ew.writeln(`Configuration for Exoscale.`)
		ew.writeln(`Code:	'exoscale'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "EXOSCALE_API_KEY":	API key`)
		ew.writeln(`	- "EXOSCALE_API_SECRET":	API secret`)
		ew.writeln(`	- "EXOSCALE_ENDPOINT":	API endpoint URL`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "EXOSCALE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "EXOSCALE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "EXOSCALE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "EXOSCALE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/exoscale`)

	case "fastdns":
		// generated from: providers/dns/fastdns/fastdns.toml
		ew.writeln(`Configuration for FastDNS.`)
		ew.writeln(`Code:	'fastdns'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "AKAMAI_ACCESS_TOKEN":	Access token`)
		ew.writeln(`	- "AKAMAI_CLIENT_SECRET":	Client secret`)
		ew.writeln(`	- "AKAMAI_CLIENT_TOKEN":	Client token`)
		ew.writeln(`	- "AKAMAI_HOST":	API host`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "AKAMAI_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "AKAMAI_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "AKAMAI_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/fastdns`)

	case "gandi":
		// generated from: providers/dns/gandi/gandi.toml
		ew.writeln(`Configuration for Gandi.`)
		ew.writeln(`Code:	'gandi'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "GANDI_API_KEY":	API key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "GANDI_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "GANDI_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "GANDI_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "GANDI_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/gandi`)

	case "gandiv5":
		// generated from: providers/dns/gandiv5/gandiv5.toml
		ew.writeln(`Configuration for Gandi Live DNS (v5).`)
		ew.writeln(`Code:	'gandiv5'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "GANDIV5_API_KEY":	API key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "GANDIV5_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "GANDIV5_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "GANDIV5_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "GANDIV5_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/gandiv5`)

	case "gcloud":
		// generated from: providers/dns/gcloud/gcloud.toml
		ew.writeln(`Configuration for Google Cloud.`)
		ew.writeln(`Code:	'gcloud'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "Application Default Credentials":	[Documentation](https://cloud.google.com/docs/authentication/production#providing_credentials_to_your_application)`)
		ew.writeln(`	- "GCE_PROJECT":	Project name`)
		ew.writeln(`	- "GCE_SERVICE_ACCOUNT":	Account`)
		ew.writeln(`	- "GCE_SERVICE_ACCOUNT_FILE":	Account file path`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "GCE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "GCE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "GCE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/gcloud`)

	case "glesys":
		// generated from: providers/dns/glesys/glesys.toml
		ew.writeln(`Configuration for Glesys.`)
		ew.writeln(`Code:	'glesys'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "GLESYS_API_KEY":	API key`)
		ew.writeln(`	- "GLESYS_API_USER":	API user`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "GLESYS_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "GLESYS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "GLESYS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "GLESYS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/glesys`)

	case "godaddy":
		// generated from: providers/dns/godaddy/godaddy.toml
		ew.writeln(`Configuration for Go Daddy.`)
		ew.writeln(`Code:	'godaddy'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "GODADDY_API_KEY":	API key`)
		ew.writeln(`	- "GODADDY_API_SECRET":	API secret`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "GODADDY_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "GODADDY_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "GODADDY_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "GODADDY_SEQUENCE_INTERVAL":	Interval between iteration`)
		ew.writeln(`	- "GODADDY_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/godaddy`)

	case "hostingde":
		// generated from: providers/dns/hostingde/hostingde.toml
		ew.writeln(`Configuration for Hosting.de.`)
		ew.writeln(`Code:	'hostingde'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "HOSTINGDE_API_KEY":	API key`)
		ew.writeln(`	- "HOSTINGDE_ZONE_NAME":	Zone name in ACE format`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "HOSTINGDE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "HOSTINGDE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "HOSTINGDE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "HOSTINGDE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/hostingde`)

	case "httpreq":
		// generated from: providers/dns/httpreq/httpreq.toml
		ew.writeln(`Configuration for HTTP request.`)
		ew.writeln(`Code:	'httpreq'`)
		ew.writeln(`Since:	'v2.0.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "HTTPREQ_ENDPOINT":	The URL of the server`)
		ew.writeln(`	- "HTTPREQ_MODE":	'RAW', none`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "HTTPREQ_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "HTTPREQ_PASSWORD":	Basic authentication password`)
		ew.writeln(`	- "HTTPREQ_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "HTTPREQ_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "HTTPREQ_USERNAME":	Basic authentication username`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/httpreq`)

	case "iij":
		// generated from: providers/dns/iij/iij.toml
		ew.writeln(`Configuration for Internet Initiative Japan.`)
		ew.writeln(`Code:	'iij'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "IIJ_API_ACCESS_KEY":	API access key`)
		ew.writeln(`	- "IIJ_API_SECRET_KEY":	API secret key`)
		ew.writeln(`	- "IIJ_DO_SERVICE_CODE":	DO service code`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "IIJ_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "IIJ_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "IIJ_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/iij`)

	case "inwx":
		// generated from: providers/dns/inwx/inwx.toml
		ew.writeln(`Configuration for INWX.`)
		ew.writeln(`Code:	'inwx'`)
		ew.writeln(`Since:	'v2.0.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "INWX_PASSWORD":	Password`)
		ew.writeln(`	- "INWX_USERNAME":	Username`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "INWX_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "INWX_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "INWX_SANDBOX":	Activate the sandbox (boolean)`)
		ew.writeln(`	- "INWX_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/inwx`)

	case "joker":
		// generated from: providers/dns/joker/joker.toml
		ew.writeln(`Configuration for Joker.`)
		ew.writeln(`Code:	'joker'`)
		ew.writeln(`Since:	'v2.6.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "JOKER_API_KEY":	API key`)
		ew.writeln(`	- "JOKER_PASSWORD":	Joker.com password`)
		ew.writeln(`	- "JOKER_USERNAME":	Joker.com username (email address)`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "JOKER_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "JOKER_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "JOKER_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "JOKER_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/joker`)

	case "lightsail":
		// generated from: providers/dns/lightsail/lightsail.toml
		ew.writeln(`Configuration for Amazon Lightsail.`)
		ew.writeln(`Code:	'lightsail'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "AWS_ACCESS_KEY_ID":	Access key ID`)
		ew.writeln(`	- "AWS_SECRET_ACCESS_KEY":	Secret access key`)
		ew.writeln(`	- "DNS_ZONE":	DNS zone`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "LIGHTSAIL_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "LIGHTSAIL_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/lightsail`)

	case "linode":
		// generated from: providers/dns/linode/linode.toml
		ew.writeln(`Configuration for Linode (deprecated).`)
		ew.writeln(`Code:	'linode'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "LINODE_API_KEY":	API key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "LINODE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "LINODE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "LINODE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/linode`)

	case "linodev4":
		// generated from: providers/dns/linodev4/linodev4.toml
		ew.writeln(`Configuration for Linode (v4).`)
		ew.writeln(`Code:	'linodev4'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "LINODE_TOKEN":	API token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "LINODE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "LINODE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "LINODE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "LINODE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/linodev4`)

	case "liquidweb":
		// generated from: providers/dns/liquidweb/liquidweb.toml
		ew.writeln(`Configuration for Liquid Web.`)
		ew.writeln(`Code:	'liquidweb'`)
		ew.writeln(`Since:	'v3.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "LIQUID_WEB_PASSWORD":	Storm API Password`)
		ew.writeln(`	- "LIQUID_WEB_USERNAME":	Storm API Username`)
		ew.writeln(`	- "LIQUID_WEB_ZONE":	DNS Zone`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "LIQUID_WEB_HTTP_TIMEOUT":	Maximum waiting time for the DNS records to be created (not verified)`)
		ew.writeln(`	- "LIQUID_WEB_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "LIQUID_WEB_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "LIQUID_WEB_TTL":	The TTL of the TXT record used for the DNS challenge`)
		ew.writeln(`	- "LIQUID_WEB_URL":	Storm API endpoint`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/liquidweb`)

	case "mydnsjp":
		// generated from: providers/dns/mydnsjp/mydnsjp.toml
		ew.writeln(`Configuration for MyDNS.jp.`)
		ew.writeln(`Code:	'mydnsjp'`)
		ew.writeln(`Since:	'v1.2.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "MYDNSJP_MASTER_ID":	Master ID`)
		ew.writeln(`	- "MYDNSJP_PASSWORD":	Password`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "MYDNSJP_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "MYDNSJP_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "MYDNSJP_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "MYDNSJP_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/mydnsjp`)

	case "namecheap":
		// generated from: providers/dns/namecheap/namecheap.toml
		ew.writeln(`Configuration for Namecheap.`)
		ew.writeln(`Code:	'namecheap'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "NAMECHEAP_API_KEY":	API key`)
		ew.writeln(`	- "NAMECHEAP_API_USER":	API user`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "NAMECHEAP_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "NAMECHEAP_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "NAMECHEAP_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "NAMECHEAP_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/namecheap`)

	case "namedotcom":
		// generated from: providers/dns/namedotcom/namedotcom.toml
		ew.writeln(`Configuration for Name.com.`)
		ew.writeln(`Code:	'namedotcom'`)
		ew.writeln(`Since:	'v0.5.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "NAMECOM_API_TOKEN":	API token`)
		ew.writeln(`	- "NAMECOM_USERNAME":	Username`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "NAMECOM_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "NAMECOM_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "NAMECOM_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "NAMECOM_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/namedotcom`)

	case "namesilo":
		// generated from: providers/dns/namesilo/namesilo.toml
		ew.writeln(`Configuration for Namesilo.`)
		ew.writeln(`Code:	'namesilo'`)
		ew.writeln(`Since:	'v2.7.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "NAMESILO_API_KEY":	Client ID`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "NAMESILO_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "NAMESILO_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation, it is better to set larger than 15m`)
		ew.writeln(`	- "NAMESILO_TTL":	The TTL of the TXT record used for the DNS challenge, should be in [3600, 2592000]`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/namesilo`)

	case "netcup":
		// generated from: providers/dns/netcup/netcup.toml
		ew.writeln(`Configuration for Netcup.`)
		ew.writeln(`Code:	'netcup'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "NETCUP_API_KEY":	API key`)
		ew.writeln(`	- "NETCUP_API_PASSWORD":	API password`)
		ew.writeln(`	- "NETCUP_CUSTOMER_NUMBER":	Customer number`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "NETCUP_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "NETCUP_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "NETCUP_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "NETCUP_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/netcup`)

	case "nifcloud":
		// generated from: providers/dns/nifcloud/nifcloud.toml
		ew.writeln(`Configuration for NIFCloud.`)
		ew.writeln(`Code:	'nifcloud'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "NIFCLOUD_ACCESS_KEY_ID":	Access key`)
		ew.writeln(`	- "NIFCLOUD_SECRET_ACCESS_KEY":	Secret access key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "NIFCLOUD_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "NIFCLOUD_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "NIFCLOUD_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "NIFCLOUD_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/nifcloud`)

	case "ns1":
		// generated from: providers/dns/ns1/ns1.toml
		ew.writeln(`Configuration for NS1.`)
		ew.writeln(`Code:	'ns1'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "NS1_API_KEY":	API key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "NS1_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "NS1_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "NS1_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "NS1_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/ns1`)

	case "oraclecloud":
		// generated from: providers/dns/oraclecloud/oraclecloud.toml
		ew.writeln(`Configuration for Oracle Cloud.`)
		ew.writeln(`Code:	'oraclecloud'`)
		ew.writeln(`Since:	'v2.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "OCI_COMPARTMENT_OCID":	Compartment OCID`)
		ew.writeln(`	- "OCI_PRIVKEY_FILE":	Private key file`)
		ew.writeln(`	- "OCI_PRIVKEY_PASS":	Private key password`)
		ew.writeln(`	- "OCI_PUBKEY_FINGERPRINT":	Public key fingerprint`)
		ew.writeln(`	- "OCI_REGION":	Region`)
		ew.writeln(`	- "OCI_TENANCY_OCID":	Tenanct OCID`)
		ew.writeln(`	- "OCI_USER_OCID":	User OCID`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "OCI_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "OCI_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "OCI_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/oraclecloud`)

	case "otc":
		// generated from: providers/dns/otc/otc.toml
		ew.writeln(`Configuration for Open Telekom Cloud.`)
		ew.writeln(`Code:	'otc'`)
		ew.writeln(`Since:	'v0.4.1'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "OTC_DOMAIN_NAME":	Domain name`)
		ew.writeln(`	- "OTC_IDENTITY_ENDPOINT":	Identity endpoint URL`)
		ew.writeln(`	- "OTC_PASSWORD":	Password`)
		ew.writeln(`	- "OTC_PROJECT_NAME":	Project name`)
		ew.writeln(`	- "OTC_USER_NAME":	User name`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "OTC_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "OTC_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "OTC_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "OTC_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/otc`)

	case "ovh":
		// generated from: providers/dns/ovh/ovh.toml
		ew.writeln(`Configuration for OVH.`)
		ew.writeln(`Code:	'ovh'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "OVH_APPLICATION_KEY":	Application key`)
		ew.writeln(`	- "OVH_APPLICATION_SECRET":	Application secret`)
		ew.writeln(`	- "OVH_CONSUMER_KEY":	Consumer key`)
		ew.writeln(`	- "OVH_ENDPOINT":	Endpoint URL (ovh-eu or ovh-ca)`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "OVH_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "OVH_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "OVH_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "OVH_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/ovh`)

	case "pdns":
		// generated from: providers/dns/pdns/pdns.toml
		ew.writeln(`Configuration for PowerDNS.`)
		ew.writeln(`Code:	'pdns'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "PDNS_API_KEY":	API key`)
		ew.writeln(`	- "PDNS_API_URL":	API url`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "PDNS_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "PDNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "PDNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "PDNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/pdns`)

	case "rackspace":
		// generated from: providers/dns/rackspace/rackspace.toml
		ew.writeln(`Configuration for Rackspace.`)
		ew.writeln(`Code:	'rackspace'`)
		ew.writeln(`Since:	'v0.4.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "RACKSPACE_API_KEY":	API key`)
		ew.writeln(`	- "RACKSPACE_USER":	API user`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "RACKSPACE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "RACKSPACE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "RACKSPACE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "RACKSPACE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/rackspace`)

	case "rfc2136":
		// generated from: providers/dns/rfc2136/rfc2136.toml
		ew.writeln(`Configuration for RFC2136.`)
		ew.writeln(`Code:	'rfc2136'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "RFC2136_NAMESERVER":	Network address in the form "host" or "host:port"`)
		ew.writeln(`	- "RFC2136_TSIG_ALGORITHM":	TSIG algorythm. See [miekg/dns#tsig.go](https://github.com/miekg/dns/blob/master/tsig.go) for supported values. To disable TSIG authentication, leave the 'RFC2136_TSIG*' variables unset.`)
		ew.writeln(`	- "RFC2136_TSIG_KEY":	Name of the secret key as defined in DNS server configuration. To disable TSIG authentication, leave the 'RFC2136_TSIG*' variables unset.`)
		ew.writeln(`	- "RFC2136_TSIG_SECRET":	Secret key payload. To disable TSIG authentication, leave the' RFC2136_TSIG*' variables unset.`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "RFC2136_DNS_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "RFC2136_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "RFC2136_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "RFC2136_SEQUENCE_INTERVAL":	Interval between iteration`)
		ew.writeln(`	- "RFC2136_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/rfc2136`)

	case "route53":
		// generated from: providers/dns/route53/route53.toml
		ew.writeln(`Configuration for Amazon Route 53.`)
		ew.writeln(`Code:	'route53'`)
		ew.writeln(`Since:	'v0.3.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "AWS_ACCESS_KEY_ID":	Managed by the AWS client`)
		ew.writeln(`	- "AWS_HOSTED_ZONE_ID":	Override the hosted zone ID`)
		ew.writeln(`	- "AWS_REGION":	Managed by the AWS client`)
		ew.writeln(`	- "AWS_SECRET_ACCESS_KEY":	Managed by the AWS client`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "AWS_MAX_RETRIES":	The number of maximum returns the service will use to make an individual API request`)
		ew.writeln(`	- "AWS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "AWS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "AWS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/route53`)

	case "sakuracloud":
		// generated from: providers/dns/sakuracloud/sakuracloud.toml
		ew.writeln(`Configuration for Sakura Cloud.`)
		ew.writeln(`Code:	'sakuracloud'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "SAKURACLOUD_ACCESS_TOKEN":	Access token`)
		ew.writeln(`	- "SAKURACLOUD_ACCESS_TOKEN_SECRET":	Access token secret`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "SAKURACLOUD_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "SAKURACLOUD_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "SAKURACLOUD_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "SAKURACLOUD_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/sakuracloud`)

	case "selectel":
		// generated from: providers/dns/selectel/selectel.toml
		ew.writeln(`Configuration for Selectel.`)
		ew.writeln(`Code:	'selectel'`)
		ew.writeln(`Since:	'v1.2.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "SELECTEL_API_TOKEN":	API token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "SELECTEL_BASE_URL":	API endpoint URL`)
		ew.writeln(`	- "SELECTEL_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "SELECTEL_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "SELECTEL_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "SELECTEL_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/selectel`)

	case "stackpath":
		// generated from: providers/dns/stackpath/stackpath.toml
		ew.writeln(`Configuration for Stackpath.`)
		ew.writeln(`Code:	'stackpath'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "STACKPATH_CLIENT_ID":	Client ID`)
		ew.writeln(`	- "STACKPATH_CLIENT_SECRET":	Client secret`)
		ew.writeln(`	- "STACKPATH_STACK_ID":	Stack ID`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "STACKPATH_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "STACKPATH_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "STACKPATH_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/stackpath`)

	case "transip":
		// generated from: providers/dns/transip/transip.toml
		ew.writeln(`Configuration for TransIP.`)
		ew.writeln(`Code:	'transip'`)
		ew.writeln(`Since:	'v2.0.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "TRANSIP_ACCOUNT_NAME":	Account name`)
		ew.writeln(`	- "TRANSIP_PRIVATE_KEY_PATH":	Private key path`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "TRANSIP_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "TRANSIP_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "TRANSIP_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/transip`)

	case "vegadns":
		// generated from: providers/dns/vegadns/vegadns.toml
		ew.writeln(`Configuration for VegaDNS.`)
		ew.writeln(`Code:	'vegadns'`)
		ew.writeln(`Since:	'v1.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "SECRET_VEGADNS_KEY":	API key`)
		ew.writeln(`	- "SECRET_VEGADNS_SECRET":	API secret`)
		ew.writeln(`	- "VEGADNS_URL":	API endpoint URL`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "VEGADNS_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "VEGADNS_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "VEGADNS_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/vegadns`)

	case "versio":
		// generated from: providers/dns/versio/versio.toml
		ew.writeln(`Configuration for Versio.[nl|eu|uk].`)
		ew.writeln(`Code:	'versio'`)
		ew.writeln(`Since:	'v2.7.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "VERSIO_PASSWORD":	Basic authentication password`)
		ew.writeln(`	- "VERSIO_USERNAME":	Basic authentication username`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "VERSIO_ENDPOINT":	The endpoint URL of the API Server`)
		ew.writeln(`	- "VERSIO_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "VERSIO_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "VERSIO_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "VERSIO_SEQUENCE_INTERVAL":	Interval between iteration, default 60s`)
		ew.writeln(`	- "VERSIO_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/versio`)

	case "vscale":
		// generated from: providers/dns/vscale/vscale.toml
		ew.writeln(`Configuration for Vscale.`)
		ew.writeln(`Code:	'vscale'`)
		ew.writeln(`Since:	'v2.0.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "VSCALE_API_TOKEN":	API token`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "VSCALE_BASE_URL":	API enddpoint URL`)
		ew.writeln(`	- "VSCALE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "VSCALE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "VSCALE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "VSCALE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/vscale`)

	case "vultr":
		// generated from: providers/dns/vultr/vultr.toml
		ew.writeln(`Configuration for Vultr.`)
		ew.writeln(`Code:	'vultr'`)
		ew.writeln(`Since:	'v0.3.1'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "VULTR_API_KEY":	API key`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "VULTR_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "VULTR_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "VULTR_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "VULTR_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/vultr`)

	case "zoneee":
		// generated from: providers/dns/zoneee/zoneee.toml
		ew.writeln(`Configuration for Zone.ee.`)
		ew.writeln(`Code:	'zoneee'`)
		ew.writeln(`Since:	'v2.1.0'`)
		ew.writeln()

		ew.writeln(`Credentials:`)
		ew.writeln(`	- "ZONEEE_API_KEY":	API key`)
		ew.writeln(`	- "ZONEEE_API_USER":	API user`)
		ew.writeln()

		ew.writeln(`Additional Configuration:`)
		ew.writeln(`	- "ZONEEE_ENDPOINT":	API endpoint URL`)
		ew.writeln(`	- "ZONEEE_HTTP_TIMEOUT":	API request timeout`)
		ew.writeln(`	- "ZONEEE_POLLING_INTERVAL":	Time between DNS propagation check`)
		ew.writeln(`	- "ZONEEE_PROPAGATION_TIMEOUT":	Maximum waiting time for DNS propagation`)
		ew.writeln(`	- "ZONEEE_TTL":	The TTL of the TXT record used for the DNS challenge`)

		ew.writeln()
		ew.writeln(`More information: https://go-acme.github.io/lego/dns/zoneee`)

	case "manual":
		ew.writeln(`Solving the DNS-01 challenge using CLI prompt.`)
	default:
		return fmt.Errorf("%q is not yet supported", name)
	}

	if ew.err != nil {
		return fmt.Errorf("error: %v", ew.err)
	}

	return w.Flush()
}
