package dns

import (
	"fmt"

	"github.com/go-acme/lego/v3/challenge"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/providers/dns/acmedns"
	"github.com/go-acme/lego/v3/providers/dns/alidns"
	"github.com/go-acme/lego/v3/providers/dns/arvancloud"
	"github.com/go-acme/lego/v3/providers/dns/auroradns"
	"github.com/go-acme/lego/v3/providers/dns/autodns"
	"github.com/go-acme/lego/v3/providers/dns/azure"
	"github.com/go-acme/lego/v3/providers/dns/bindman"
	"github.com/go-acme/lego/v3/providers/dns/bluecat"
	"github.com/go-acme/lego/v3/providers/dns/checkdomain"
	"github.com/go-acme/lego/v3/providers/dns/clouddns"
	"github.com/go-acme/lego/v3/providers/dns/cloudflare"
	"github.com/go-acme/lego/v3/providers/dns/cloudns"
	"github.com/go-acme/lego/v3/providers/dns/cloudxns"
	"github.com/go-acme/lego/v3/providers/dns/conoha"
	"github.com/go-acme/lego/v3/providers/dns/constellix"
	"github.com/go-acme/lego/v3/providers/dns/desec"
	"github.com/go-acme/lego/v3/providers/dns/designate"
	"github.com/go-acme/lego/v3/providers/dns/digitalocean"
	"github.com/go-acme/lego/v3/providers/dns/dnsimple"
	"github.com/go-acme/lego/v3/providers/dns/dnsmadeeasy"
	"github.com/go-acme/lego/v3/providers/dns/dnspod"
	"github.com/go-acme/lego/v3/providers/dns/dode"
	"github.com/go-acme/lego/v3/providers/dns/dreamhost"
	"github.com/go-acme/lego/v3/providers/dns/duckdns"
	"github.com/go-acme/lego/v3/providers/dns/dyn"
	"github.com/go-acme/lego/v3/providers/dns/dynu"
	"github.com/go-acme/lego/v3/providers/dns/easydns"
	"github.com/go-acme/lego/v3/providers/dns/edgedns"
	"github.com/go-acme/lego/v3/providers/dns/exec"
	"github.com/go-acme/lego/v3/providers/dns/exoscale"
	"github.com/go-acme/lego/v3/providers/dns/fastdns"
	"github.com/go-acme/lego/v3/providers/dns/gandi"
	"github.com/go-acme/lego/v3/providers/dns/gandiv5"
	"github.com/go-acme/lego/v3/providers/dns/gcloud"
	"github.com/go-acme/lego/v3/providers/dns/glesys"
	"github.com/go-acme/lego/v3/providers/dns/godaddy"
	"github.com/go-acme/lego/v3/providers/dns/hetzner"
	"github.com/go-acme/lego/v3/providers/dns/hostingde"
	"github.com/go-acme/lego/v3/providers/dns/httpreq"
	"github.com/go-acme/lego/v3/providers/dns/iij"
	"github.com/go-acme/lego/v3/providers/dns/inwx"
	"github.com/go-acme/lego/v3/providers/dns/joker"
	"github.com/go-acme/lego/v3/providers/dns/lightsail"
	"github.com/go-acme/lego/v3/providers/dns/linode"
	"github.com/go-acme/lego/v3/providers/dns/linodev4"
	"github.com/go-acme/lego/v3/providers/dns/liquidweb"
	"github.com/go-acme/lego/v3/providers/dns/luadns"
	"github.com/go-acme/lego/v3/providers/dns/mydnsjp"
	"github.com/go-acme/lego/v3/providers/dns/mythicbeasts"
	"github.com/go-acme/lego/v3/providers/dns/namecheap"
	"github.com/go-acme/lego/v3/providers/dns/namedotcom"
	"github.com/go-acme/lego/v3/providers/dns/namesilo"
	"github.com/go-acme/lego/v3/providers/dns/netcup"
	"github.com/go-acme/lego/v3/providers/dns/netlify"
	"github.com/go-acme/lego/v3/providers/dns/nifcloud"
	"github.com/go-acme/lego/v3/providers/dns/ns1"
	"github.com/go-acme/lego/v3/providers/dns/oraclecloud"
	"github.com/go-acme/lego/v3/providers/dns/otc"
	"github.com/go-acme/lego/v3/providers/dns/ovh"
	"github.com/go-acme/lego/v3/providers/dns/pdns"
	"github.com/go-acme/lego/v3/providers/dns/rackspace"
	"github.com/go-acme/lego/v3/providers/dns/regru"
	"github.com/go-acme/lego/v3/providers/dns/rfc2136"
	"github.com/go-acme/lego/v3/providers/dns/rimuhosting"
	"github.com/go-acme/lego/v3/providers/dns/route53"
	"github.com/go-acme/lego/v3/providers/dns/sakuracloud"
	"github.com/go-acme/lego/v3/providers/dns/scaleway"
	"github.com/go-acme/lego/v3/providers/dns/selectel"
	"github.com/go-acme/lego/v3/providers/dns/servercow"
	"github.com/go-acme/lego/v3/providers/dns/stackpath"
	"github.com/go-acme/lego/v3/providers/dns/transip"
	"github.com/go-acme/lego/v3/providers/dns/vegadns"
	"github.com/go-acme/lego/v3/providers/dns/versio"
	"github.com/go-acme/lego/v3/providers/dns/vscale"
	"github.com/go-acme/lego/v3/providers/dns/vultr"
	"github.com/go-acme/lego/v3/providers/dns/yandex"
	"github.com/go-acme/lego/v3/providers/dns/zoneee"
	"github.com/go-acme/lego/v3/providers/dns/zonomi"
)

