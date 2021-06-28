package dns

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/acmedns"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/arvancloud"
	"github.com/go-acme/lego/v4/providers/dns/auroradns"
	"github.com/go-acme/lego/v4/providers/dns/autodns"
	"github.com/go-acme/lego/v4/providers/dns/azure"
	"github.com/go-acme/lego/v4/providers/dns/bindman"
	"github.com/go-acme/lego/v4/providers/dns/bluecat"
	"github.com/go-acme/lego/v4/providers/dns/checkdomain"
	"github.com/go-acme/lego/v4/providers/dns/clouddns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/cloudns"
	"github.com/go-acme/lego/v4/providers/dns/cloudxns"
	"github.com/go-acme/lego/v4/providers/dns/conoha"
	"github.com/go-acme/lego/v4/providers/dns/constellix"
	"github.com/go-acme/lego/v4/providers/dns/desec"
	"github.com/go-acme/lego/v4/providers/dns/designate"
	"github.com/go-acme/lego/v4/providers/dns/digitalocean"
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
	"github.com/go-acme/lego/v4/providers/dns/exec"
	"github.com/go-acme/lego/v4/providers/dns/exoscale"
	"github.com/go-acme/lego/v4/providers/dns/gandi"
	"github.com/go-acme/lego/v4/providers/dns/gandiv5"
	"github.com/go-acme/lego/v4/providers/dns/gcloud"
	"github.com/go-acme/lego/v4/providers/dns/glesys"
	"github.com/go-acme/lego/v4/providers/dns/godaddy"
	"github.com/go-acme/lego/v4/providers/dns/hetzner"
	"github.com/go-acme/lego/v4/providers/dns/hostingde"
	"github.com/go-acme/lego/v4/providers/dns/httpreq"
	"github.com/go-acme/lego/v4/providers/dns/hurricane"
	"github.com/go-acme/lego/v4/providers/dns/hyperone"
	"github.com/go-acme/lego/v4/providers/dns/iij"
	"github.com/go-acme/lego/v4/providers/dns/infoblox"
	"github.com/go-acme/lego/v4/providers/dns/infomaniak"
	"github.com/go-acme/lego/v4/providers/dns/internetbs"
	"github.com/go-acme/lego/v4/providers/dns/inwx"
	"github.com/go-acme/lego/v4/providers/dns/ionos"
	"github.com/go-acme/lego/v4/providers/dns/joker"
	"github.com/go-acme/lego/v4/providers/dns/lightsail"
	"github.com/go-acme/lego/v4/providers/dns/linode"
	"github.com/go-acme/lego/v4/providers/dns/liquidweb"
	"github.com/go-acme/lego/v4/providers/dns/loopia"
	"github.com/go-acme/lego/v4/providers/dns/luadns"
	"github.com/go-acme/lego/v4/providers/dns/mydnsjp"
	"github.com/go-acme/lego/v4/providers/dns/mythicbeasts"
	"github.com/go-acme/lego/v4/providers/dns/namecheap"
	"github.com/go-acme/lego/v4/providers/dns/namedotcom"
	"github.com/go-acme/lego/v4/providers/dns/namesilo"
	"github.com/go-acme/lego/v4/providers/dns/netcup"
	"github.com/go-acme/lego/v4/providers/dns/netlify"
	"github.com/go-acme/lego/v4/providers/dns/nifcloud"
	"github.com/go-acme/lego/v4/providers/dns/njalla"
	"github.com/go-acme/lego/v4/providers/dns/ns1"
	"github.com/go-acme/lego/v4/providers/dns/oraclecloud"
	"github.com/go-acme/lego/v4/providers/dns/otc"
	"github.com/go-acme/lego/v4/providers/dns/ovh"
	"github.com/go-acme/lego/v4/providers/dns/pdns"
	"github.com/go-acme/lego/v4/providers/dns/porkbun"
	"github.com/go-acme/lego/v4/providers/dns/rackspace"
	"github.com/go-acme/lego/v4/providers/dns/regru"
	"github.com/go-acme/lego/v4/providers/dns/rfc2136"
	"github.com/go-acme/lego/v4/providers/dns/rimuhosting"
	"github.com/go-acme/lego/v4/providers/dns/route53"
	"github.com/go-acme/lego/v4/providers/dns/sakuracloud"
	"github.com/go-acme/lego/v4/providers/dns/scaleway"
	"github.com/go-acme/lego/v4/providers/dns/selectel"
	"github.com/go-acme/lego/v4/providers/dns/servercow"
	"github.com/go-acme/lego/v4/providers/dns/simply"
	"github.com/go-acme/lego/v4/providers/dns/sonic"
	"github.com/go-acme/lego/v4/providers/dns/stackpath"
	"github.com/go-acme/lego/v4/providers/dns/transip"
	"github.com/go-acme/lego/v4/providers/dns/vegadns"
	"github.com/go-acme/lego/v4/providers/dns/versio"
	"github.com/go-acme/lego/v4/providers/dns/vinyldns"
	"github.com/go-acme/lego/v4/providers/dns/vscale"
	"github.com/go-acme/lego/v4/providers/dns/vultr"
	"github.com/go-acme/lego/v4/providers/dns/wedos"
	"github.com/go-acme/lego/v4/providers/dns/yandex"
	"github.com/go-acme/lego/v4/providers/dns/zoneee"
	"github.com/go-acme/lego/v4/providers/dns/zonomi"
)

