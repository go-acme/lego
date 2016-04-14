package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// dnshelpCmd represents the dnshelp command
var dnshelpCmd = &cobra.Command{
	Use:   "dnshelp",
	Short: "Shows additional help for the --dns global option",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
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
		fmt.Fprintln(w, "\tcloudflare:\tCLOUDFLARE_EMAIL, CLOUDFLARE_API_KEY")
		fmt.Fprintln(w, "\tdigitalocean:\tDO_AUTH_TOKEN")
		fmt.Fprintln(w, "\tdnsimple:\tDNSIMPLE_EMAIL, DNSIMPLE_API_KEY")
		fmt.Fprintln(w, "\tgandi:\tGANDI_API_KEY")
		fmt.Fprintln(w, "\tgcloud:\tGCE_PROJECT")
		fmt.Fprintln(w, "\tmanual:\tnone")
		fmt.Fprintln(w, "\tnamecheap:\tNAMECHEAP_API_USER, NAMECHEAP_API_KEY")
		fmt.Fprintln(w, "\trfc2136:\tRFC2136_TSIG_KEY, RFC2136_TSIG_SECRET,\n\t\tRFC2136_TSIG_ALGORITHM, RFC2136_NAMESERVER")
		fmt.Fprintln(w, "\troute53:\tAWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION")
		fmt.Fprintln(w, "\tdyn:\tDYN_CUSTOMER_NAME, DYN_USER_NAME, DYN_PASSWORD")
		fmt.Fprintln(w, "\tvultr:\tVULTR_API_KEY")
		w.Flush()

		fmt.Println(`
For a more detailed explanation of a DNS provider's credential variables,
please consult their online documentation.`)
	},
}

func init() {
	RootCmd.AddCommand(dnshelpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dnshelpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dnshelpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
