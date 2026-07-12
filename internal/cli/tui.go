package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/clipboard"
	"github.com/zhongyangchuwu/shelf-go/internal/tui"
	"golang.org/x/term"
)

func newTUICmd(appSvc *app.App, clipboardWriter clipboard.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open the interactive secret workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireInteractiveTerminal(cmd); err != nil {
				return err
			}
			configPath, vaultPath := runtimePaths(cmd)
			return tui.Run(appSvc, clipboardWriter, tui.Options{ConfigPath: configPath, VaultPath: vaultPath})
		},
	}
}

func requireInteractiveTerminal(cmd *cobra.Command) error {
	in, inOK := cmd.InOrStdin().(*os.File)
	out, outOK := cmd.OutOrStdout().(*os.File)
	if !inOK || !outOK || !term.IsTerminal(int(in.Fd())) || !term.IsTerminal(int(out.Fd())) {
		return fmt.Errorf("tui requires an interactive terminal")
	}
	return nil
}
