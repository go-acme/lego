package cmd

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/xenolf/lego/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/xenolf/lego/acme"
)

func saveCertRes(certRes acme.CertificateResource, conf *utils.Configuration) {
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certOut := path.Join(conf.CertPath(), certRes.Domain+".crt")
	privOut := path.Join(conf.CertPath(), certRes.Domain+".key")
	metaOut := path.Join(conf.CertPath(), certRes.Domain+".json")

	err := ioutil.WriteFile(certOut, certRes.Certificate, 0600)
	if err != nil {
		log.Fatalf("Unable to save Certificate for domain %s\n\t%s", certRes.Domain, err.Error())
	}

	err = ioutil.WriteFile(privOut, certRes.PrivateKey, 0600)
	if err != nil {
		log.Fatalf("Unable to save PrivateKey for domain %s\n\t%s", certRes.Domain, err.Error())
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		log.Fatalf("Unable to marshal CertResource for domain %s\n\t%s", certRes.Domain, err.Error())
	}

	err = ioutil.WriteFile(metaOut, jsonBytes, 0600)
	if err != nil {
		log.Fatalf("Unable to save CertResource for domain %s\n\t%s", certRes.Domain, err.Error())
	}
}

func setup(cmd *cobra.Command) (*utils.Configuration, *utils.Account, *acme.Client) {
	path := cmd.Flag("path").Value.String()
	log.Println("checking Path", path)
	utils.CheckFolder(path)
	conf := utils.NewConfiguration(cmd)
	email, err := cmd.PersistentFlags().GetString("email")
	if err != nil || len(email) == 0 {
		log.Fatalln("You have to pass an account (email address) to the program using --email or -m")
	}

	acc := utils.NewAccount(email, conf)

	server, err := cmd.PersistentFlags().GetString("server")
	if err != nil {
		log.Fatalln("Error on getting server value")
	}

	client, err := acme.NewClient(server, acc, conf.RsaBits())
	if err != nil {
		log.Fatalf("Could not create client: %s", err.Error())
	}

	if exclude, _ := cmd.PersistentFlags().GetStringSlice("exclude"); len(exclude) > 0 {
		client.ExcludeChallenges(conf.ExcludedSolvers())
	}

	http, err := RootCmd.PersistentFlags().GetString("http")
	if err != nil {
		log.Fatalln(err.Error())
	}
	if len(http) > 0 {
		client.SetHTTPAddress(http)
	}

	tls, err := RootCmd.PersistentFlags().GetString("tls")
	if err != nil {
		log.Fatalln(err.Error())
	}
	if len(tls) > 0 {
		client.SetTLSAddress(tls)
	}

	log.Println("end of setup")
	return conf, acc, client
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs lego",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		conf, acc, client := setup(RootCmd)
		if acc.Registration == nil {
			reg, err := client.Register()
			if err != nil {
				log.Fatalf("Could not complete registration\n\t%s", err.Error())
			}

			email, err := RootCmd.PersistentFlags().GetString("email")
			if err != nil {
				log.Fatalf(err.Error())
			}

			acc.Registration = reg
			acc.Save()

			log.Print("!!!! HEADS UP !!!!")
			log.Printf(`
		Your account credentials have been saved in your Let's Encrypt
		configuration directory at "%s".
		You should make a secure backup	of this folder now. This
		configuration directory will also contain certificates and
		private keys obtained from Let's Encrypt so making regular
		backups of this folder is ideal.`, conf.AccountPath(email))

		}

		if acc.Registration.Body.Agreement == "" {
			reader := bufio.NewReader(os.Stdin)
			log.Printf("Please review the TOS at %s", acc.Registration.TosURL)

			for {
				log.Println("Do you accept the TOS? Y/n")
				text, err := reader.ReadString('\n')
				if err != nil {
					log.Fatalf("Could not read from console -> %s", err.Error())
				}

				text = strings.Trim(text, "\r\n")

				if text == "n" {
					log.Fatal("You did not accept the TOS. Unable to proceed.")
				}

				if text == "Y" || text == "y" || text == "" {
					err = client.AgreeToTOS()
					if err != nil {
						log.Fatalf("Could not agree to tos -> %s", err)
					}
					acc.Save()
					break
				}

				log.Println("Your input was invalid. Please answer with one of Y/y, n or by pressing enter.")
			}
		}

		domains, err := RootCmd.PersistentFlags().GetStringSlice("domains")
		if err != nil {
			log.Fatalln(err.Error())
		}
		if len(domains) == 0 {
			log.Fatal("Please specify --domains or -d")
		}

		cert, failures := client.ObtainCertificate(domains, true, nil)
		if len(failures) > 0 {
			for k, v := range failures {
				log.Printf("[%s] Could not obtain certificates\n\t%s", k, v.Error())
			}

			// Make sure to return a non-zero exit code if ObtainSANCertificate
			// returned at least one error. Due to us not returning partial
			// certificate we can just exit here instead of at the end.
			os.Exit(1)
		}

		err = utils.CheckFolder(conf.CertPath())
		if err != nil {
			log.Fatalf("Cound not check/create path: %s", err.Error())
		}

		saveCertRes(cert, conf)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
