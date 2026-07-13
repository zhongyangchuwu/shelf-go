package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/cliui"
)

func newSecretAuditCmd(appSvc *app.App) *cobra.Command {
	var withinDays int
	var format string
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Report credential expiry health without values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if format != "text" && format != "json" {
				return fmt.Errorf("unsupported audit format: %s", format)
			}
			configPath, vaultPath := runtimePaths(cmd)
			report, err := appSvc.CredentialAudit(configPath, vaultPath, app.CredentialAuditRequest{WithinDays: withinDays})
			if err != nil {
				return err
			}
			if err := renderCredentialAudit(cmd.OutOrStdout(), report, format); err != nil {
				return err
			}
			return credentialAuditResultError(report)
		},
	}
	cmd.Flags().IntVar(&withinDays, "within", app.DefaultCredentialAuditWithinDays, "Warn when a credential expires within this many days")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	_ = cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions([]string{"text", "json"}, cobra.ShellCompDirectiveNoFileComp))
	return cmd
}

func renderCredentialAudit(w io.Writer, report app.CredentialAuditReport, format string) error {
	if format == "json" {
		return json.NewEncoder(w).Encode(report)
	}
	rows := make([][]string, 0, len(report.Items))
	if len(report.Items) == 0 {
		cliui.Empty(w, "No credentials with expiry metadata found.")
		return nil
	}
	for _, item := range report.Items {
		expires := item.Expires
		if expires == "" {
			expires = "-"
		}
		rows = append(rows, []string{item.ID, item.Field, expires, auditHealthLabel(item.Status)})
	}
	cliui.RenderTable(w, cliui.Table{Title: fmt.Sprintf("Credential health (within %d days)", report.WithinDays), Headers: []string{"ID", "FIELD", "EXPIRES", "STATUS"}, Rows: rows})
	return nil
}

type credentialAuditExitError struct {
	code int
}

func (e credentialAuditExitError) Error() string {
	switch e.code {
	case 2:
		return "credential health warnings found"
	case 3:
		return "expired credentials found"
	default:
		return "credential audit failed"
	}
}

func (e credentialAuditExitError) ExitCode() int {
	return e.code
}

func credentialAuditResultError(report app.CredentialAuditReport) error {
	if report.HasExpired() {
		return credentialAuditExitError{code: 3}
	}
	if report.HasWarnings() {
		return credentialAuditExitError{code: 2}
	}
	return nil
}

func auditHealthLabel(status app.CredentialHealthStatus) string {
	switch status {
	case app.CredentialHealthExpiring:
		return "due"
	case app.CredentialHealthInvalid:
		return "invalid"
	default:
		return string(status)
	}
}
