package main

import (
	"log"
	"os"

	"github.com/dcsunny/socks5"
	"github.com/spf13/cobra"
)

var (
	listenAddr     string
	downProxy      string
	useSystemProxy bool
	username       string
	password       string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "socks5",
	Short: "A SOCKS5 proxy server",
	Long:  `A SOCKS5 proxy server that can use a downstream proxy.`,
	Run: func(cmd *cobra.Command, args []string) {
		s := socks5.NewServer(useSystemProxy, listenAddr, downProxy, username, password)
		s.Run()
	},
}

func main() {
	Execute()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&listenAddr, "listen", "l", "0.0.0.0:21080", "Address to listen on")
	rootCmd.Flags().StringVarP(&downProxy, "down-proxy", "d", "", "Downstream proxy type (socks5, http)")
	rootCmd.Flags().BoolVarP(&useSystemProxy, "system-proxy", "s", true, "use system proxy")
	rootCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
}
