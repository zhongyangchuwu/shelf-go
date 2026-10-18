package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestSecretSetCreatesAndUpdatesFromStdin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	for _, value := range []string{"first-private-value", "second-private-value"} {
		out, err := runShelfWithInput(t, value+"\n", "--vault", path, "secret", "set", "QUICK_TOKEN")
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}
		if strings.Contains(out, value) {
			t.Fatal("set output leaked input")
		}
		got, err := runShelf(t, "--vault", path, "secret", "get", "--env", "QUICK_TOKEN")
		if err != nil || got != value+"\n" {
			t.Fatalf("get after set = %q, %v", got, err)
		}
	}
	vault, err := vaultjson.NewRepository(path).Load()
	if err != nil || len(vault.Items) != 1 || vault.Items[0].Title != "QUICK_TOKEN" {
		t.Fatalf("set update did not retain a single item: %v", err)
	}
	for _, args := range [][]string{
		{"secret", "set", "QUICK_TOKEN", "argv-value"},
		{"secret", "set", "QUICK_TOKEN", "--value", "argv-value"},
	} {
		if _, err := runShelfWithInput(t, "unused", append([]string{"--vault", path}, args...)...); err == nil {
			t.Fatalf("set accepted unsupported argv contract %v", args)
		}
	}
}

func TestSecretSetAmbiguityRequiresConcreteTargetWithoutReadingInput(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{
		"alpha": {env: "SHARED_TOKEN", value: "alpha-private-value", labels: []string{"personal"}},
		"beta":  {env: "SHARED_TOKEN", value: "beta-private-value", labels: []string{"work"}},
	})
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	input := &countingSetInput{}
	cmd := NewRootCmd(testApp())
	var out bytes.Buffer
	cmd.SetIn(input)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--vault", path, "secret", "set", "SHARED_TOKEN"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("ambiguous set succeeded")
	}
	for _, want := range []string{"item_alpha.value", "item_beta.value", "--target ITEM.FIELD"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("ambiguity error missing %q: %v", want, err)
		}
	}
	if input.reads != 0 {
		t.Fatal("ambiguous noninteractive set read secret input")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("ambiguous set mutated storage: %v", err)
	}
	for _, target := range []string{"alpha.value", "item_alpha.env_value", "item_missing.value"} {
		if _, err := runShelfWithInput(t, "new-private-value", "--vault", path, "secret", "set", "SHARED_TOKEN", "--target", target); err == nil {
			t.Fatalf("accepted invalid concrete target %s", target)
		}
	}
	result, err := runShelfWithInput(t, "new-private-value", "--vault", path, "secret", "set", "SHARED_TOKEN", "--target", "item_beta.value")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"alpha-private-value", "beta-private-value", "new-private-value"} {
		if strings.Contains(result+out.String(), value) {
			t.Fatal("set output leaked a value")
		}
	}
	for selector, want := range map[string]string{"item_alpha": "alpha-private-value\n", "item_beta": "new-private-value\n"} {
		got, err := runShelf(t, "--vault", path, "secret", "get", selector)
		if err != nil || got != want {
			t.Fatalf("field after explicit selection = %q, %v", got, err)
		}
	}
	for _, action := range []string{"get", "copy"} {
		if _, err := runShelf(t, "--vault", path, "secret", action, "--env", "SHARED_TOKEN"); err == nil {
			t.Fatalf("%s ambiguity behavior changed", action)
		}
	}
}

func TestSecretSetCompletionsIncludeAmbiguityAndRemainReadOnly(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{
		"alpha": {env: "SHARED_TOKEN", value: "alpha-private-value"},
		"beta":  {env: "SHARED_TOKEN", value: "beta-private-value"},
	})
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{"--vault", path}
	if got := dynamicCompletions(t, flags, "secret", "set", "SHARED"); !reflect.DeepEqual(got, []string{"SHARED_TOKEN"}) {
		t.Fatalf("set env completions = %v", got)
	}
	targets := dynamicCompletions(t, flags, "secret", "set", "SHARED_TOKEN", "--target", "item_")
	if len(targets) != 2 || !hasCompletion(targets, "item_alpha.value") || !hasCompletion(targets, "item_beta.value") {
		t.Fatalf("set target completions = %v", targets)
	}
	if got := dynamicCompletions(t, flags, "secret", "set", "UNMAPPED", "--target", ""); len(got) != 0 {
		t.Fatalf("unrelated target completions = %v", got)
	}
	for _, action := range []string{"get", "copy"} {
		if got := dynamicCompletions(t, flags, "secret", action, "--env", "SHARED"); len(got) != 0 {
			t.Fatalf("%s ambiguity completion changed: %v", action, got)
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("completion mutated vault: %v", err)
	}
	missing := filepath.Join(t.TempDir(), "missing", "vault.json")
	if got := dynamicCompletions(t, []string{"--vault", missing}, "secret", "set", ""); len(got) != 0 {
		t.Fatalf("missing vault completions = %v", got)
	}
	if _, err := os.Stat(filepath.Dir(missing)); !os.IsNotExist(err) {
		t.Fatalf("completion initialized storage: %v", err)
	}
}

func TestSecretSetUsesConcreteIDDespiteTitleCollision(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{
		"alpha": {env: "FIRST_TOKEN", value: "alpha-private-value"},
		"beta":  {env: "SECOND_TOKEN", value: "beta-private-value"},
	})
	repo := vaultjson.NewRepository(path)
	vault, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	for index := range vault.Items {
		if vault.Items[index].ID == "item_beta" {
			vault.Items[index].Title = "item_alpha"
		}
	}
	if err := repo.Save(vault); err != nil {
		t.Fatal(err)
	}
	if _, err := runShelfWithInput(t, "new-private-value", "--vault", path, "secret", "set", "FIRST_TOKEN", "--target", "item_alpha.value"); err != nil {
		t.Fatal(err)
	}
	after, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(after)
	for _, item := range after.Items {
		_, field, err := resolver.Field(item, "value")
		if err != nil {
			t.Fatal(err)
		}
		want := "beta-private-value"
		if item.ID == "item_alpha" {
			want = "new-private-value"
		}
		if string(field.Value) != string(rawJSON(want)) {
			t.Fatalf("concrete target wrote the wrong item")
		}
	}
}

type countingSetInput struct{ reads int }

func (r *countingSetInput) Read([]byte) (int, error) {
	r.reads++
	return 0, os.ErrClosed
}
