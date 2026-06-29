package gopass

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/source"
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

func TestReaderGetShowsPasswordOnly(t *testing.T) {
	runner := &fakeRunner{outputs: map[string][]byte{"show --password app/token": []byte("secret\n")}}
	reader := Reader{Binary: "gopass-test", Runner: runner}
	secret, err := reader.Get("app:token")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if secret.Path != "app:token" || secret.Value != "secret" {
		t.Fatalf("unexpected secret: %#v", secret)
	}
	wantCalls := []string{"gopass-test show --password app/token"}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestReaderGetMapsMissingSecret(t *testing.T) {
	runner := &fakeRunner{errs: map[string]error{"show --password missing/key": errors.New("secret not found")}}
	reader := Reader{Runner: runner}
	_, err := reader.Get("missing:key")
	if !errors.Is(err, source.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestReaderListMapsAndFiltersPrefix(t *testing.T) {
	runner := &fakeRunner{outputs: map[string][]byte{"list --flat": []byte("db/prod/password\napp/token\napp/other\ninvalid\n")}}
	reader := Reader{Runner: runner}
	paths, err := reader.List("app")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := []string{"app:other", "app:token"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %#v, want %#v", paths, want)
	}
}

func TestReaderTagsUnsupported(t *testing.T) {
	reader := Reader{}
	_, err := reader.ListByTags("", []string{"prod"})
	if !errors.Is(err, ErrTagsUnsupported) {
		t.Fatalf("err = %v, want ErrTagsUnsupported", err)
	}
}
