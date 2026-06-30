package cli

import (
	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func NewRootCmd(appSvc *app.App) *cobra.Command {
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
	root.AddCommand(newSetupCmd(appSvc))
	root.AddCommand(newVaultCmd(appSvc))
	root.AddCommand(newManagerCmd(appSvc))
	root.AddCommand(newSecretCmd(appSvc))
	root.AddCommand(newDoctorCmd(appSvc))
	root.AddCommand(newProjectCmd(appSvc))
	return root
}

func runtimePaths(cmd *cobra.Command) (string, string) {
	configPath, _ := cmd.Flags().GetString("config")
	vaultPath, _ := cmd.Flags().GetString("vault")
	return configPath, vaultPath
}
