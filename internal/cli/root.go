package cli

import (
	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/clipboard"
)

func NewRootCmd(appSvc *app.App) *cobra.Command {
	return newRootCmd(appSvc, clipboard.New())
}

func newRootCmd(appSvc *app.App, clipboardWriter clipboard.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "shelf",
		Short:         "Shelf Go rewrite",
		Version:       app.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("config", "", "Path to config.yaml")
	root.PersistentFlags().String("vault", "", "Path to Shelf item vault JSON")

	root.AddCommand(newCompletionCmd())
	root.AddCommand(newSetupCmd(appSvc))
	root.AddCommand(newSecretCmd(appSvc, clipboardWriter))
	root.AddCommand(newEnvCmd(appSvc))
	root.AddCommand(newTUICmd(appSvc, clipboardWriter))
	root.AddCommand(newDoctorCmd(appSvc))
	root.AddCommand(newProjectCmd(appSvc))
	return root
}

func runtimePaths(cmd *cobra.Command) (string, string) {
	configPath, _ := cmd.Flags().GetString("config")
	vaultPath, _ := cmd.Flags().GetString("vault")
	return configPath, vaultPath
}
