package project

import (
	"fmt"
	"io"

	"github.com/zhongyangchuwu/shelf-go/internal/util"
)

type Binding struct {
	Path    string
	EnvName string
	Value   string
}

type Diagnostic struct {
	Status  string
	Path    string
	Message string
}

func RenderDiagnostics(w io.Writer, diagnostics []Diagnostic) {
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(w, "%s %s %s\n", diagnostic.Status, diagnostic.Path, diagnostic.Message)
	}
}

func HasFailures(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Status == "fail" {
			return true
		}
	}
	return false
}

func BindingsForRender(entries []Binding) []util.Binding {
	bindings := make([]util.Binding, 0, len(entries))
	for _, entry := range entries {
		bindings = append(bindings, util.Binding{EnvName: entry.EnvName, Value: entry.Value})
	}
	return bindings
}
