package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	config  *Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "icu",
	Short: "ICU - Internal Catalog Utility for satellite data",
	Long: `ICU is a CLI tool for fetching and managing satellite catalog data
, including TLE (Two-Line Element) and SATCAT
(Satellite Catalog) information.`,
	// Default behavior: show stats
	Run: func(cmd *cobra.Command, args []string) {
		statsCmd.Run(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.icu/config.yaml)")
}

func initConfig() {
	var err error
	config, err = InitConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}
}
