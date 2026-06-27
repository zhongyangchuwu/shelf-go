package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
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
			configPath, err := resolveInitConfigPath(configPathFlag)
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
			configCreated, err := ensureConfigFile(configPath, cfg, force)
			if err != nil {
				return err
			}
			runtime, err := config.Resolve(configPath, "")
			if err != nil {
				return err
			}
			vaultCreated, err := ensureVaultFile(runtime.VaultPath, vault.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
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
	if runtime, err := config.Resolve(configPath, ""); err == nil {
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
		c.VaultPath = promptDefault(cmd, "Vault path", config.DefaultVaultPath)
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
		identity, err := ensureInitIdentity(c.IdentityPaths[0])
		if err != nil {
			return err
		}
		c.Recipients = []string{identity.Recipient().String()}
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

func resolveInitConfigPath(flag string) (string, error) {
	if flag != "" {
		return filepath.Abs(flag)
	}
	if env := os.Getenv("SHELF_CONFIG"); env != "" {
		return filepath.Abs(env)
	}
	return expandInitPath(config.DefaultConfigPath)
}

func ensureInitIdentity(path string) (*age.X25519Identity, error) {
	if bytes, err := os.ReadFile(path); err == nil {
		identities, err := age.ParseIdentities(strings.NewReader(string(bytes)))
		if err != nil {
			return nil, fmt.Errorf("parse age identity %s: %w", path, err)
		}
		for _, identity := range identities {
			if x25519, ok := identity.(*age.X25519Identity); ok {
				return x25519, nil
			}
		}
		return nil, fmt.Errorf("age identity %s contains no X25519 identity", path)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read age identity %s: %w", path, err)
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(identity.String()+"\n"), 0o600); err != nil {
		return nil, err
	}
	return identity, nil
}

func ensureVaultFile(vaultPath string, options vault.VaultOptions) (bool, error) {
	if _, err := os.Stat(vaultPath); err == nil {
		return false, nil
	} else if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	v, err := vault.NewVault(vaultPath, options)
	if err != nil {
		return false, err
	}
	if err := v.Save(&vault.Store{Data: vault.NewData()}); err != nil {
		return false, err
	}
	return true, nil
}

func ensureConfigFile(configPath string, cfg initConfig, force bool) (bool, error) {
	if _, err := os.Stat(configPath); err == nil && !force {
		return false, nil
	}
	vaultPath, err := relativeIfDescendant(cfg.VaultPath, filepath.Dir(configPath))
	if err != nil {
		vaultPath = cfg.VaultPath
	}
	identityPaths := make([]string, 0, len(cfg.IdentityPaths))
	for _, path := range cfg.IdentityPaths {
		rel, err := relativeIfDescendant(path, filepath.Dir(configPath))
		if err != nil {
			rel = path
		}
		identityPaths = append(identityPaths, rel)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "version: 1\nvault_path: %s\nrecipients:\n", vaultPath)
	for _, recipient := range cfg.Recipients {
		fmt.Fprintf(&b, "  - %s\n", recipient)
	}
	b.WriteString("identity_paths:\n")
	for _, path := range identityPaths {
		fmt.Fprintf(&b, "  - %s\n", path)
	}
	return true, vault.Write(configPath, []byte(b.String()), vault.Options{FileMode: 0o600, DirMode: 0o700})
}

func expandInitPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	path = os.ExpandEnv(path)
	return filepath.Abs(path)
}

func relativeIfDescendant(target, base string) (string, error) {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("not descendant")
	}
	return rel, nil
}
