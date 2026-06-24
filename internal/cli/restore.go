package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

func newRestoreCmd() *cobra.Command {
	var sourcePath string
	var targetPath string
	var force bool
	cmd := &cobra.Command{
		Use:   "restore --from <backup.age>",
		Short: "Restore an encrypted vault backup",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sourcePath == "" {
				return fmt.Errorf("--from is required")
			}
			runtime, configuredVault, err := loadVault(cmd)
			if err != nil {
				return err
			}

			sourceVault, err := store.NewVault(sourcePath, store.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
			if err != nil {
				return err
			}

			targetVault := configuredVault
			if targetPath != "" && targetPath != runtime.VaultPath {
				targetVault, err = store.NewVault(targetPath, store.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
				if err != nil {
					return err
				}
			}

			if err := restoreEncryptedVault(sourceVault, targetVault, force); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "restored encrypted vault %s to %s\n", sourceVault.Path(), targetVault.Path())
			fmt.Fprintf(cmd.OutOrStdout(), "run shelf vault status to verify the restored vault before using it\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&sourcePath, "from", "", "Path to encrypted vault backup")
	cmd.Flags().StringVar(&targetPath, "to", "", "Path to encrypted vault target")
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing encrypted vault target")
	return cmd
}

func restoreEncryptedVault(sourceVault, targetVault *store.Vault, force bool) error {
	sourceFormat, err := store.DetectFileFormat(sourceVault.Path())
	if err != nil {
		return fmt.Errorf("inspect restore source: %w", err)
	}
	if sourceFormat != store.FileFormatEncryptedVault {
		if sourceFormat == store.FileFormatPlaintextStore {
			return fmt.Errorf("restore source is plaintext JSON; use shelf vault migrate for plaintext stores")
		}
		return fmt.Errorf("restore source must be an encrypted vault: %s", sourceVault.Path())
	}

	st, err := sourceVault.Load()
	if err != nil {
		return fmt.Errorf("load restore source: %w", err)
	}

	lock, err := targetVault.Lock()
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()

	targetFormat, err := store.DetectFileFormat(targetVault.Path())
	if err != nil {
		return fmt.Errorf("inspect target vault: %w", err)
	}
	if targetFormat == store.FileFormatPlaintextStore {
		return fmt.Errorf("target vault is plaintext JSON; choose a different --to path or move it before restore")
	}
	if targetFormat != store.FileFormatMissing && targetFormat != store.FileFormatEmpty && !force {
		return fmt.Errorf("target vault already exists; pass --force to replace %s", targetVault.Path())
	}

	if err := targetVault.Save(st); err != nil {
		return fmt.Errorf("write restored vault: %w", err)
	}
	verified, err := targetVault.Load()
	if err != nil {
		return fmt.Errorf("verify restored vault: %w", err)
	}
	if len(verified.Data.Secrets) != len(st.Data.Secrets) {
		return fmt.Errorf("verify restored vault: secret count mismatch")
	}
	return nil
}
