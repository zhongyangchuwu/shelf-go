package cli

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
			report := newDiagnosticReport(cmd.OutOrStdout())
			configPath, _ := cmd.Flags().GetString("config")
			vaultPath, _ := cmd.Flags().GetString("vault")
			runtime, err := config.Resolve(configPath, vaultPath)
			if err != nil {
				report.fail("config", err.Error())
				return report.err("vault status")
			}

			report.ok("config", runtime.ConfigPath)
			report.ok("vault path", runtime.VaultPath)
			checkVaultRecipients(report, runtime)
			checkVaultLoads(report, runtime)
			return report.err("vault status")
		},
	}
}

func checkVaultRecipients(report *diagnosticReport, runtime config.Runtime) {
	if len(runtime.Recipients) == 0 {
		report.fail("vault recipients", vaultMissingRecipientsDetail())
		return
	}
	report.ok("vault recipients", fmt.Sprintf("%d configured", len(runtime.Recipients)))
}

func vaultMissingRecipientsDetail() string {
	return "no age recipients configured; run shelf vault init --force --recipient AGE_RECIPIENT --identity PATH before creating or updating secrets"
}

func vaultFormatDetail(format store.FileFormat, path string) string {
	switch format {
	case store.FileFormatMissing:
		return path + " is missing; run shelf vault init or write a secret after configuring recipients"
	case store.FileFormatEmpty:
		return path + " is empty; run shelf vault init or write a secret after configuring recipients"
	case store.FileFormatPlaintextStore:
		return "plaintext JSON store; run shelf vault migrate --from " + path + " --to <vault.age>, update config, then move/delete/archive the plaintext source"
	case store.FileFormatUnsupportedVault:
		return "unsupported shelf vault format; upgrade Shelf if this vault came from a newer version, or restore a compatible encrypted backup"
	default:
		return "unsupported file content; choose a valid vault path or restore a compatible encrypted backup"
	}
}

func vaultLoadErrorDetail(err error) string {
	message := err.Error()
	switch {
	case strings.Contains(message, "no age identity paths"):
		return message + "; add identity_paths in config or run shelf vault init --identity PATH"
	case strings.Contains(message, "read age identity"):
		return message + "; fix identity_paths or identity file permissions"
	case strings.Contains(message, "parse age identity") || strings.Contains(message, "no age identities loaded"):
		return message + "; fix the identity file contents or run shelf vault init --identity PATH"
	case strings.Contains(message, "no configured age identity matched"):
		return message + "; configure the age identity that matches this vault recipient"
	case strings.Contains(message, "could not decrypt vault"):
		return message + "; verify identity_paths match the vault recipient or restore a known-good encrypted backup"
	case strings.Contains(message, "invalid decrypted store"):
		return message + "; restore a known-good encrypted backup"
	default:
		return message
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
