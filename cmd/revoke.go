package cmd

import (
	"io/ioutil"
	"path"

	"github.com/xenolf/lego/cmd/utils"
	"github.com/spf13/cobra"
)

func revokeHandler(cmd *cobra.Command, args []string) {
	conf, _, client := utils.Setup(RootCmd)

	err := utils.CheckFolder(conf.CertPath())
	if err != nil {
		logger().Fatalf("Could not check/create path: %s", err.Error())
	}

    domains, err := RootCmd.PersistentFlags().GetStringSlice("domains")
    if err != nil {
        logger().Fatalln(err.Error())
    }
    
	for _, domain := range domains {
		logger().Printf("Trying to revoke certificate for domain %s", domain)

		certPath := path.Join(conf.CertPath(), domain+".crt")
		certBytes, err := ioutil.ReadFile(certPath)

		err = client.RevokeCertificate(certBytes)
		if err != nil {
			logger().Fatalf("Error while revoking the certificate for domain %s\n\t%s", domain, err.Error())
		} else {
			logger().Print("Certificate was revoked.")
		}
	}
}

// revokeCmd represents the revoke command
var revokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a certificate",
	Long:  ``,
	Run: revokeHandler,
}

func init() {
	RootCmd.AddCommand(revokeCmd)

}
