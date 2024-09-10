package dns

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/acmedns"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/allinkl"
	"github.com/go-acme/lego/v4/providers/dns/arvancloud"
	"github.com/go-acme/lego/v4/providers/dns/auroradns"
	"github.com/go-acme/lego/v4/providers/dns/autodns"
	"github.com/go-acme/lego/v4/providers/dns/azure"
	"github.com/go-acme/lego/v4/providers/dns/azuredns"
	"github.com/go-acme/lego/v4/providers/dns/bindman"
	"github.com/go-acme/lego/v4/providers/dns/bluecat"
	"github.com/go-acme/lego/v4/providers/dns/brandit"
	"github.com/go-acme/lego/v4/providers/dns/bunny"
	"github.com/go-acme/lego/v4/providers/dns/checkdomain"
	"github.com/go-acme/lego/v4/providers/dns/civo"
	"github.com/go-acme/lego/v4/providers/dns/clouddns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/cloudns"
	"github.com/go-acme/lego/v4/providers/dns/cloudru"
	"github.com/go-acme/lego/v4/providers/dns/cloudxns"
	"github.com/go-acme/lego/v4/providers/dns/conoha"
	"github.com/go-acme/lego/v4/providers/dns/constellix"
	"github.com/go-acme/lego/v4/providers/dns/cpanel"
	"github.com/go-acme/lego/v4/providers/dns/derak"
	"github.com/go-acme/lego/v4/providers/dns/desec"
	"github.com/go-acme/lego/v4/providers/dns/designate"
	"github.com/go-acme/lego/v4/providers/dns/digitalocean"
	"github.com/go-acme/lego/v4/providers/dns/directadmin"
	"github.com/go-acme/lego/v4/providers/dns/dnshomede"
	"github.com/go-acme/lego/v4/providers/dns/dnsimple"
	"github.com/go-acme/lego/v4/providers/dns/dnsmadeeasy"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/providers/dns/dode"
	"github.com/go-acme/lego/v4/providers/dns/domeneshop"
	"github.com/go-acme/lego/v4/providers/dns/dreamhost"
	"github.com/go-acme/lego/v4/providers/dns/duckdns"
	"github.com/go-acme/lego/v4/providers/dns/dyn"
	"github.com/go-acme/lego/v4/providers/dns/dynu"
	"github.com/go-acme/lego/v4/providers/dns/easydns"
	"github.com/go-acme/lego/v4/providers/dns/edgedns"
	"github.com/go-acme/lego/v4/providers/dns/efficientip"
	"github.com/go-acme/lego/v4/providers/dns/epik"
	"github.com/go-acme/lego/v4/providers/dns/exec"
	"github.com/go-acme/lego/v4/providers/dns/exoscale"
	"github.com/go-acme/lego/v4/providers/dns/freemyip"
	"github.com/go-acme/lego/v4/providers/dns/gandi"
	"github.com/go-acme/lego/v4/providers/dns/gandiv5"
	"github.com/go-acme/lego/v4/providers/dns/gcloud"
	"github.com/go-acme/lego/v4/providers/dns/gcore"
	"github.com/go-acme/lego/v4/providers/dns/glesys"
	"github.com/go-acme/lego/v4/providers/dns/godaddy"
	"github.com/go-acme/lego/v4/providers/dns/googledomains"
	"github.com/go-acme/lego/v4/providers/dns/hetzner"
	"github.com/go-acme/lego/v4/providers/dns/hostingde"
	"github.com/go-acme/lego/v4/providers/dns/hosttech"
	"github.com/go-acme/lego/v4/providers/dns/httpnet"
	"github.com/go-acme/lego/v4/providers/dns/httpreq"
	"github.com/go-acme/lego/v4/providers/dns/huaweicloud"
	"github.com/go-acme/lego/v4/providers/dns/hurricane"
	"github.com/go-acme/lego/v4/providers/dns/hyperone"
	"github.com/go-acme/lego/v4/providers/dns/ibmcloud"
	"github.com/go-acme/lego/v4/providers/dns/iij"
	"github.com/go-acme/lego/v4/providers/dns/iijdpf"
	"github.com/go-acme/lego/v4/providers/dns/infoblox"
	"github.com/go-acme/lego/v4/providers/dns/infomaniak"
	"github.com/go-acme/lego/v4/providers/dns/internetbs"
	"github.com/go-acme/lego/v4/providers/dns/inwx"
	"github.com/go-acme/lego/v4/providers/dns/ionos"
	"github.com/go-acme/lego/v4/providers/dns/ipv64"
	"github.com/go-acme/lego/v4/providers/dns/iwantmyname"
	"github.com/go-acme/lego/v4/providers/dns/joker"
	"github.com/go-acme/lego/v4/providers/dns/liara"
	"github.com/go-acme/lego/v4/providers/dns/lightsail"
	"github.com/go-acme/lego/v4/providers/dns/limacity"
	"github.com/go-acme/lego/v4/providers/dns/linode"
	"github.com/go-acme/lego/v4/providers/dns/liquidweb"
	"github.com/go-acme/lego/v4/providers/dns/loopia"
	"github.com/go-acme/lego/v4/providers/dns/luadns"
	"github.com/go-acme/lego/v4/providers/dns/mailinabox"
	"github.com/go-acme/lego/v4/providers/dns/metaname"
	"github.com/go-acme/lego/v4/providers/dns/mijnhost"
	"github.com/go-acme/lego/v4/providers/dns/mittwald"
	"github.com/go-acme/lego/v4/providers/dns/mydnsjp"
	"github.com/go-acme/lego/v4/providers/dns/mythicbeasts"
	"github.com/go-acme/lego/v4/providers/dns/namecheap"
	"github.com/go-acme/lego/v4/providers/dns/namedotcom"
	"github.com/go-acme/lego/v4/providers/dns/namesilo"
	"github.com/go-acme/lego/v4/providers/dns/nearlyfreespeech"
	"github.com/go-acme/lego/v4/providers/dns/netcup"
	"github.com/go-acme/lego/v4/providers/dns/netlify"
	"github.com/go-acme/lego/v4/providers/dns/nicmanager"
	"github.com/go-acme/lego/v4/providers/dns/nifcloud"
	"github.com/go-acme/lego/v4/providers/dns/njalla"
	"github.com/go-acme/lego/v4/providers/dns/nodion"
	"github.com/go-acme/lego/v4/providers/dns/ns1"
	"github.com/go-acme/lego/v4/providers/dns/oraclecloud"
	"github.com/go-acme/lego/v4/providers/dns/otc"
	"github.com/go-acme/lego/v4/providers/dns/ovh"
	"github.com/go-acme/lego/v4/providers/dns/pdns"
	"github.com/go-acme/lego/v4/providers/dns/plesk"
	"github.com/go-acme/lego/v4/providers/dns/porkbun"
	"github.com/go-acme/lego/v4/providers/dns/rackspace"
	"github.com/go-acme/lego/v4/providers/dns/rcodezero"
	"github.com/go-acme/lego/v4/providers/dns/regru"
	"github.com/go-acme/lego/v4/providers/dns/rfc2136"
	"github.com/go-acme/lego/v4/providers/dns/rimuhosting"
	"github.com/go-acme/lego/v4/providers/dns/route53"
	"github.com/go-acme/lego/v4/providers/dns/safedns"
	"github.com/go-acme/lego/v4/providers/dns/sakuracloud"
	"github.com/go-acme/lego/v4/providers/dns/scaleway"
	"github.com/go-acme/lego/v4/providers/dns/selectel"
	"github.com/go-acme/lego/v4/providers/dns/selectelv2"
	"github.com/go-acme/lego/v4/providers/dns/servercow"
	"github.com/go-acme/lego/v4/providers/dns/shellrent"
	"github.com/go-acme/lego/v4/providers/dns/simply"
	"github.com/go-acme/lego/v4/providers/dns/sonic"
	"github.com/go-acme/lego/v4/providers/dns/stackpath"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/providers/dns/transip"
	"github.com/go-acme/lego/v4/providers/dns/ultradns"
	"github.com/go-acme/lego/v4/providers/dns/variomedia"
	"github.com/go-acme/lego/v4/providers/dns/vegadns"
	"github.com/go-acme/lego/v4/providers/dns/vercel"
	"github.com/go-acme/lego/v4/providers/dns/versio"
	"github.com/go-acme/lego/v4/providers/dns/vinyldns"
	"github.com/go-acme/lego/v4/providers/dns/vkcloud"
	"github.com/go-acme/lego/v4/providers/dns/vscale"
	"github.com/go-acme/lego/v4/providers/dns/vultr"
	"github.com/go-acme/lego/v4/providers/dns/webnames"
	"github.com/go-acme/lego/v4/providers/dns/websupport"
	"github.com/go-acme/lego/v4/providers/dns/wedos"
	"github.com/go-acme/lego/v4/providers/dns/yandex"
	"github.com/go-acme/lego/v4/providers/dns/yandex360"
	"github.com/go-acme/lego/v4/providers/dns/yandexcloud"
	"github.com/go-acme/lego/v4/providers/dns/zoneee"
	"github.com/go-acme/lego/v4/providers/dns/zonomi"
)

