package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/xenolf/lego/acme"
)

// renewCmd represents the renew command
var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew a certificate",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		conf, _, client := setup(RootCmd)

		domains, err := RootCmd.PersistentFlags().GetStringSlice("domains")
		if err != nil {
			log.Fatalln(err.Error())
		}
		if len(domains) <= 0 {
			log.Fatal("Please specify at least one domain.")
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
			log.Fatalf("Error while loading the certificate for domain %s\n\t%s", domain, err.Error())
		}
        
		days, err := cmd.PersistentFlags().GetInt("days")
		if err != nil {
			log.Fatalln(err.Error())
		}
        
		if days > 0 {
			expTime, err := acme.GetPEMCertExpiration(certBytes)
			if err != nil {
				log.Printf("Could not get Certification expiration for domain %s", domain)
			}

			if int(expTime.Sub(time.Now()).Hours()/24.0) > days {
				return
			}
		}

		metaBytes, err := ioutil.ReadFile(metaPath)
		if err != nil {
			log.Fatalf("Error while loading the meta data for domain %s\n\t%s", domain, err.Error())
		}

		var certRes acme.CertificateResource
		err = json.Unmarshal(metaBytes, &certRes)
		if err != nil {
			log.Fatalf("Error while marshalling the meta data for domain %s\n\t%s", domain, err.Error())
		}

        reuse, err := cmd.PersistentFlags().GetBool("reuse-key")
        if err != nil {
            log.Fatalln(err.Error())
        }
        
		if reuse {
			keyBytes, err := ioutil.ReadFile(privPath)
			if err != nil {
				log.Fatalf("Error while loading the private key for domain %s\n\t%s", domain, err.Error())
			}
			certRes.PrivateKey = keyBytes
		}

		certRes.Certificate = certBytes

		newCert, err := client.RenewCertificate(certRes, true)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		saveCertRes(newCert, conf)
	},
}

func init() {
	RootCmd.AddCommand(renewCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// renewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// renewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	renewCmd.PersistentFlags().Int("days", 0, "The number of days left on a certificate to renew it.")
	renewCmd.PersistentFlags().Bool("reuse-key", false, "Used to indicate you want to reuse your current private key for the new certificate.")
}
