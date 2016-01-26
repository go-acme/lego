package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lego",
	Short: "Let's Encrypt client and ACME library written in Go",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Could not determine current working directory. Please pass --path.")
	}
	defaultPath := path.Join(cwd, ".lego")
    
	cobra.OnInitialize(initConfig)


	RootCmd.PersistentFlags().StringSliceP("domains", "d", nil, "Add domains to the process")
	RootCmd.PersistentFlags().StringP("email", "m", "", "Email used for registration and recovery contact.")
    RootCmd.PersistentFlags().StringSliceP("exclude", "x", nil, `Explicitly disallow solvers by name from being used. Solvers: "http-01", "tls-sni-01".`)
    
    RootCmd.PersistentFlags().StringP("server", "s", "https://acme-v01.api.letsencrypt.org/directory", "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.")
	RootCmd.PersistentFlags().IntP("rsa-key-size", "B", 2048, "Size of the RSA key.")
	RootCmd.PersistentFlags().String("path", defaultPath, "Directory to use for storing the data")
    RootCmd.PersistentFlags().String("http", "", "Set the port and interface to use for HTTP based challenges to listen on. Supported: interface:port or :port.")
    RootCmd.PersistentFlags().String("tls", "", "Set the port and interface to use for TLS based challenges to listen on. Supported: interface:port or :port.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	//viper.SetConfigName(".lego") // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
