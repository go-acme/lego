package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"
)

func createDNSHelp() cli.Command {
	return cli.Command{
		Name:   "dnshelp",
		Usage:  "Shows additional help for the --dns global option",
		Action: dnsHelp,
	}
}

func dnsHelp(_ *cli.Context) error {
	fmt.Printf(
		`Credentials for DNS providers must be passed through environment variables.

Here is an example bash command using the CloudFlare DNS provider:

  $ CLOUDFLARE_EMAIL=foo@bar.com \
    CLOUDFLARE_API_KEY=b9841238feb177a84330febba8a83208921177bffe733 \
    lego --dns cloudflare --domains www.example.com --email me@bar.com run

`)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "Valid providers and their associated credential environment variables:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "\tacme-dns:\tACME_DNS_API_BASE, ACME_DNS_STORAGE_PATH")
	fmt.Fprintln(w, "\talidns:\tALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY")
	fmt.Fprintln(w, "\tauroradns:\tAURORA_USER_ID, AURORA_KEY, AURORA_ENDPOINT")
	fmt.Fprintln(w, "\tazure:\tAZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID, AZURE_RESOURCE_GROUP")
	fmt.Fprintln(w, "\tbluecat:\tBLUECAT_SERVER_URL, BLUECAT_USER_NAME, BLUECAT_PASSWORD, BLUECAT_CONFIG_NAME, BLUECAT_DNS_VIEW")
	fmt.Fprintln(w, "\tcloudflare:\tCLOUDFLARE_EMAIL, CLOUDFLARE_API_KEY")
	fmt.Fprintln(w, "\tcloudxns:\tCLOUDXNS_API_KEY, CLOUDXNS_SECRET_KEY")
	fmt.Fprintln(w, "\tdigitalocean:\tDO_AUTH_TOKEN")
	fmt.Fprintln(w, "\tdnsimple:\tDNSIMPLE_EMAIL, DNSIMPLE_OAUTH_TOKEN")
	fmt.Fprintln(w, "\tdnsmadeeasy:\tDNSMADEEASY_API_KEY, DNSMADEEASY_API_SECRET")
	fmt.Fprintln(w, "\tdnspod:\tDNSPOD_API_KEY")
	fmt.Fprintln(w, "\tdreamhost:\tDREAMHOST_API_KEY")
	fmt.Fprintln(w, "\tduckdns:\tDUCKDNS_TOKEN")
	fmt.Fprintln(w, "\tdyn:\tDYN_CUSTOMER_NAME, DYN_USER_NAME, DYN_PASSWORD")
	fmt.Fprintln(w, "\texec:\tEXEC_PATH, EXEC_MODE")
	fmt.Fprintln(w, "\texoscale:\tEXOSCALE_API_KEY, EXOSCALE_API_SECRET, EXOSCALE_ENDPOINT")
	fmt.Fprintln(w, "\tfastdns:\tAKAMAI_HOST, AKAMAI_CLIENT_TOKEN, AKAMAI_CLIENT_SECRET, AKAMAI_ACCESS_TOKEN")
	fmt.Fprintln(w, "\tgandi:\tGANDI_API_KEY")
	fmt.Fprintln(w, "\tgandiv5:\tGANDIV5_API_KEY")
	fmt.Fprintln(w, "\tgcloud:\tGCE_PROJECT, GCE_SERVICE_ACCOUNT_FILE")
	fmt.Fprintln(w, "\tglesys:\tGLESYS_API_USER, GLESYS_API_KEY")
	fmt.Fprintln(w, "\tgodaddy:\tGODADDY_API_KEY, GODADDY_API_SECRET")
	fmt.Fprintln(w, "\thostingde:\tHOSTINGDE_API_KEY, HOSTINGDE_ZONE_NAME")
	fmt.Fprintln(w, "\tiij:\tIIJ_API_ACCESS_KEY, IIJ_API_SECRET_KEY, IIJ_DO_SERVICE_CODE")
	fmt.Fprintln(w, "\tlightsail:\tAWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, DNS_ZONE")
	fmt.Fprintln(w, "\tlinode:\tLINODE_API_KEY")
	fmt.Fprintln(w, "\tlinodev4:\tLINODE_TOKEN")
	fmt.Fprintln(w, "\tmanual:\tnone")
	fmt.Fprintln(w, "\tnamecheap:\tNAMECHEAP_API_USER, NAMECHEAP_API_KEY")
	fmt.Fprintln(w, "\tnamedotcom:\tNAMECOM_USERNAME, NAMECOM_API_TOKEN")
	fmt.Fprintln(w, "\tnetcup:\tNETCUP_CUSTOMER_NUMBER, NETCUP_API_KEY, NETCUP_API_PASSWORD")
	fmt.Fprintln(w, "\tnifcloud:\tNIFCLOUD_ACCESS_KEY_ID, NIFCLOUD_SECRET_ACCESS_KEY")
	fmt.Fprintln(w, "\tns1:\tNS1_API_KEY")
	fmt.Fprintln(w, "\totc:\tOTC_USER_NAME, OTC_PASSWORD, OTC_PROJECT_NAME, OTC_DOMAIN_NAME, OTC_IDENTITY_ENDPOINT")
	fmt.Fprintln(w, "\tovh:\tOVH_ENDPOINT, OVH_APPLICATION_KEY, OVH_APPLICATION_SECRET, OVH_CONSUMER_KEY")
	fmt.Fprintln(w, "\tpdns:\tPDNS_API_KEY, PDNS_API_URL")
	fmt.Fprintln(w, "\trackspace:\tRACKSPACE_USER, RACKSPACE_API_KEY")
	fmt.Fprintln(w, "\trfc2136:\tRFC2136_TSIG_KEY, RFC2136_TSIG_SECRET,\n\t\tRFC2136_TSIG_ALGORITHM, RFC2136_NAMESERVER")
	fmt.Fprintln(w, "\troute53:\tAWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION, AWS_HOSTED_ZONE_ID")
	fmt.Fprintln(w, "\tsakuracloud:\tSAKURACLOUD_ACCESS_TOKEN, SAKURACLOUD_ACCESS_TOKEN_SECRET")
	fmt.Fprintln(w, "\tstackpath:\tSTACKPATH_CLIENT_ID, STACKPATH_CLIENT_SECRET, STACKPATH_STACK_ID")
	fmt.Fprintln(w, "\tvegadns:\tSECRET_VEGADNS_KEY, SECRET_VEGADNS_SECRET, VEGADNS_URL")
	fmt.Fprintln(w, "\tvultr:\tVULTR_API_KEY")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Additional configuration environment variables:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "\talidns:\tALICLOUD_POLLING_INTERVAL, ALICLOUD_PROPAGATION_TIMEOUT, ALICLOUD_TTL, ALICLOUD_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tauroradns:\tAURORA_POLLING_INTERVAL, AURORA_PROPAGATION_TIMEOUT, AURORA_TTL")
	fmt.Fprintln(w, "\tazure:\tAZURE_POLLING_INTERVAL, AZURE_PROPAGATION_TIMEOUT, AZURE_TTL, AZURE_METADATA_ENDPOINT")
	fmt.Fprintln(w, "\tbluecat:\tBLUECAT_POLLING_INTERVAL, BLUECAT_PROPAGATION_TIMEOUT, BLUECAT_TTL, BLUECAT_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tcloudflare:\tCLOUDFLARE_POLLING_INTERVAL, CLOUDFLARE_PROPAGATION_TIMEOUT, CLOUDFLARE_TTL, CLOUDFLARE_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tcloudxns:\tCLOUDXNS_POLLING_INTERVAL, CLOUDXNS_PROPAGATION_TIMEOUT, CLOUDXNS_TTL, CLOUDXNS_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tdigitalocean:\tDO_POLLING_INTERVAL, DO_PROPAGATION_TIMEOUT, DO_TTL, DO_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tdnsimple:\tDNSIMPLE_TTL, DNSIMPLE_POLLING_INTERVAL, DNSIMPLE_PROPAGATION_TIMEOUT")
	fmt.Fprintln(w, "\tdnsmadeeasy:\tDNSMADEEASY_POLLING_INTERVAL, DNSMADEEASY_PROPAGATION_TIMEOUT, DNSMADEEASY_TTL, DNSMADEEASY_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tdnspod:\tDNSPOD_POLLING_INTERVAL, DNSPOD_PROPAGATION_TIMEOUT, DNSPOD_TTL, DNSPOD_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tdreamhost:\tDREAMHOST_POLLING_INTERVAL, DREAMHOST_PROPAGATION_TIMEOUT, DREAMHOST_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tduckdns:\tDUCKDNS_POLLING_INTERVAL, DUCKDNS_PROPAGATION_TIMEOUT, DUCKDNS_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tdyn:\tDYN_POLLING_INTERVAL, DYN_PROPAGATION_TIMEOUT, DYN_TTL, DYN_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\texoscale:\tEXOSCALE_POLLING_INTERVAL, EXOSCALE_PROPAGATION_TIMEOUT, EXOSCALE_TTL, EXOSCALE_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tfastdns:\tAKAMAI_POLLING_INTERVAL, AKAMAI_PROPAGATION_TIMEOUT, AKAMAI_TTL")
	fmt.Fprintln(w, "\tgandi:\tGANDI_POLLING_INTERVAL, GANDI_PROPAGATION_TIMEOUT, GANDI_TTL, GANDI_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tgandiv5:\tGANDIV5_POLLING_INTERVAL, GANDIV5_PROPAGATION_TIMEOUT, GANDIV5_TTL, GANDIV5_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tgcloud:\tGCE_POLLING_INTERVAL, GCE_PROPAGATION_TIMEOUT, GCE_TTL")
	fmt.Fprintln(w, "\tglesys:\tGLESYS_POLLING_INTERVAL, GLESYS_PROPAGATION_TIMEOUT, GLESYS_TTL, GLESYS_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tgodaddy:\tGODADDY_POLLING_INTERVAL, GODADDY_PROPAGATION_TIMEOUT, GODADDY_TTL, GODADDY_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\thostingde:\tHOSTINGDE_POLLING_INTERVAL, HOSTINGDE_PROPAGATION_TIMEOUT, HOSTINGDE_TTL, HOSTINGDE_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tiij:\tIIJ_POLLING_INTERVAL, IIJ_PROPAGATION_TIMEOUT, IIJ_TTL")
	fmt.Fprintln(w, "\tlightsail:\tLIGHTSAIL_POLLING_INTERVAL, LIGHTSAIL_PROPAGATION_TIMEOUT")
	fmt.Fprintln(w, "\tlinode:\tLINODE_POLLING_INTERVAL, LINODE_TTL, LINODE_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tlinodev4:\tLINODE_POLLING_INTERVAL, LINODE_TTL, LINODE_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tnamecheap:\tNAMECHEAP_POLLING_INTERVAL, NAMECHEAP_PROPAGATION_TIMEOUT, NAMECHEAP_TTL, NAMECHEAP_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tnamedotcom:\tNAMECOM_POLLING_INTERVAL, NAMECOM_PROPAGATION_TIMEOUT, NAMECOM_TTL, NAMECOM_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tnetcup:\tNETCUP_POLLING_INTERVAL, NETCUP_PROPAGATION_TIMEOUT, NETCUP_TTL, NETCUP_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tnifcloud:\tNIFCLOUD_POLLING_INTERVAL, NIFCLOUD_PROPAGATION_TIMEOUT, NIFCLOUD_TTL, NIFCLOUD_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tns1:\tNS1_POLLING_INTERVAL, NS1_PROPAGATION_TIMEOUT, NS1_TTL, NS1_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\totc:\tOTC_POLLING_INTERVAL, OTC_PROPAGATION_TIMEOUT, OTC_TTL, OTC_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tovh:\tOVH_POLLING_INTERVAL, OVH_PROPAGATION_TIMEOUT, OVH_TTL, OVH_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\tpdns:\tPDNS_POLLING_INTERVAL, PDNS_PROPAGATION_TIMEOUT, PDNS_TTL, PDNS_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\trackspace:\tRACKSPACE_POLLING_INTERVAL, RACKSPACE_PROPAGATION_TIMEOUT, RACKSPACE_TTL, RACKSPACE_HTTP_TIMEOUT")
	fmt.Fprintln(w, "\trfc2136:\tRFC2136_POLLING_INTERVAL, RFC2136_PROPAGATION_TIMEOUT, RFC2136_TTL")
	fmt.Fprintln(w, "\troute53:\tAWS_POLLING_INTERVAL, AWS_PROPAGATION_TIMEOUT, AWS_TTL")
	fmt.Fprintln(w, "\tsakuracloud:\tSAKURACLOUD_POLLING_INTERVAL, SAKURACLOUD_PROPAGATION_TIMEOUT, SAKURACLOUD_TTL")
	fmt.Fprintln(w, "\tstackpath:\tSTACKPATH_POLLING_INTERVAL, STACKPATH_PROPAGATION_TIMEOUT, STACKPATH_TTL")
	fmt.Fprintln(w, "\tvegadns:\tVEGADNS_POLLING_INTERVAL, VEGADNS_PROPAGATION_TIMEOUT, VEGADNS_TTL")
	fmt.Fprintln(w, "\tvultr:\tVULTR_POLLING_INTERVAL, VULTR_PROPAGATION_TIMEOUT, VULTR_TTL, VULTR_HTTP_TIMEOUT")

	w.Flush()

	fmt.Println(`
For a more detailed explanation of a DNS provider's credential variables,
please consult their online documentation.`)

	return nil
}
