package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/project"
)

func newProjectConfigureCmd(appSvc *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Interactively configure project secret bindings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			options, err := appSvc.ProjectConfigureOptions(configPath, vaultPath)
			if err != nil {
				return err
			}
			return runProjectConfigure(cmd.InOrStdin(), cmd.OutOrStdout(), appSvc, configPath, vaultPath, options)
		},
	}
}

func runProjectConfigure(in io.Reader, out io.Writer, appSvc *app.App, configPath, vaultPath string, options app.ProjectConfigureOptions) error {
	if len(options.EnvFiles) == 0 {
		fmt.Fprintln(out, "no env files found")
		return nil
	}
	if len(options.SecretPaths) == 0 {
		fmt.Fprintln(out, "no vault secrets found")
		return nil
	}

	reader := bufio.NewReader(in)
	envFile, err := promptEnvFile(reader, out, options.EnvFiles)
	if err != nil {
		return err
	}
	envVar, err := promptEnvVar(reader, out, envFile)
	if err != nil {
		return err
	}
	secretPath, err := promptChoice(reader, out, "Secret path", options.SecretPaths)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "plan:")
	fmt.Fprintf(out, "  env:  %s\n", envVar.Name)
	fmt.Fprintf(out, "  path: %s\n", secretPath)
	confirmed, err := promptConfirm(reader, out, "Write this project binding? [y/N]")
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Fprintln(out, "cancelled")
		return nil
	}
	result, err := appSvc.ProjectConfigureBind(configPath, vaultPath, envVar.Name, secretPath)
	if err != nil {
		return err
	}
	fmt.Fprint(out, result)
	return nil
}

func promptEnvFile(reader *bufio.Reader, out io.Writer, files []project.EnvFileDetail) (project.EnvFileDetail, error) {
	choices := make([]string, 0, len(files))
	byName := make(map[string]project.EnvFileDetail, len(files))
	for _, file := range files {
		if file.Error != "" || !hasUnboundEnvVar(file) {
			continue
		}
		choices = append(choices, file.Name)
		byName[file.Name] = file
	}
	name, err := promptChoice(reader, out, "Env file", choices)
	if err != nil {
		return project.EnvFileDetail{}, err
	}
	return byName[name], nil
}

func promptEnvVar(reader *bufio.Reader, out io.Writer, file project.EnvFileDetail) (project.EnvVarStatus, error) {
	choices := make([]string, 0, len(file.Vars))
	byName := make(map[string]project.EnvVarStatus, len(file.Vars))
	for _, variable := range file.Vars {
		if variable.Bound {
			continue
		}
		choices = append(choices, variable.Name)
		byName[variable.Name] = variable
	}
	name, err := promptChoice(reader, out, "Env variable", choices)
	if err != nil {
		return project.EnvVarStatus{}, err
	}
	return byName[name], nil
}

func hasUnboundEnvVar(file project.EnvFileDetail) bool {
	for _, variable := range file.Vars {
		if !variable.Bound {
			return true
		}
	}
	return false
}

func promptChoice(reader *bufio.Reader, out io.Writer, label string, choices []string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("no choices available for %s", strings.ToLower(label))
	}
	fmt.Fprintf(out, "%s:\n", label)
	for _, choice := range choices {
		fmt.Fprintf(out, "  %s\n", choice)
	}
	fmt.Fprintf(out, "> ")
	answer, err := reader.ReadString('\n')
	if err != nil && answer == "" {
		return "", err
	}
	answer = strings.TrimSpace(answer)
	for _, choice := range choices {
		if answer == choice {
			return choice, nil
		}
	}
	return "", fmt.Errorf("invalid %s: %s", strings.ToLower(label), answer)
}

func promptConfirm(reader *bufio.Reader, out io.Writer, label string) (bool, error) {
	fmt.Fprintf(out, "%s ", label)
	answer, err := reader.ReadString('\n')
	if err != nil && answer == "" {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}