// NewDNSChallengeProviderByName Factory for DNS providers.
func NewDNSChallengeProviderByName(name string, config map[string]string) (challenge.Provider, error) {
	switch name {
	case "acme-dns":
		return acmedns.NewDNSProvider(config)
	case "alidns":
		return alidns.NewDNSProvider(config)
	case "arvancloud":
		return arvancloud.NewDNSProvider(config)
	case "azure":
		return azure.NewDNSProvider(config)
	case "auroradns":
		return auroradns.NewDNSProvider(config)
	case "autodns":
		return autodns.NewDNSProvider(config)
	case "bindman":
		return bindman.NewDNSProvider(config)
	case "bluecat":
		return bluecat.NewDNSProvider(config)
	case "checkdomain":
		return checkdomain.NewDNSProvider(config)
	case "clouddns":
		return clouddns.NewDNSProvider(config)
	case "cloudflare":
		return cloudflare.NewDNSProvider(config)
	case "cloudns":
		return cloudns.NewDNSProvider(config)
	case "cloudxns":
		return cloudxns.NewDNSProvider(config)
	case "conoha":
		return conoha.NewDNSProvider(config)
	case "constellix":
		return constellix.NewDNSProvider(config)
	case "desec":
		return desec.NewDNSProvider(config)
	case "designate":
		return designate.NewDNSProvider(config)
	case "digitalocean":
		return digitalocean.NewDNSProvider(config)
	case "dnsimple":
		return dnsimple.NewDNSProvider(config)
	case "dnsmadeeasy":
		return dnsmadeeasy.NewDNSProvider(config)
	case "dnspod":
		return dnspod.NewDNSProvider(config)
	case "dode":
		return dode.NewDNSProvider(config)
	case "dreamhost":
		return dreamhost.NewDNSProvider(config)
	case "duckdns":
		return duckdns.NewDNSProvider(config)
	case "dyn":
		return dyn.NewDNSProvider(config)
	case "dynu":
		return dynu.NewDNSProvider(config)
	case "edgedns":
		return edgedns.NewDNSProvider(config)
	case "fastdns":
		return fastdns.NewDNSProvider(config)
	case "easydns":
		return easydns.NewDNSProvider(config)
	case "exec":
		return exec.NewDNSProvider(config)
	case "exoscale":
		return exoscale.NewDNSProvider(config)
	case "gandi":
		return gandi.NewDNSProvider(config)
	case "gandiv5":
		return gandiv5.NewDNSProvider(config)
	case "glesys":
		return glesys.NewDNSProvider(config)
	case "gcloud":
		return gcloud.NewDNSProvider(config)
	case "godaddy":
		return godaddy.NewDNSProvider(config)
	case "hetzner":
		return hetzner.NewDNSProvider(config)
	case "hostingde":
		return hostingde.NewDNSProvider(config)
	case "httpreq":
		return httpreq.NewDNSProvider(config)
	case "iij":
		return iij.NewDNSProvider(config)
	case "inwx":
		return inwx.NewDNSProvider(config)
	case "joker":
		return joker.NewDNSProvider(config)
	case "lightsail":
		return lightsail.NewDNSProvider(config)
	case "linode":
		return linode.NewDNSProvider(config)
	case "linodev4":
		return linodev4.NewDNSProvider(config)
	case "liquidweb":
		return liquidweb.NewDNSProvider(config)
	case "luadns":
		return luadns.NewDNSProvider(config)
	case "manual":
		return dns01.NewDNSProviderManual(config)
	case "mydnsjp":
		return mydnsjp.NewDNSProvider(config)
	case "mythicbeasts":
		return mythicbeasts.NewDNSProvider(config)
	case "namecheap":
		return namecheap.NewDNSProvider(config)
	case "namedotcom":
		return namedotcom.NewDNSProvider(config)
	case "namesilo":
		return namesilo.NewDNSProvider(config)
	case "netcup":
		return netcup.NewDNSProvider(config)
	case "netlify":
		return netlify.NewDNSProvider(config)
	case "nifcloud":
		return nifcloud.NewDNSProvider(config)
	case "ns1":
		return ns1.NewDNSProvider(config)
	case "oraclecloud":
		return oraclecloud.NewDNSProvider(config)
	case "otc":
		return otc.NewDNSProvider(config)
	case "ovh":
		return ovh.NewDNSProvider(config)
	case "pdns":
		return pdns.NewDNSProvider(config)
	case "rackspace":
		return rackspace.NewDNSProvider(config)
	case "regru":
		return regru.NewDNSProvider(config)
	case "rfc2136":
		return rfc2136.NewDNSProvider(config)
	case "rimuhosting":
		return rimuhosting.NewDNSProvider(config)
	case "route53":
		return route53.NewDNSProvider(config)
	case "sakuracloud":
		return sakuracloud.NewDNSProvider(config)
	case "scaleway":
		return scaleway.NewDNSProvider(config)
	case "selectel":
		return selectel.NewDNSProvider(config)
	case "servercow":
		return servercow.NewDNSProvider(config)
	case "stackpath":
		return stackpath.NewDNSProvider(config)
	case "transip":
		return transip.NewDNSProvider(config)
	case "vegadns":
		return vegadns.NewDNSProvider(config)
	case "versio":
		return versio.NewDNSProvider(config)
	case "vultr":
		return vultr.NewDNSProvider(config)
	case "vscale":
		return vscale.NewDNSProvider(config)
	case "yandex":
		return yandex.NewDNSProvider(config)
	case "zoneee":
		return zoneee.NewDNSProvider(config)
	case "zonomi":
		return zonomi.NewDNSProvider(config)
	default:
		return nil, fmt.Errorf("unrecognized DNS provider: %s", name)
	}
}
