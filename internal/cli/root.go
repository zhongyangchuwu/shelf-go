package cli

import (
	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
	"github.com/zhongyangchuwu/shelf-go/internal/version"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "shelf",
		Short:         "Shelf Go rewrite",
		Version:       version.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("config", "", "Path to config.yaml")
	root.PersistentFlags().String("vault", "", "Path to encrypted vault")

	root.AddCommand(newCompletionCmd())
	root.AddCommand(newSetupCmd())
	root.AddCommand(newVaultCmd())
	root.AddCommand(newSecretCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newProjectCmd())
	return root
}

func loadVault(cmd *cobra.Command) (config.Runtime, *store.Vault, error) {
	configPath, _ := cmd.Flags().GetString("config")
	vaultPath, _ := cmd.Flags().GetString("vault")
	runtime, err := config.Resolve(configPath, vaultPath)
	if err != nil {
		return config.Runtime{}, nil, err
	}
	vault, err := store.NewVault(runtime.VaultPath, store.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
	if err != nil {
		return config.Runtime{}, nil, err
	}
	return runtime, vault, nil
}

func loadRuntime(cmd *cobra.Command) (config.Runtime, *store.Store, error) {
	runtime, vault, err := loadVault(cmd)
	if err != nil {
		return config.Runtime{}, nil, err
	}
	st, err := vault.Load()
	if err != nil {
		return config.Runtime{}, nil, err
	}
	return runtime, st, nil
}

func updateVault(cmd *cobra.Command, fn func(*store.Store) error) error {
	_, vault, err := loadVault(cmd)
	if err != nil {
		return err
	}
	return vault.Update(fn)
}

func readVault(cmd *cobra.Command, fn func(*store.Store) error) error {
	_, vault, err := loadVault(cmd)
	if err != nil {
		return err
	}
	return vault.Read(fn)
}
