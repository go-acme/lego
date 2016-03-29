// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"

	"github.com/xenolf/lego/cmd/utils"
	"github.com/spf13/cobra"
)

func runHandler(cmd *cobra.Command, args []string) {
	conf, acc, client := utils.Setup(RootCmd)
	if acc.Registration == nil {
		reg, err := client.Register()
		if err != nil {
			logger().Fatalf("Could not complete registration\n\t%s", err.Error())
		}

		acc.Registration = reg
		acc.Save()
		email, err := RootCmd.PersistentFlags().GetString("email")
		if err != nil {
			logger().Fatalln(err.Error())
		}
		logger().Print("!!!! HEADS UP !!!!")
		logger().Printf(`
		Your account credentials have been saved in your Let's Encrypt
		configuration directory at "%s".
		You should make a secure backup	of this folder now. This
		configuration directory will also contain certificates and
		private keys obtained from Let's Encrypt so making regular
		backups of this folder is ideal.`, conf.AccountPath(email))

	}

	// If the agreement URL is empty, the account still needs to accept the LE TOS.
	if acc.Registration.Body.Agreement == "" {
		utils.HandleTOS(RootCmd, client, acc)
	}

	domains, err := RootCmd.PersistentFlags().GetStringSlice("domains")
	if err != nil {
		logger().Fatalln(err.Error())
	}

	if len(domains) == 0 {
		logger().Fatal("Please specify --domains or -d")
	}

	nobundle, err := cmd.PersistentFlags().GetBool("no-bundle")
	if err != nil {
		logger().Fatalln(err.Error())
	}
	cert, failures := client.ObtainCertificate(domains, !nobundle, nil)
	if len(failures) > 0 {
		for k, v := range failures {
			logger().Printf("[%s] Could not obtain certificates\n\t%s", k, v.Error())
		}

		// Make sure to return a non-zero exit code if ObtainSANCertificate
		// returned at least one error. Due to us not returning partial
		// certificate we can just exit here instead of at the end.
		os.Exit(1)
	}

	err = utils.CheckFolder(conf.CertPath())
	if err != nil {
		logger().Fatalf("Could not check/create path: %s", err.Error())
	}

	utils.SaveCertRes(cert, conf)
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Register an account, then create and install a certificate",
	Long:  ``,
	Run:   runHandler,
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().Bool("no-bundle", false, "Do not create a certificate bundle by adding the issuers certificate to the new certificate.")

}
