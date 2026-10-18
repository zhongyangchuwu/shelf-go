package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate shell completion script",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		ValidArgsFunction:     completeFixedArgs([]cobra.Completion{"bash", "zsh", "fish", "powershell"}),
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(cmd.OutOrStdout(), true)
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
			return nil
		},
	}
	return cmd
}

func completeFixedArgs(values []cobra.Completion) func(*cobra.Command, []string, string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		comps := make([]cobra.Completion, 0, len(values))
		for _, value := range values {
			if strings.HasPrefix(string(value), toComplete) {
				comps = append(comps, value)
			}
		}
		return comps, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeStringValues(values []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	completions := make([]cobra.Completion, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" || !strings.HasPrefix(value, toComplete) {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		completions = append(completions, cobra.Completion(value))
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func completeSecretMetadata(appSvc *app.App, cmd *cobra.Command, toComplete string, req app.SecretCompletionRequest) ([]cobra.Completion, cobra.ShellCompDirective) {
	configPath, vaultPath := runtimePaths(cmd)
	values, err := appSvc.SecretCompletions(configPath, vaultPath, req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeStringValues(values, toComplete)
}

func completeSecretSelectors(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	if envName, _ := cmd.Flags().GetString("env"); envName != "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretSelectors})
}

func completeSecretListQueries(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretQueries})
}

func completeSecretLabels(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretLabels})
}

func completeSecretEnvNames(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretEnvNames})
}

type secretFieldCompletionOptions struct {
	useEnvSelector  bool
	projectableOnly bool
	splitSelector   bool
}

func completeSecretFields(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string, options secretFieldCompletionOptions) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	req := app.SecretCompletionRequest{Target: app.CompleteSecretFields, ProjectableOnly: options.projectableOnly}
	if len(args) > 0 {
		req.Selector = args[0]
		if options.splitSelector {
			req.Selector, _, _ = strings.Cut(req.Selector, ".")
		}
	}
	if options.useEnvSelector {
		req.Env, _ = cmd.Flags().GetString("env")
	}
	return completeSecretMetadata(appSvc, cmd, toComplete, req)
}

func completeSecretTemplates(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	profile, _ := cmd.Flags().GetString("profile")
	return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretTemplates, Profile: profile})
}

func completeSecretProfiles(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	template, _ := cmd.Flags().GetString("template")
	return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretProfiles, Template: template})
}
