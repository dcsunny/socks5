package main

import (
	"fmt"
	"os"

	"github.com/dcsunny/socks5"
	"github.com/spf13/cobra"
)

var (
	listenAddr string
	downProxy  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "socks5",
	Short: "A SOCKS5 proxy server",
	Long:  `A SOCKS5 proxy server that can use a downstream proxy.`,
	Run: func(cmd *cobra.Command, args []string) {
		socks5.Server(listenAddr, downProxy)
	},
}

func main() {
	Execute()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&listenAddr, "listen", "l", "0.0.0.0:21080", "Address to listen on")
	rootCmd.Flags().StringVarP(&downProxy, "down-proxy", "d", "", "Downstream proxy type (socks5, http)")
}
