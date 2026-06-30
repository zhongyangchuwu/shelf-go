package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/manager"
)

func newVaultCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{Use: "vault", Short: "Manage encrypted vault"}
	cmd.AddCommand(newVaultInitCmd(appSvc))
	cmd.AddCommand(newMigrateCmd(appSvc))
	cmd.AddCommand(newVaultImportCmd(appSvc))
	cmd.AddCommand(newVaultStatusCmd(appSvc))
	return cmd
}

func newVaultStatusCmd(appSvc *app.App) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"check"},
		Short:   "Check encrypted vault status",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := newDiagnosticReport(cmd.OutOrStdout())
			configPath, vaultPath := runtimePaths(cmd)
			reportChecks, err := appSvc.ResolveStatus(configPath, vaultPath)
			if err != nil {
				report.fail("config", err.Error())
				return report.err("vault status")
			}

			report.write(reportChecks)
			return report.err("vault status")
		},
	}
}

func newManagerCmd(appSvc *app.App) *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "manager",
		Short: "Open the local Shelf manager",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			runtime, err := manager.Open(appSvc, configPath, vaultPath, addr)
			if err != nil {
				return err
			}
			defer runtime.Close()
			fmt.Fprintf(cmd.OutOrStdout(), "manager: http://%s/?token=%s\n", runtime.Addr(), runtime.Token())
			return runtime.ServeUntilSignal(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:0", "Loopback address to listen on")
	return cmd
}