type (
	// SupportedProvider index the framework supported provider.
	SupportedProvider int

	providerFactory func() (challenge.Provider, error)
)

const (
	ProviderAcmeDNS SupportedProvider = iota
	ProviderAliDNS
	ProviderArvanCloud
	ProviderAzure
	ProviderAuroraDNS
	ProviderAutoDNS
	ProviderBindman
	ProviderBluecat
	ProviderCheckDomain
	ProviderCloudDNS
	ProviderCloudflare
	ProviderCloudns
	ProviderCloudXns
	ProviderConoha
	ProviderConstellix
	ProviderDesec
	ProviderDesignate
	ProviderDigitalOcean
	ProviderDNSimple
	ProviderDNSMadeEasy
	ProviderDNSpod
	ProviderDode
	ProviderDomeneShop
	ProviderDomainNameShop
	ProviderDeamHost
	ProviderDuckDNS
	ProviderDyn
	ProviderDynu
	ProviderEasyDNS
	ProviderEdgeDNS
	ProviderExec
	ProviderExoscale
	ProviderFastDNS
	ProviderGandi
	ProviderGandiv5
	ProviderGlesys
	ProviderGcloud
	ProviderGodaddy
	ProviderHetzner
	ProviderHostingde
	ProviderHttpreq
	ProviderHurricane
	ProviderHypreone
	ProviderIij
	ProviderInfoblox
	ProviderInfomaniak
	ProviderInternetBs
	ProviderInwx
	ProviderIonos
	ProviderJoker
	ProviderLightSail
	ProviderLinode
	ProviderLinodEv4
	ProviderLiquidWeb
	ProviderLudDNS
	ProviderLoopia
	ProviderManual
	ProviderMyDNSJp
	ProviderMythicBeasts
	ProviderNamecheap
	ProviderNameDocCom
	ProviderNamesilo
	ProviderNetcup
	ProviderNetlify
	ProviderNifCloud
	ProviderNialla
	ProviderNs1
	ProviderOracleCloud
	ProviderOtc
	ProviderOvh
	ProviderPdns
	ProviderPorkbun
	ProviderRackSpace
	ProviderRegru
	ProviderRfc2136
	ProviderRimuHosting
	ProviderRoute53
	ProviderSakuraCloud
	ProviderScaleway
	ProviderSelectEl
	ProviderServerCow
	ProviderSimply
	ProviderSonic
	ProviderStackPath
	ProviderTansip
	ProviderVegaDNS
	ProviderVersio
	ProviderVinylDNS
	ProviderVultr
	ProviderVscale
	ProviderWedos
	ProviderYandex
	ProviderZonnee
	ProviderZonomi
)

