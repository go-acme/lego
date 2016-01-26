package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/xenolf/lego/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xenolf/lego/acme"
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

// setup checks every parameter for lego, creates the .lego path and generates a client for ACME
func setup(cmd *cobra.Command) (*utils.Configuration, *utils.Account, *acme.Client) {
	path := cmd.Flag("path").Value.String()
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
