package cli

import (
	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
	"github.com/zhongyangchuwu/shelf-go/internal/version"
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
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newProjectCmd())
	root.AddCommand(newRunCmd())
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

func loadRuntimeForWrite(cmd *cobra.Command) (config.Runtime, *store.Store, func(), error) {
	configPath, _ := cmd.Flags().GetString("config")
	dataPath, _ := cmd.Flags().GetString("data")
	runtime, err := config.Resolve(configPath, dataPath)
	if err != nil {
		return config.Runtime{}, nil, nil, err
	}
	lock, err := store.LockFile(runtime.DataPath)
	if err != nil {
		return config.Runtime{}, nil, nil, err
	}
	unlock := func() { _ = lock.Unlock() }
	st, err := store.Load(runtime.DataPath)
	if err != nil {
		unlock()
		return config.Runtime{}, nil, nil, err
	}
	return runtime, st, unlock, nil
}
