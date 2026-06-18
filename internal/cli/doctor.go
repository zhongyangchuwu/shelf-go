package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
	"github.com/zhongyangchuwu/shelf-go/internal/version"
)

type doctorReport struct {
	out    *cobra.Command
	failed bool
}

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check local Shelf configuration and data health",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := doctorReport{out: cmd}
			configPath, _ := cmd.Flags().GetString("config")
			vaultPath, _ := cmd.Flags().GetString("vault")
			runtime, err := config.Resolve(configPath, vaultPath)
			if err != nil {
				report.fail("config resolves", err.Error())
				return fmt.Errorf("doctor found failures")
			}
			report.ok("config resolves", runtime.ConfigPath)
			report.ok("version", version.String())

			checkVaultFile(&report, runtime.VaultPath)
			checkVaultLoads(&report, runtime)
			checkVaultTracking(&report, runtime.VaultPath)
			checkCompletion(&report)

			if report.failed {
				return fmt.Errorf("doctor found failures")
			}
			return nil
		},
	}
	return cmd
}

func (r *doctorReport) ok(check, detail string) {
	fmt.Fprintf(r.out.OutOrStdout(), "ok   %s", check)
	if detail != "" {
		fmt.Fprintf(r.out.OutOrStdout(), " (%s)", detail)
	}
	fmt.Fprintln(r.out.OutOrStdout())
}

func (r *doctorReport) warn(check, detail string) {
	fmt.Fprintf(r.out.OutOrStdout(), "warn %s", check)
	if detail != "" {
		fmt.Fprintf(r.out.OutOrStdout(), " (%s)", detail)
	}
	fmt.Fprintln(r.out.OutOrStdout())
}

func (r *doctorReport) fail(check, detail string) {
	r.failed = true
	fmt.Fprintf(r.out.OutOrStdout(), "fail %s", check)
	if detail != "" {
		fmt.Fprintf(r.out.OutOrStdout(), " (%s)", detail)
	}
	fmt.Fprintln(r.out.OutOrStdout())
}

func checkVaultFile(report *doctorReport, vaultPath string) {
	info, err := os.Stat(vaultPath)
	if os.IsNotExist(err) {
		report.warn("vault file exists", vaultPath+" will be created on first write")
		return
	}
	if err != nil {
		report.fail("vault file exists", err.Error())
		return
	}
	if info.IsDir() {
		report.fail("vault file is regular file", vaultPath+" is a directory")
		return
	}
	report.ok("vault file exists", vaultPath)
	mode := info.Mode().Perm()
	if mode&0o077 == 0 {
		report.ok("vault file mode", mode.String())
	} else {
		report.warn("vault file mode", mode.String()+" is broader than 0600")
	}
}
func checkVaultLoads(report *doctorReport, runtime config.Runtime) {
	vault, err := store.NewVault(runtime.VaultPath, store.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
	if err != nil {
		report.fail("vault loads", err.Error())
		return
	}
	if _, err := vault.Load(); err != nil {
		report.fail("vault loads", err.Error())
		return
	}
	report.ok("vault loads", runtime.VaultPath)
}

func checkVaultTracking(report *doctorReport, vaultPath string) {
	abs, err := filepath.Abs(vaultPath)
	if err != nil {
		report.warn("git tracking", err.Error())
		return
	}
	rootBytes, err := exec.Command("git", "-C", filepath.Dir(abs), "rev-parse", "--show-toplevel").Output()
	if err != nil {
		report.ok("git tracking", "vault file is not inside a Git worktree")
		return
	}
	root := strings.TrimSpace(string(rootBytes))
	rel, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		report.ok("git tracking", "vault file is outside Git worktree")
		return
	}
	if err := exec.Command("git", "-C", root, "ls-files", "--error-unmatch", "--", rel).Run(); err == nil {
		report.warn("git tracking", "vault file appears tracked by ordinary git: "+rel)
		return
	}
	report.ok("git tracking", "vault file is not tracked by ordinary git")
}

func checkCompletion(report *doctorReport) {
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
