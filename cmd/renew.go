package cmd

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"time"

	"github.com/xenolf/lego/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/xenolf/lego/acme"
)

func renewHandler(cmd *cobra.Command, args []string) {
	conf, _, client := utils.Setup(RootCmd)

    domains, err := RootCmd.PersistentFlags().GetStringSlice("domains")
    if err != nil {
        logger().Fatalln(err.Error())
    }
    
	if len(domains) <= 0 {
		logger().Fatal("Please specify at least one domain.")
	}

	domain := domains[0]

	// load the cert resource from files.
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certPath := path.Join(conf.CertPath(), domain+".crt")
	privPath := path.Join(conf.CertPath(), domain+".key")
	metaPath := path.Join(conf.CertPath(), domain+".json")

	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		logger().Fatalf("Error while loading the certificate for domain %s\n\t%s", domain, err.Error())
	}
    
    days, err := cmd.PersistentFlags().GetInt("days")
    if err != nil {
        logger().Fatalln(err.Error())
    }
    
	if days > 0 {
		expTime, err := acme.GetPEMCertExpiration(certBytes)
		if err != nil {
			logger().Printf("Could not get Certification expiration for domain %s", domain)
		}

		if int(expTime.Sub(time.Now()).Hours()/24.0) > days {
			return
		}
	}

	metaBytes, err := ioutil.ReadFile(metaPath)
	if err != nil {
		logger().Fatalf("Error while loading the meta data for domain %s\n\t%s", domain, err.Error())
	}

	var certRes acme.CertificateResource
	err = json.Unmarshal(metaBytes, &certRes)
	if err != nil {
		logger().Fatalf("Error while marshalling the meta data for domain %s\n\t%s", domain, err.Error())
	}

    reusekey, err := cmd.PersistentFlags().GetBool("reuse-key")
    if err != nil {
        logger().Fatalln(err.Error())
    }
    
	if reusekey {
		keyBytes, err := ioutil.ReadFile(privPath)
		if err != nil {
			logger().Fatalf("Error while loading the private key for domain %s\n\t%s", domain, err.Error())
		}
		certRes.PrivateKey = keyBytes
	}

	certRes.Certificate = certBytes

    nobundle, err := cmd.PersistentFlags().GetBool("no-bundle")
    if err != nil {
        logger().Fatalln(err.Error())
    }
    
	newCert, err := client.RenewCertificate(certRes, !nobundle)
	if err != nil {
		logger().Fatalf("%s", err.Error())
	}

	utils.SaveCertRes(newCert, conf)
}

// renewCmd represents the renew command
var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew a certificate",
	Long:  ``,
	Run:   renewHandler,
}

func init() {
	RootCmd.AddCommand(renewCmd)

	renewCmd.PersistentFlags().Int("days", 0, "The number of days left on a certificate to renew it.")
	renewCmd.PersistentFlags().Bool("resuse-key", false, "Used to indicate you want to reuse your current private key for the new certificate.")
	renewCmd.PersistentFlags().Bool("no-bundle", false, "Do not create a certificate bundle by adding the issuers certificate to the new certificate.")

}
