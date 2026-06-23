package cli

import (
	"fmt"
	"io"
)

type diagnosticReport struct {
	out    io.Writer
	failed bool
}

func newDiagnosticReport(out io.Writer) *diagnosticReport {
	return &diagnosticReport{out: out}
}

func (r *diagnosticReport) ok(check, detail string) {
	r.line("ok  ", check, detail)
}

func (r *diagnosticReport) warn(check, detail string) {
	r.line("warn", check, detail)
}

func (r *diagnosticReport) fail(check, detail string) {
	r.failed = true
	r.line("fail", check, detail)
}

func (r *diagnosticReport) err(scope string) error {
	if r.failed {
		return fmt.Errorf("%s found failures", scope)
	}
	return nil
}

func (r *diagnosticReport) line(level, check, detail string) {
	fmt.Fprintf(r.out, "%s %s", level, check)
	if detail != "" {
		fmt.Fprintf(r.out, " (%s)", detail)
	}
	fmt.Fprintln(r.out)
}
