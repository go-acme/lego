package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// renewCmd represents the renew command
var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew a certificate",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("renew called")
	},
}

func init() {
	RootCmd.AddCommand(renewCmd)

	renewCmd.PersistentFlags().Int("days", 0, "The number of days left on a certificate to renew it.")
    renewCmd.PersistentFlags().Bool("resuse-key", false, "Used to indicate you want to reuse your current private key for the new certificate.")
    renewCmd.PersistentFlags().Bool("no-bundle", false, "Do not create a certificate bundle by adding the issuers certificate to the new certificate.")

}
