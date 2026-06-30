package gopass

import (
	"reflect"
	"strings"
	"testing"
)

type fakeRunner struct {
	outputs map[string][]byte
	errs    map[string]error
	calls   []string
}

func (r *fakeRunner) Run(name string, args ...string) ([]byte, error) {
	call := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, call)
	if err := r.errs[strings.Join(args, " ")]; err != nil {
		return nil, err
	}
	return r.outputs[strings.Join(args, " ")], nil
}

func TestClientShowPasswordUsesPasswordOnly(t *testing.T) {
	runner := &fakeRunner{outputs: map[string][]byte{"show --password app/token": []byte("secret\n")}}
	client := Client{Binary: "gopass-test", Runner: runner}
	secret, err := client.ShowPassword("app/token")
	if err != nil {
		t.Fatalf("show password: %v", err)
	}
	if secret != "secret" {
		t.Fatalf("secret = %q, want secret", secret)
	}
	wantCalls := []string{"gopass-test show --password app/token"}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestClientListFlatFiltersPrefix(t *testing.T) {
	runner := &fakeRunner{outputs: map[string][]byte{"list --flat": []byte("db/prod/password\napp/token\napp/other\n")}}
	client := Client{Runner: runner}
	paths, err := client.ListFlat("app")
	if err != nil {
		t.Fatalf("list flat: %v", err)
	}
	want := []string{"app/other", "app/token"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %#v, want %#v", paths, want)
	}
}
