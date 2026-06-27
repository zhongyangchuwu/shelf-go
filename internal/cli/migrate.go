package cli

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func newMigrateCmd() *cobra.Command {
	var sourcePath string
	var targetPath string
	var force bool
	cmd := &cobra.Command{
		Use:   "migrate --from <plaintext.json>",
		Short: "Migrate a plaintext store into an encrypted vault",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sourcePath == "" {
				return fmt.Errorf("--from is required")
			}
			runtime, configuredVault, err := loadVault(cmd)
			if err != nil {
				return err
			}
			targetVault := configuredVault
			if targetPath != "" && targetPath != runtime.VaultPath {
				targetVault, err = vault.NewVault(targetPath, vault.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
				if err != nil {
					return err
				}
			}
			if err := migratePlaintextStore(sourcePath, targetVault, force); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "migrated plaintext store %s to encrypted vault %s\n", sourcePath, targetVault.Path())
			fmt.Fprintf(cmd.OutOrStdout(), "plaintext source preserved at %s; move, delete, or archive it after confirming your new config\n", sourcePath)
			return nil
		},
	}
	cmd.Flags().StringVar(&sourcePath, "from", "", "Path to plaintext Shelf JSON store")
	cmd.Flags().StringVar(&targetPath, "to", "", "Path to encrypted vault target")
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing encrypted vault target")
	return cmd
}

func migratePlaintextStore(sourcePath string, targetVault *vault.Vault, force bool) error {
	before, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read plaintext source: %w", err)
	}
	st, err := vault.Load(sourcePath)
	if err != nil {
		return fmt.Errorf("load plaintext source: %w", err)
	}
	format, err := vault.DetectFileFormat(targetVault.Path())
	if err != nil {
		return fmt.Errorf("inspect target vault: %w", err)
	}
	if format != vault.FileFormatMissing && format != vault.FileFormatEmpty && !force {
		return fmt.Errorf("target vault already exists; pass --force to replace %s", targetVault.Path())
	}
	if format == vault.FileFormatPlaintextStore {
		return fmt.Errorf("target vault is plaintext JSON; choose a different --to path or move it before migration")
	}
	if err := targetVault.Save(st); err != nil {
		return fmt.Errorf("write encrypted vault: %w", err)
	}
	verified, err := targetVault.Load()
	if err != nil {
		return fmt.Errorf("verify encrypted vault: %w", err)
	}
	if len(verified.Data.Secrets) != len(st.Data.Secrets) {
		return fmt.Errorf("verify encrypted vault: secret count mismatch")
	}
	after, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("re-read plaintext source: %w", err)
	}
	if !bytes.Equal(before, after) {
		return fmt.Errorf("plaintext source changed during migration")
	}
	return nil
}
