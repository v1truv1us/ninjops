package app

import (
	"fmt"
	"os"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "ninjops",
	Short: "Agentic orchestration CLI for Invoice Ninja quote/invoice lifecycle",
	Long: `Ninjops is a production-ready CLI for managing the full quote/invoice
lifecycle for Invoice Ninja v5. It provides deterministic-first generation
with templates, optional AI-powered assistance, and safe Invoice Ninja integration.`,
	Version: "1.0.0",
}

func Execute() error {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file override (default: ~/.config/ninjops/config.json, legacy files still supported)")

	addCommands()
	addSubcommands()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}

func addCommands() {
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newConfigureCmd())
	rootCmd.AddCommand(newNewCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newShowCmd())
	rootCmd.AddCommand(newConvertCmd())
	rootCmd.AddCommand(newEditCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newAssistCmd())
	rootCmd.AddCommand(newServeCmd())
}

func addSubcommands() {
	ninjaCmd := &cobra.Command{
		Use:   "ninja",
		Short: "Invoice Ninja operations",
	}

	ninjaCmd.AddCommand(newNinjaTestCmd())
	ninjaCmd.AddCommand(newNinjaPullCmd())
	ninjaCmd.AddCommand(newNinjaSyncCmd())
	ninjaCmd.AddCommand(newNinjaDiffCmd())

	rootCmd.AddCommand(ninjaCmd)
}
