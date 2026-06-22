package cli

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/manager"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

func newVaultCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "vault", Short: "Manage encrypted vault"}
	cmd.AddCommand(newVaultInitCmd())
	cmd.AddCommand(newMigrateCmd())
	cmd.AddCommand(newVaultStatusCmd())
	cmd.AddCommand(newVaultOpenCmd())
	return cmd
}

func newVaultStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"check"},
		Short:   "Check encrypted vault status",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			vaultPath, _ := cmd.Flags().GetString("vault")
			runtime, err := config.Resolve(configPath, vaultPath)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "fail config (%s)\n", err)
				return fmt.Errorf("vault status found failures")
			}

			failed := false
			fmt.Fprintf(cmd.OutOrStdout(), "ok   config (%s)\n", runtime.ConfigPath)
			fmt.Fprintf(cmd.OutOrStdout(), "ok   vault path (%s)\n", runtime.VaultPath)
			format, err := store.DetectFileFormat(runtime.VaultPath)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "fail vault format (%s)\n", err)
				return fmt.Errorf("vault status found failures")
			}
			switch format {
			case store.FileFormatMissing:
				fmt.Fprintf(cmd.OutOrStdout(), "warn vault format (missing; run shelf vault init or write a secret to create it)\n")
				return nil
			case store.FileFormatEmpty:
				fmt.Fprintf(cmd.OutOrStdout(), "warn vault format (empty; run shelf vault init or write a secret to encrypt it)\n")
				return nil
			case store.FileFormatEncryptedVault:
				fmt.Fprintf(cmd.OutOrStdout(), "ok   vault format (encrypted shelf-vault/v1)\n")
			case store.FileFormatPlaintextStore:
				fmt.Fprintf(cmd.OutOrStdout(), "fail vault format (plaintext JSON store; run shelf vault migrate before using encrypted vault mode)\n")
				failed = true
			case store.FileFormatUnsupportedVault:
				fmt.Fprintf(cmd.OutOrStdout(), "fail vault format (unsupported shelf vault format)\n")
				failed = true
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "fail vault format (unsupported file content)\n")
				failed = true
			}
			if failed {
				return fmt.Errorf("vault status found failures")
			}
			vault, err := store.NewVault(runtime.VaultPath, store.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "fail vault loads (%s)\n", err)
				return fmt.Errorf("vault status found failures")
			}
			if _, err := vault.Load(); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "fail vault loads (%s; check identity_paths or run shelf vault init)\n", err)
				return fmt.Errorf("vault status found failures")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok   vault loads (%s)\n", runtime.VaultPath)
			return nil
		},
	}
}

func newVaultOpenCmd() *cobra.Command {
	return newManagerCmdWithUse("open")
}

func newManagerCmdWithUse(use string) *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   use,
		Short: "Open a localhost vault manager",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, vault, err := loadVault(cmd)
			if err != nil {
				return err
			}
			listener, err := listenLoopback(addr)
			if err != nil {
				return err
			}
			defer listener.Close()
			token, err := managerToken()
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

func listenLoopback(addr string) (net.Listener, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, fmt.Errorf("manager address must be loopback: %s", addr)
	}
	return net.Listen("tcp", addr)
}

func managerToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}
