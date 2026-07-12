package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func newSetupCmd(appSvc *app.App) *cobra.Command {
	var force bool
	var vaultPath string
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up Shelf config and item vault",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPathFlag, _ := cmd.Flags().GetString("config")
			vaultFlag, _ := cmd.Flags().GetString("vault")
			configPath, err := app.ResolveInitConfigPath(configPathFlag)
			if err != nil {
				return err
			}
			if vaultPath == "" {
				vaultPath = vaultFlag
			}
			if vaultPath == "" {
				vaultPath = promptDefault(cmd, "Vault path", app.DefaultVaultPath())
			}
			if vaultPath == "" {
				return fmt.Errorf("vault path is required")
			}
			vaultPath, err = app.ExpandInitPath(vaultPath)
			if err != nil {
				return err
			}
			configCreated, err := app.EnsureConfigFile(configPath, app.InitConfig{VaultPath: vaultPath}, force)
			if err != nil {
				return err
			}
			runtime, err := app.ResolveRuntime(configPath, "")
			if err != nil {
				return err
			}
			vaultCreated, err := appSvc.EnsureVaultForRuntime(runtime)
			if err != nil {
				return err
			}
			vaultStatus := "exists"
			if vaultCreated {
				vaultStatus = "created"
			}
			configStatus := "exists"
			if configCreated {
				configStatus = "created"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "vault:  %s (%s)\n", runtime.VaultPath, vaultStatus)
			fmt.Fprintf(cmd.OutOrStdout(), "config: %s (%s)\n", runtime.ConfigPath, configStatus)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite config file")
	cmd.Flags().StringVar(&vaultPath, "vault-path", "", "Path to Shelf item vault JSON")
	return cmd
}

func promptDefault(cmd *cobra.Command, label, fallback string) string {
	in := bufio.NewReader(cmd.InOrStdin())
	if fallback == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "%s: ", label)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "%s [%s]: ", label, fallback)
	}
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return fallback
	}
	return line
}
