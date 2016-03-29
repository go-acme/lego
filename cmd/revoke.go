package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// revokeCmd represents the revoke command
var revokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a certificate",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("revoke called")
	},
}

func init() {
	RootCmd.AddCommand(revokeCmd)
    
}
