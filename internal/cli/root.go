package cli

import (
	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
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

func runtimePaths(cmd *cobra.Command) (string, string) {
	configPath, _ := cmd.Flags().GetString("config")
	vaultPath, _ := cmd.Flags().GetString("vault")
	return configPath, vaultPath
}
