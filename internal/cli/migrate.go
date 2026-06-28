package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
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
			if err := app.MigratePlaintextStore(sourcePath, targetVault, force); err != nil {
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
