package cli

import (
	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "shelf",
		Short:         "Shelf Go rewrite",
		Version:       app.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("config", "", "Path to config.yaml")
	root.PersistentFlags().String("vault", "", "Path to encrypted vault")

	root.AddCommand(newCompletionCmd())
	root.AddCommand(newSetupCmd())
	root.AddCommand(newVaultCmd())
	root.AddCommand(newManagerCmd())
	root.AddCommand(newSecretCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newProjectCmd())
	return root
}

func loadVault(cmd *cobra.Command) (config.Runtime, *vault.Vault, error) {
	configPath, _ := cmd.Flags().GetString("config")
	vaultPath, _ := cmd.Flags().GetString("vault")
	return app.LoadVault(configPath, vaultPath)
}

func loadRuntime(cmd *cobra.Command) (config.Runtime, *vault.Store, error) {
	configPath, _ := cmd.Flags().GetString("config")
	vaultPath, _ := cmd.Flags().GetString("vault")
	return app.LoadRuntime(configPath, vaultPath)
}

func updateVault(cmd *cobra.Command, fn func(*vault.Store) error) error {
	configPath, _ := cmd.Flags().GetString("config")
	vaultPath, _ := cmd.Flags().GetString("vault")
	return app.UpdateVault(configPath, vaultPath, fn)
}

func readVault(cmd *cobra.Command, fn func(*vault.Store) error) error {
	configPath, _ := cmd.Flags().GetString("config")
	vaultPath, _ := cmd.Flags().GetString("vault")
	return app.ReadVault(configPath, vaultPath, fn)
}
