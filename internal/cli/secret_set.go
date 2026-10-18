package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"golang.org/x/term"
)

func newSecretSetCmd(appSvc *app.App) *cobra.Command {
	var target string
	cmd := &cobra.Command{
		Use:   "set <ENV>",
		Short: "Set a mapped environment value, or create a single secret",
		Long:  "Set an environment value from piped stdin or an automatic hidden terminal prompt. Existing mappings update only the selected field. Unmapped names create a single secret; ID or title collisions fail without changing existing items.",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretSetEnvNames})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			targets, err := appSvc.SecretSetTargets(configPath, vaultPath, args[0])
			if err != nil {
				return err
			}
			selected := target
			if cmd.Flags().Changed("target") {
				valid := false
				for _, candidate := range targets {
					if candidate == selected {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("--target must be one of %s's existing item-ID.field mappings", args[0])
				}
			} else {
				switch len(targets) {
				case 0:
				case 1:
					selected = targets[0]
				default:
					selected, err = pickSecretSetTarget(cmd, args[0], targets)
					if err != nil {
						return err
					}
				}
			}
			input, err := updateFieldInput(cmd, args[0])
			if err != nil {
				return fmt.Errorf("read secret input failed")
			}
			if _, err := appSvc.SetSecret(configPath, vaultPath, app.SetSecretRequest{Env: args[0], Target: selected, In: input}); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Set secret value.")
			return nil
		},
	}
	cmd.Flags().StringVar(&target, "target", "", "Concrete item-ID.field mapped to ENV (required for noninteractive ambiguity)")
	_ = cmd.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretSetTargets, Env: args[0]})
	})
	return cmd
}

func pickSecretSetTarget(cmd *cobra.Command, env string, targets []string) (string, error) {
	input, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(input.Fd())) {
		return "", fmt.Errorf("environment name %s maps to multiple fields:\n- %s\nchoose one with --target ITEM.FIELD", env, strings.Join(targets, "\n- "))
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Choose a target for %s:\n", env)
	for index, target := range targets {
		fmt.Fprintf(cmd.ErrOrStderr(), "  %d) %s\n", index+1, target)
	}
	fmt.Fprint(cmd.ErrOrStderr(), "Target number (blank cancels): ")
	// Read only the choice line: buffering may otherwise consume bytes meant
	// for the subsequent hidden prompt.
	var choice strings.Builder
	var one [1]byte
	for {
		_, err := io.ReadFull(input, one[:])
		if err != nil {
			return "", fmt.Errorf("target selection cancelled")
		}
		if one[0] == '\n' {
			break
		}
		choice.WriteByte(one[0])
	}
	number, err := strconv.Atoi(strings.TrimSpace(choice.String()))
	if err != nil || number < 1 || number > len(targets) {
		return "", fmt.Errorf("target selection cancelled: choose a listed number")
	}
	return targets[number-1], nil
}
