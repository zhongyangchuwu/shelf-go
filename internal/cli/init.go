package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func newSetupCmd() *cobra.Command {
	return newInitFilesCmd("setup", "Set up Shelf config and encrypted vault")
}

func newVaultInitCmd() *cobra.Command {
	return newInitFilesCmd("init", "Initialize config and encrypted vault")
}

func newInitFilesCmd(use, short string) *cobra.Command {
	var force bool
	var vaultPath string
	var recipients []string
	var identityPaths []string
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
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
			if !force && initConfigExists(configPath) && initFlagsWereProvided(cmd) {
				return fmt.Errorf("config already exists; rerun with --force to update vault config")
			}
			cfg := initConfig{VaultPath: vaultPath, Recipients: recipients, IdentityPaths: identityPaths}
			if err := cfg.fill(cmd, configPath); err != nil {
				return err
			}
			configCreated, err := app.EnsureConfigFile(configPath, app.InitConfig{VaultPath: cfg.VaultPath, Recipients: cfg.Recipients, IdentityPaths: cfg.IdentityPaths}, force)
			if err != nil {
				return err
			}
			runtime, err := app.ResolveRuntime(configPath, "")
			if err != nil {
				return err
			}
			vaultCreated, err := app.EnsureVaultForRuntime(runtime)
			if err != nil {
				return err
			}
			label := map[bool]string{true: "created", false: "exists"}
			fmt.Fprintf(cmd.OutOrStdout(), "vault:  %s (%s)\n", runtime.VaultPath, label[vaultCreated])
			fmt.Fprintf(cmd.OutOrStdout(), "config: %s (%s)\n", runtime.ConfigPath, label[configCreated])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite config file")
	cmd.Flags().StringVar(&vaultPath, "vault-path", "", "Path to encrypted vault")
	cmd.Flags().StringArrayVar(&recipients, "recipient", nil, "Age recipient")
	cmd.Flags().StringArrayVar(&identityPaths, "identity", nil, "Age identity path")
	return cmd
}

func initConfigExists(configPath string) bool {
	_, err := os.Stat(configPath)
	return err == nil
}

func initFlagsWereProvided(cmd *cobra.Command) bool {
	flags := cmd.Flags()
	return flags.Changed("vault-path") || flags.Changed("recipient") || flags.Changed("identity")
}

type initConfig struct {
	VaultPath     string
	Recipients    []string
	IdentityPaths []string
}

func (c *initConfig) fill(cmd *cobra.Command, configPath string) error {
	if runtime, err := app.ResolveRuntime(configPath, ""); err == nil {
		if c.VaultPath == "" {
			c.VaultPath = runtime.VaultPath
		}
		if len(c.IdentityPaths) == 0 {
			c.IdentityPaths = runtime.IdentityPaths
		}
		if len(c.Recipients) == 0 {
			c.Recipients = runtime.Recipients
		}
	}
	if c.VaultPath == "" {
		c.VaultPath = promptDefault(cmd, "Vault path", app.DefaultVaultPath())
	}
	if len(c.IdentityPaths) == 0 {
		identityPath := promptDefault(cmd, "Age identity path", filepath.Join(filepath.Dir(configPath), "identity.txt"))
		c.IdentityPaths = []string{identityPath}
	}
	if len(c.Recipients) == 0 {
		line := promptDefault(cmd, "Age recipient (blank to generate from identity)", "")
		c.Recipients = splitCSV(line)
	}
	if len(c.Recipients) == 0 {
		identity, err := app.EnsureInitIdentity(c.IdentityPaths[0])
		if err != nil {
			return err
		}
		c.Recipients = []string{identity.Recipient()}
	}
	if c.VaultPath == "" {
		return fmt.Errorf("vault path is required")
	}
	if len(c.Recipients) == 0 {
		return fmt.Errorf("at least one age recipient is required")
	}
	if len(c.IdentityPaths) == 0 || c.IdentityPaths[0] == "" {
		return fmt.Errorf("at least one age identity path is required")
	}
	return nil
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

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
