package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	vaultsvc "github.com/zhongyangchuwu/shelf-go/internal/vault"
	"github.com/zhongyangchuwu/shelf-go/internal/version"
)

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check local Shelf configuration and data health",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := newDiagnosticReport(cmd.OutOrStdout())
			configPath, _ := cmd.Flags().GetString("config")
			vaultPath, _ := cmd.Flags().GetString("vault")
			runtime, err := config.Resolve(configPath, vaultPath)
			if err != nil {
				report.fail("config resolves", err.Error())
				return report.err("doctor")
			}
			report.ok("config resolves", runtime.ConfigPath)
			report.ok("version", version.String())

			report.write(vaultsvc.Doctor(runtime))
			checkCompletion(report)

			return report.err("doctor")
		},
	}
	return cmd
}

func checkCompletion(report *diagnosticReport) {
	paths := completionSearchPaths()
	if len(paths) == 0 {
		report.warn("completion installed", "FPATH/fpath is not set")
		return
	}
	for _, dir := range paths {
		path := filepath.Join(dir, "_shelf")
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			report.ok("completion installed", path)
			return
		}
	}
	report.warn("completion installed", "_shelf not found in fpath")
}

func completionSearchPaths() []string {
	raw := os.Getenv("FPATH")
	if raw == "" {
		raw = os.Getenv("fpath")
	}
	if raw == "" {
		return nil
	}
	home, _ := os.UserHomeDir()
	parts := strings.Split(raw, ":")
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "~/") && home != "" {
			part = filepath.Join(home, strings.TrimPrefix(part, "~/"))
		}
		paths = append(paths, part)
	}
	return paths
}