//nolint: gochecknoglobals
var (
	_str2provider = map[string]SupportedProvider{
		"acme-dns":       ProviderAcmeDNS,
		"alidns":         ProviderAliDNS,
		"arvancloud":     ProviderArvanCloud,
		"azure":          ProviderAzure,
		"auroradns":      ProviderAuroraDNS,
		"autodns":        ProviderAutoDNS,
		"bindman":        ProviderBindman,
		"bluecat":        ProviderBluecat,
		"checkdomain":    ProviderCheckDomain,
		"clouddns":       ProviderCloudDNS,
		"cloudflare":     ProviderCloudflare,
		"cloudns":        ProviderCloudns,
		"cloudxns":       ProviderCloudXns,
		"conoha":         ProviderConoha,
		"constellix":     ProviderConstellix,
		"desec":          ProviderDesec,
		"designate":      ProviderDesignate,
		"digitalocean":   ProviderDigitalOcean,
		"dnsimple":       ProviderDNSimple,
		"dnsmadeeasy":    ProviderDNSMadeEasy,
		"dnspod":         ProviderDNSpod,
		"dode":           ProviderDode,
		"domeneshop":     ProviderDomeneShop,
		"domainnameshop": ProviderDomainNameShop,
		"dreamhost":      ProviderDeamHost,
		"duckdns":        ProviderDuckDNS,
		"dyn":            ProviderDyn,
		"dynu":           ProviderDynu,
		"easydns":        ProviderEasyDNS,
		"edgedns":        ProviderEdgeDNS,
		"exec":           ProviderExec,
		"exoscale":       ProviderExoscale,
		"fastdns":        ProviderFastDNS,
		"gandi":          ProviderGandi,
		"gandiv5":        ProviderGandiv5,
		"glesys":         ProviderGlesys,
		"gcloud":         ProviderGcloud,
		"godaddy":        ProviderGodaddy,
		"hetzner":        ProviderHetzner,
		"hostingde":      ProviderHostingde,
		"httpreq":        ProviderHttpreq,
		"hurricane":      ProviderHurricane,
		"hyperone":       ProviderHypreone,
		"iij":            ProviderIij,
		"infoblox":       ProviderInfoblox,
		"infomaniak":     ProviderInfomaniak,
		"internetbs":     ProviderInternetBs,
		"inwx":           ProviderInwx,
		"ionos":          ProviderIonos,
		"joker":          ProviderJoker,
		"lightsail":      ProviderLightSail,
		"linode":         ProviderLinode,
		"linodev4":       ProviderLinodEv4,
		"liquidweb":      ProviderLiquidWeb,
		"luadns":         ProviderLudDNS,
		"loopia":         ProviderLoopia,
		"manual":         ProviderManual,
		"mydnsjp":        ProviderMyDNSJp,
		"mythicbeasts":   ProviderMythicBeasts,
		"namecheap":      ProviderNamecheap,
		"namedotcom":     ProviderNameDocCom,
		"namesilo":       ProviderNamesilo,
		"netcup":         ProviderNetcup,
		"netlify":        ProviderNetlify,
		"nifcloud":       ProviderNifCloud,
		"njalla":         ProviderNialla,
		"ns1":            ProviderNs1,
		"oraclecloud":    ProviderOracleCloud,
		"otc":            ProviderOtc,
		"ovh":            ProviderOvh,
		"pdns":           ProviderPdns,
		"porkbun":        ProviderPorkbun,
		"rackspace":      ProviderRackSpace,
		"regru":          ProviderRegru,
		"rfc2136":        ProviderRfc2136,
		"rimuhosting":    ProviderRimuHosting,
		"route53":        ProviderRoute53,
		"sakuracloud":    ProviderSakuraCloud,
		"scaleway":       ProviderScaleway,
		"selectel":       ProviderSelectEl,
		"servercow":      ProviderServerCow,
		"simply":         ProviderSimply,
		"sonic":          ProviderSonic,
		"stackpath":      ProviderStackPath,
		"transip":        ProviderTansip,
		"vegadns":        ProviderVegaDNS,
		"versio":         ProviderVersio,
		"vinyldns":       ProviderVinylDNS,
		"vultr":          ProviderVultr,
		"vscale":         ProviderVscale,
		"wedos":          ProviderWedos,
		"yandex":         ProviderYandex,
		"zoneee":         ProviderZonnee,
		"zonomi":         ProviderZonomi,
	}

	_provider2str = map[SupportedProvider]string{
		ProviderAcmeDNS:        "acme-dns",
		ProviderAliDNS:         "alidns",
		ProviderArvanCloud:     "arvancloud",
		ProviderAzure:          "azure",
		ProviderAuroraDNS:      "auroradns",
		ProviderAutoDNS:        "autodns",
		ProviderBindman:        "bindman",
		ProviderBluecat:        "bluecat",
		ProviderCheckDomain:    "checkdomain",
		ProviderCloudDNS:       "clouddns",
		ProviderCloudflare:     "cloudflare",
		ProviderCloudns:        "cloudns",
		ProviderCloudXns:       "cloudxns",
		ProviderConoha:         "conoha",
		ProviderConstellix:     "constellix",
		ProviderDesec:          "desec",
		ProviderDesignate:      "designate",
		ProviderDigitalOcean:   "digitalocean",
		ProviderDNSimple:       "dnsimple",
		ProviderDNSMadeEasy:    "dnsmadeeasy",
		ProviderDNSpod:         "dnspod",
		ProviderDode:           "dode",
		ProviderDomeneShop:     "domeneshop",
		ProviderDomainNameShop: "domainnameshop",
		ProviderDeamHost:       "dreamhost",
		ProviderDuckDNS:        "duckdns",
		ProviderDyn:            "dyn",
		ProviderDynu:           "dynu",
		ProviderEasyDNS:        "easydns",
		ProviderEdgeDNS:        "edgedns",
		ProviderExec:           "exec",
		ProviderExoscale:       "exoscale",
		ProviderFastDNS:        "fastdns",
		ProviderGandi:          "gandi",
		ProviderGandiv5:        "gandiv5",
		ProviderGlesys:         "glesys",
		ProviderGcloud:         "gcloud",
		ProviderGodaddy:        "godaddy",
		ProviderHetzner:        "hetzner",
		ProviderHostingde:      "hostingde",
		ProviderHttpreq:        "httpreq",
		ProviderHurricane:      "hurricane",
		ProviderHypreone:       "hyperone",
		ProviderIij:            "iij",
		ProviderInfoblox:       "infoblox",
		ProviderInfomaniak:     "infomaniak",
		ProviderInternetBs:     "internetbs",
		ProviderInwx:           "inwx",
		ProviderIonos:          "ionos",
		ProviderJoker:          "joker",
		ProviderLightSail:      "lightsail",
		ProviderLinode:         "linode",
		ProviderLinodEv4:       "linodev4",
		ProviderLiquidWeb:      "liquidweb",
		ProviderLudDNS:         "luadns",
		ProviderLoopia:         "loopia",
		ProviderManual:         "manual",
		ProviderMyDNSJp:        "mydnsjp",
		ProviderMythicBeasts:   "mythicbeasts",
		ProviderNamecheap:      "namecheap",
		ProviderNameDocCom:     "namedotcom",
		ProviderNamesilo:       "namesilo",
		ProviderNetcup:         "netcup",
		ProviderNetlify:        "netlify",
		ProviderNifCloud:       "nifcloud",
		ProviderNialla:         "njalla",
		ProviderNs1:            "ns1",
		ProviderOracleCloud:    "oraclecloud",
		ProviderOtc:            "otc",
		ProviderOvh:            "ovh",
		ProviderPdns:           "pdns",
		ProviderPorkbun:        "porkbun",
		ProviderRackSpace:      "rackspace",
		ProviderRegru:          "regru",
		ProviderRfc2136:        "rfc2136",
		ProviderRimuHosting:    "rimuhosting",
		ProviderRoute53:        "route53",
		ProviderSakuraCloud:    "sakuracloud",
		ProviderScaleway:       "scaleway",
		ProviderSelectEl:       "selectel",
		ProviderServerCow:      "servercow",
		ProviderSimply:         "simply",
		ProviderSonic:          "sonic",
		ProviderStackPath:      "stackpath",
		ProviderTansip:         "transip",
		ProviderVegaDNS:        "vegadns",
		ProviderVersio:         "versio",
		ProviderVinylDNS:       "vinyldns",
		ProviderVultr:          "vultr",
		ProviderVscale:         "vscale",
		ProviderWedos:          "wedos",
		ProviderYandex:         "yandex",
		ProviderZonnee:         "zoneee",
		ProviderZonomi:         "zonomi",
	}

	_provider2func = map[SupportedProvider]providerFactory{
		ProviderAcmeDNS:        acmedns.NewDNSProvider,
		ProviderAliDNS:         alidns.NewDNSProvider,
		ProviderArvanCloud:     arvancloud.NewDNSProvider,
		ProviderAzure:          azure.NewDNSProvider,
		ProviderAuroraDNS:      auroradns.NewDNSProvider,
		ProviderAutoDNS:        autodns.NewDNSProvider,
		ProviderBindman:        bindman.NewDNSProvider,
		ProviderBluecat:        bluecat.NewDNSProvider,
		ProviderCheckDomain:    checkdomain.NewDNSProvider,
		ProviderCloudDNS:       clouddns.NewDNSProvider,
		ProviderCloudflare:     cloudflare.NewDNSProvider,
		ProviderCloudns:        cloudns.NewDNSProvider,
		ProviderCloudXns:       cloudxns.NewDNSProvider,
		ProviderConoha:         conoha.NewDNSProvider,
		ProviderConstellix:     constellix.NewDNSProvider,
		ProviderDesec:          desec.NewDNSProvider,
		ProviderDesignate:      designate.NewDNSProvider,
		ProviderDigitalOcean:   digitalocean.NewDNSProvider,
		ProviderDNSimple:       dnsimple.NewDNSProvider,
		ProviderDNSMadeEasy:    dnsmadeeasy.NewDNSProvider,
		ProviderDNSpod:         dnspod.NewDNSProvider,
		ProviderDode:           dode.NewDNSProvider,
		ProviderDomeneShop:     domeneshop.NewDNSProvider,
		ProviderDomainNameShop: domeneshop.NewDNSProvider,
		ProviderDeamHost:       dreamhost.NewDNSProvider,
		ProviderDuckDNS:        duckdns.NewDNSProvider,
		ProviderDyn:            dyn.NewDNSProvider,
		ProviderDynu:           dynu.NewDNSProvider,
		ProviderEasyDNS:        easydns.NewDNSProvider,
		ProviderEdgeDNS:        edgedns.NewDNSProvider,
		ProviderExec:           exec.NewDNSProvider,
		ProviderExoscale:       exoscale.NewDNSProvider,
		ProviderFastDNS:        edgedns.NewDNSProvider,
		ProviderGandi:          gandi.NewDNSProvider,
		ProviderGandiv5:        gandiv5.NewDNSProvider,
		ProviderGlesys:         glesys.NewDNSProvider,
		ProviderGcloud:         gcloud.NewDNSProvider,
		ProviderGodaddy:        godaddy.NewDNSProvider,
		ProviderHetzner:        hetzner.NewDNSProvider,
		ProviderHostingde:      hostingde.NewDNSProvider,
		ProviderHttpreq:        httpreq.NewDNSProvider,
		ProviderHurricane:      hurricane.NewDNSProvider,
		ProviderHypreone:       hyperone.NewDNSProvider,
		ProviderIij:            iij.NewDNSProvider,
		ProviderInfoblox:       infoblox.NewDNSProvider,
		ProviderInfomaniak:     infomaniak.NewDNSProvider,
		ProviderInternetBs:     internetbs.NewDNSProvider,
		ProviderInwx:           inwx.NewDNSProvider,
		ProviderIonos:          ionos.NewDNSProvider,
		ProviderJoker:          joker.NewDNSProvider,
		ProviderLightSail:      lightsail.NewDNSProvider,
		ProviderLinode:         linode.NewDNSProvider,
		ProviderLinodEv4:       linode.NewDNSProvider,
		ProviderLiquidWeb:      liquidweb.NewDNSProvider,
		ProviderLudDNS:         luadns.NewDNSProvider,
		ProviderLoopia:         loopia.NewDNSProvider,
		ProviderManual:         dns01.NewDNSProviderManual,
		ProviderMyDNSJp:        mydnsjp.NewDNSProvider,
		ProviderMythicBeasts:   mythicbeasts.NewDNSProvider,
		ProviderNamecheap:      namecheap.NewDNSProvider,
		ProviderNameDocCom:     namedotcom.NewDNSProvider,
		ProviderNamesilo:       namesilo.NewDNSProvider,
		ProviderNetcup:         netcup.NewDNSProvider,
		ProviderNetlify:        netlify.NewDNSProvider,
		ProviderNifCloud:       nifcloud.NewDNSProvider,
		ProviderNialla:         njalla.NewDNSProvider,
		ProviderNs1:            ns1.NewDNSProvider,
		ProviderOracleCloud:    oraclecloud.NewDNSProvider,
		ProviderOtc:            otc.NewDNSProvider,
		ProviderOvh:            ovh.NewDNSProvider,
		ProviderPdns:           pdns.NewDNSProvider,
		ProviderPorkbun:        porkbun.NewDNSProvider,
		ProviderRackSpace:      rackspace.NewDNSProvider,
		ProviderRegru:          regru.NewDNSProvider,
		ProviderRfc2136:        rfc2136.NewDNSProvider,
		ProviderRimuHosting:    rimuhosting.NewDNSProvider,
		ProviderRoute53:        route53.NewDNSProvider,
		ProviderSakuraCloud:    sakuracloud.NewDNSProvider,
		ProviderScaleway:       scaleway.NewDNSProvider,
		ProviderSelectEl:       selectel.NewDNSProvider,
		ProviderServerCow:      servercow.NewDNSProvider,
		ProviderSimply:         simply.NewDNSProvider,
		ProviderSonic:          sonic.NewDNSProvider,
		ProviderStackPath:      stackpath.NewDNSProvider,
		ProviderTansip:         transip.NewDNSProvider,
		ProviderVegaDNS:        vegadns.NewDNSProvider,
		ProviderVersio:         versio.NewDNSProvider,
		ProviderVinylDNS:       vinyldns.NewDNSProvider,
		ProviderVultr:          vultr.NewDNSProvider,
		ProviderVscale:         vscale.NewDNSProvider,
		ProviderWedos:          wedos.NewDNSProvider,
		ProviderYandex:         yandex.NewDNSProvider,
		ProviderZonnee:         zoneee.NewDNSProvider,
		ProviderZonomi:         zonomi.NewDNSProvider,
	}
)

// String implement the Stringer interface.
func (p SupportedProvider) String() string {
	if v, ok := _provider2str[p]; ok {
		return v
	}

	return ``
}

// MarshalJSON implement the MashalJSON interface, allowing serialization to json.
func (p SupportedProvider) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString(`"`)
	b.WriteString(p.String())
	b.WriteString(`"`)

	return b.Bytes(), nil
}

// UnmarshalJSON Implement the UnmashalJSON interface, allowing deserialization from json.
func (p *SupportedProvider) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	return p.set(j)
}

func (p *SupportedProvider) set(name string) error {
	if v, ok := _str2provider[name]; ok {
		*p = v
		return nil
	}

	return ErrUnsupportedProvider{name}
}

// ErrUnsupportedProvider is yield in case of provider not supported by
// the framework.
type ErrUnsupportedProvider struct{ name string }

// Error implemente the error interface.
func (e ErrUnsupportedProvider) Error() string {
	return fmt.Sprintf("unrecognized DNS provider: %s", e.name)
}
