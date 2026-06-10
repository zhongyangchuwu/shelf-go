package cli

import (
	"github.com/han/shelf-go/internal/config"
	"github.com/han/shelf-go/internal/store"
	"github.com/han/shelf-go/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "shelf",
		Short:         "Shelf Go rewrite",
		Version:       version.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("config", "", "Path to config.yaml")
	root.PersistentFlags().String("data", "", "Path to secrets.json")

	root.AddCommand(newCompletionCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newSecretCmd())
	root.AddCommand(newExportCmd())
	root.AddCommand(newProjectCmd())
	return root
}

func loadRuntime(cmd *cobra.Command) (config.Runtime, *store.Store, error) {
	configPath, _ := cmd.Flags().GetString("config")
	dataPath, _ := cmd.Flags().GetString("data")
	runtime, err := config.Resolve(configPath, dataPath)
	if err != nil {
		return config.Runtime{}, nil, err
	}
	st, err := store.Load(runtime.DataPath)
	if err != nil {
		return config.Runtime{}, nil, err
	}
	return runtime, st, nil
}