// NewDNSChallengeProviderByName Factory for DNS providers.
func NewDNSChallengeProviderByName(name string) (challenge.Provider, error) {
	switch name {
	case "acme-dns": // TODO(ldez): remove "-" in v5
		return acmedns.NewDNSProvider()
	case "alidns":
		return alidns.NewDNSProvider()
	case "allinkl":
		return allinkl.NewDNSProvider()
	case "arvancloud":
		return arvancloud.NewDNSProvider()
	case "azure":
		return azure.NewDNSProvider()
	case "azuredns":
		return azuredns.NewDNSProvider()
	case "auroradns":
		return auroradns.NewDNSProvider()
	case "autodns":
		return autodns.NewDNSProvider()
	case "bindman":
		return bindman.NewDNSProvider()
	case "bluecat":
		return bluecat.NewDNSProvider()
	case "brandit":
		return brandit.NewDNSProvider()
	case "bunny":
		return bunny.NewDNSProvider()
	case "checkdomain":
		return checkdomain.NewDNSProvider()
	case "civo":
		return civo.NewDNSProvider()
	case "clouddns":
		return clouddns.NewDNSProvider()
	case "cloudflare":
		return cloudflare.NewDNSProvider()
	case "cloudns":
		return cloudns.NewDNSProvider()
	case "cloudru":
		return cloudru.NewDNSProvider()
	case "cloudxns":
		return cloudxns.NewDNSProvider()
	case "conoha":
		return conoha.NewDNSProvider()
	case "constellix":
		return constellix.NewDNSProvider()
	case "cpanel":
		return cpanel.NewDNSProvider()
	case "derak":
		return derak.NewDNSProvider()
	case "desec":
		return desec.NewDNSProvider()
	case "designate":
		return designate.NewDNSProvider()
	case "digitalocean":
		return digitalocean.NewDNSProvider()
	case "directadmin":
		return directadmin.NewDNSProvider()
	case "dnshomede":
		return dnshomede.NewDNSProvider()
	case "dnsimple":
		return dnsimple.NewDNSProvider()
	case "dnsmadeeasy":
		return dnsmadeeasy.NewDNSProvider()
	case "dnspod":
		return dnspod.NewDNSProvider()
	case "dode":
		return dode.NewDNSProvider()
	case "domeneshop", "domainnameshop":
		return domeneshop.NewDNSProvider()
	case "dreamhost":
		return dreamhost.NewDNSProvider()
	case "duckdns":
		return duckdns.NewDNSProvider()
	case "dyn":
		return dyn.NewDNSProvider()
	case "dynu":
		return dynu.NewDNSProvider()
	case "easydns":
		return easydns.NewDNSProvider()
	case "edgedns", "fastdns": // "fastdns" is for compatibility with v3, must be dropped in v5
		return edgedns.NewDNSProvider()
	case "efficientip":
		return efficientip.NewDNSProvider()
	case "epik":
		return epik.NewDNSProvider()
	case "exec":
		return exec.NewDNSProvider()
	case "exoscale":
		return exoscale.NewDNSProvider()
	case "freemyip":
		return freemyip.NewDNSProvider()
	case "gandi":
		return gandi.NewDNSProvider()
	case "gandiv5":
		return gandiv5.NewDNSProvider()
	case "gcloud":
		return gcloud.NewDNSProvider()
	case "gcore":
		return gcore.NewDNSProvider()
	case "glesys":
		return glesys.NewDNSProvider()
	case "godaddy":
		return godaddy.NewDNSProvider()
	case "googledomains":
		return googledomains.NewDNSProvider()
	case "hetzner":
		return hetzner.NewDNSProvider()
	case "hostingde":
		return hostingde.NewDNSProvider()
	case "hosttech":
		return hosttech.NewDNSProvider()
	case "httpnet":
		return httpnet.NewDNSProvider()
	case "httpreq":
		return httpreq.NewDNSProvider()
	case "huaweicloud":
		return huaweicloud.NewDNSProvider()
	case "hurricane":
		return hurricane.NewDNSProvider()
	case "hyperone":
		return hyperone.NewDNSProvider()
	case "ibmcloud":
		return ibmcloud.NewDNSProvider()
	case "iij":
		return iij.NewDNSProvider()
	case "iijdpf":
		return iijdpf.NewDNSProvider()
	case "infoblox":
		return infoblox.NewDNSProvider()
	case "infomaniak":
		return infomaniak.NewDNSProvider()
	case "internetbs":
		return internetbs.NewDNSProvider()
	case "inwx":
		return inwx.NewDNSProvider()
	case "ionos":
		return ionos.NewDNSProvider()
	case "ipv64":
		return ipv64.NewDNSProvider()
	case "iwantmyname":
		return iwantmyname.NewDNSProvider()
	case "joker":
		return joker.NewDNSProvider()
	case "liara":
		return liara.NewDNSProvider()
	case "lightsail":
		return lightsail.NewDNSProvider()
	case "limacity":
		return limacity.NewDNSProvider()
	case "linode", "linodev4": // "linodev4" is for compatibility with v3, must be dropped in v5
		return linode.NewDNSProvider()
	case "liquidweb":
		return liquidweb.NewDNSProvider()
	case "loopia":
		return loopia.NewDNSProvider()
	case "luadns":
		return luadns.NewDNSProvider()
	case "mailinabox":
		return mailinabox.NewDNSProvider()
	case "manual":
		return dns01.NewDNSProviderManual()
	case "metaname":
		return metaname.NewDNSProvider()
	case "mijnhost":
		return mijnhost.NewDNSProvider()
	case "mittwald":
		return mittwald.NewDNSProvider()
	case "mydnsjp":
		return mydnsjp.NewDNSProvider()
	case "mythicbeasts":
		return mythicbeasts.NewDNSProvider()
	case "namecheap":
		return namecheap.NewDNSProvider()
	case "namedotcom":
		return namedotcom.NewDNSProvider()
	case "namesilo":
		return namesilo.NewDNSProvider()
	case "nearlyfreespeech":
		return nearlyfreespeech.NewDNSProvider()
	case "netcup":
		return netcup.NewDNSProvider()
	case "netlify":
		return netlify.NewDNSProvider()
	case "nicmanager":
		return nicmanager.NewDNSProvider()
	case "nifcloud":
		return nifcloud.NewDNSProvider()
	case "njalla":
		return njalla.NewDNSProvider()
	case "nodion":
		return nodion.NewDNSProvider()
	case "ns1":
		return ns1.NewDNSProvider()
	case "oraclecloud":
		return oraclecloud.NewDNSProvider()
	case "otc":
		return otc.NewDNSProvider()
	case "ovh":
		return ovh.NewDNSProvider()
	case "pdns":
		return pdns.NewDNSProvider()
	case "plesk":
		return plesk.NewDNSProvider()
	case "porkbun":
		return porkbun.NewDNSProvider()
	case "rackspace":
		return rackspace.NewDNSProvider()
	case "rcodezero":
		return rcodezero.NewDNSProvider()
	case "regru":
		return regru.NewDNSProvider()
	case "rfc2136":
		return rfc2136.NewDNSProvider()
	case "rimuhosting":
		return rimuhosting.NewDNSProvider()
	case "route53":
		return route53.NewDNSProvider()
	case "safedns":
		return safedns.NewDNSProvider()
	case "sakuracloud":
		return sakuracloud.NewDNSProvider()
	case "scaleway":
		return scaleway.NewDNSProvider()
	case "selectel":
		return selectel.NewDNSProvider()
	case "selectelv2":
		return selectelv2.NewDNSProvider()
	case "servercow":
		return servercow.NewDNSProvider()
	case "shellrent":
		return shellrent.NewDNSProvider()
	case "simply":
		return simply.NewDNSProvider()
	case "sonic":
		return sonic.NewDNSProvider()
	case "stackpath":
		return stackpath.NewDNSProvider()
	case "tencentcloud":
		return tencentcloud.NewDNSProvider()
	case "transip":
		return transip.NewDNSProvider()
	case "ultradns":
		return ultradns.NewDNSProvider()
	case "variomedia":
		return variomedia.NewDNSProvider()
	case "vegadns":
		return vegadns.NewDNSProvider()
	case "vercel":
		return vercel.NewDNSProvider()
	case "versio":
		return versio.NewDNSProvider()
	case "vinyldns":
		return vinyldns.NewDNSProvider()
	case "vkcloud":
		return vkcloud.NewDNSProvider()
	case "vscale":
		return vscale.NewDNSProvider()
	case "vultr":
		return vultr.NewDNSProvider()
	case "webnames":
		return webnames.NewDNSProvider()
	case "websupport":
		return websupport.NewDNSProvider()
	case "wedos":
		return wedos.NewDNSProvider()
	case "yandex":
		return yandex.NewDNSProvider()
	case "yandex360":
		return yandex360.NewDNSProvider()
	case "yandexcloud":
		return yandexcloud.NewDNSProvider()
	case "zoneee":
		return zoneee.NewDNSProvider()
	case "zonomi":
		return zonomi.NewDNSProvider()
	default:
		return nil, fmt.Errorf("unrecognized DNS provider: %s", name)
	}
}
