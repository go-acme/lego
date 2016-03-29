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
var Logger *log.Logger

func logger() *log.Logger {
	if Logger == nil {
		Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return Logger
}

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lego",
	Short: "Let's Encrypt client written in Go",
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
	cobra.OnInitialize(initConfig)
	cwd, err := os.Getwd()
	if err != nil {
		logger().Fatal("Could not determine current working directory. Please pass --path.")
	}
	defaultPath := path.Join(cwd, ".lego")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.PersistentFlags().StringSliceP("domains", "d", nil, "Add domains to the process")
	RootCmd.PersistentFlags().StringP("server", "s", "https://acme-v01.api.letsencrypt.org/directory", "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.")
	RootCmd.PersistentFlags().StringP("email", "m", "", "Email used for registration and recovery contact.")
	RootCmd.PersistentFlags().BoolP("accept-tos", "a", false, "By setting this flag to true you indicate that you accept the current Let's Encrypt terms of service.")
	RootCmd.PersistentFlags().StringP("key-type", "k", "rsa2048", "Key type to use for private keys. Supported: rsa2048, rsa4096, rsa8192, ec256, ec384")
	RootCmd.PersistentFlags().String("path", defaultPath, "Directory to use for storing the data")
	RootCmd.PersistentFlags().StringSliceP("exclude", "x", nil, "Explicitly disallow solvers by name from being used. Solvers: \"http-01\", \"tls-sni-01\".")
	RootCmd.PersistentFlags().String("webroot", "", "Set the webroot folder to use for HTTP based challenges to write directly in a file in .well-known/acme-challenge")
	RootCmd.PersistentFlags().String("http", "", "Set the port and interface to use for HTTP based challenges to listen on. Supported: interface:port or :port")
	RootCmd.PersistentFlags().String("tls", "", "Set the port and interface to use for TLS based challenges to listen on. Supported: interface:port or :port")
	RootCmd.PersistentFlags().String("dns", "", "Solve a DNS challenge using the specified provider. Disables all other challenges. Run 'lego dnshelp' for help on usage.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
