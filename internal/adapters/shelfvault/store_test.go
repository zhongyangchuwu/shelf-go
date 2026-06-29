package shelfvault

import "testing"

func TestListByTagsUsesAndSemantics(t *testing.T) {
	st := &Store{Data: NewData()}
	for path, secret := range map[string]Secret{
		"app:token":  {Value: mustParseValue(t, "one"), Env: "APP_TOKEN", Tags: []string{"ai", "prod"}},
		"app:url":    {Value: mustParseValue(t, "two"), Env: "APP_URL", Tags: []string{"ai"}},
		"other:key":  {Value: mustParseValue(t, "three"), Env: "OTHER_KEY", Tags: []string{"prod"}},
		"app:hidden": {Value: mustParseValue(t, "four"), Tags: []string{"ai", "prod"}},
	} {
		if err := st.Set(path, secret, false); err != nil {
			t.Fatalf("set %s: %v", path, err)
		}
	}

	got := st.ListByTags("app", []string{"ai", "prod"})
	want := []string{"app:hidden", "app:token"}
	if len(got) != len(want) {
		t.Fatalf("ListByTags length = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ListByTags[%d] = %q, want %q; all=%v", i, got[i], want[i], got)
		}
	}
}

func TestHasTagsMatchesEmptySelector(t *testing.T) {
	secret := Secret{Tags: []string{"ai"}}
	if !HasTags(secret, nil) {
		t.Fatalf("empty selector should match")
	}
}

func mustParseValue(t *testing.T, value string) []byte {
	t.Helper()
	raw, err := ParseValue(value)
	if err != nil {
		t.Fatalf("parse value: %v", err)
	}
	return raw
}
