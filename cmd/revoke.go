package cmd

import (
	"io/ioutil"
	"log"
	"path"

	"github.com/gianluca311/lego/cmd/utils"
	"github.com/spf13/cobra"
)

// revokeCmd represents the revoke command
var revokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revokes a certificate",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		conf, _, client := setup(RootCmd)

		err := utils.CheckFolder(conf.CertPath())
		if err != nil {
			log.Fatalf("Cound not check/create path: %s", err.Error())
		}
        
        domains, err := RootCmd.PersistentFlags().GetStringSlice("domains")
		if err != nil {
			log.Fatalln(err.Error())
		}
		for _, domain := range domains {
			log.Printf("Trying to revoke certificate for domain %s", domain)

			certPath := path.Join(conf.CertPath(), domain+".crt")
			certBytes, err := ioutil.ReadFile(certPath)

			err = client.RevokeCertificate(certBytes)
			if err != nil {
				log.Fatalf("Error while revoking the certificate for domain %s\n\t%s", domain, err.Error())
			} else {
				log.Print("Certificate was revoked.")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(revokeCmd)
}
