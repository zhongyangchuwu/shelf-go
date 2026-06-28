package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/manager"
	vaultsvc "github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func newVaultCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "vault", Short: "Manage encrypted vault"}
	cmd.AddCommand(newVaultInitCmd())
	cmd.AddCommand(newMigrateCmd())
	cmd.AddCommand(newVaultStatusCmd())
	return cmd
}

func newVaultStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"check"},
		Short:   "Check encrypted vault status",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := newDiagnosticReport(cmd.OutOrStdout())
			configPath, _ := cmd.Flags().GetString("config")
			vaultPath, _ := cmd.Flags().GetString("vault")
			runtime, err := config.Resolve(configPath, vaultPath)
			if err != nil {
				report.fail("config", err.Error())
				return report.err("vault status")
			}

			report.write(vaultsvc.Status(runtime))
			return report.err("vault status")
		},
	}
}

func newManagerCmd() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "manager",
		Short: "Open the local Shelf manager",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, vault, err := loadVault(cmd)
			if err != nil {
				return err
			}
			listener, err := app.ListenLoopback(addr)
			if err != nil {
				return err
			}
			defer listener.Close()
			token, err := app.ManagerToken()
			if err != nil {
				return err
			}
			server, err := manager.NewServer(vault, token, listener.Addr().String())
			if err != nil {
				return err
			}
			httpServer := &http.Server{Handler: server.Handler()}
			errCh := make(chan error, 1)
			go func() {
				if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
					errCh <- err
					return
				}
				errCh <- nil
			}()
			fmt.Fprintf(cmd.OutOrStdout(), "manager: http://%s/?token=%s\n", listener.Addr().String(), token)
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(sigCh)
			select {
			case err := <-errCh:
				return err
			case <-sigCh:
				return httpServer.Close()
			}
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:0", "Loopback address to listen on")
	return cmd
}
