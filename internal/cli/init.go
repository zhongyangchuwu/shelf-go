package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/han/shelf-go/internal/config"
	"github.com/han/shelf-go/internal/store"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force, minimal bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize config and data files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPathFlag, _ := cmd.Flags().GetString("config")
			dataPathFlag, _ := cmd.Flags().GetString("data")
			runtime, err := config.Resolve(configPathFlag, dataPathFlag)
			if err != nil {
				return err
			}

			dataCreated, err := ensureDataFile(runtime.DataPath, force)
			if err != nil {
				return err
			}

			var configCreated bool
			if !minimal {
				configCreated, err = ensureConfigFile(runtime.ConfigPath, runtime.DataPath, force)
				if err != nil {
					return err
				}
			}

			label := map[bool]string{true: "created", false: "exists"}
			fmt.Fprintf(cmd.OutOrStdout(), "data:   %s (%s)\n", runtime.DataPath, label[dataCreated])
			if !minimal {
				fmt.Fprintf(cmd.OutOrStdout(), "config: %s (%s)\n", runtime.ConfigPath, label[configCreated])
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&minimal, "minimal", false, "Only create the data file")
	return cmd
}

func ensureDataFile(dataPath string, force bool) (bool, error) {
	if _, err := os.Stat(dataPath); err == nil && !force {
		return false, nil
	}
	st := &store.Store{Path: dataPath, Data: store.NewData()}
	if err := st.Save(); err != nil {
		return false, err
	}
	return true, nil
}

func ensureConfigFile(configPath, dataPath string, force bool) (bool, error) {
	if _, err := os.Stat(configPath); err == nil && !force {
		return false, nil
	}
	rel, err := relativeIfDescendant(dataPath, filepath.Dir(configPath))
	if err != nil {
		rel = dataPath
	}
	content := fmt.Sprintf("version: 1\ndata: %s\n", rel)
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return false, err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(configPath)+".tmp-*")
	if err != nil {
		return false, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return false, err
	}
	if err := tmp.Close(); err != nil {
		return false, err
	}
	return true, os.Rename(tmpName, configPath)
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
	if err != nil {
		return "", err
	}
	return rel, nil
}
