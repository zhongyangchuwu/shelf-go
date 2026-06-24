package manager

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

const testToken = "test-token"

func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	identityPath := filepath.Join(dir, "identity.txt")
	if err := os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0o600); err != nil {
		t.Fatalf("write identity: %v", err)
	}
	vaultPath := filepath.Join(dir, "vault.age")
	vault, err := store.NewVault(vaultPath, store.VaultOptions{Recipients: []string{identity.Recipient().String()}, IdentityPaths: []string{identityPath}})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	server, err := NewServer(vault, testToken, "127.0.0.1:4321")
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	return server, vaultPath
}

func managerRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Host = "127.0.0.1:4321"
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func serveManager(server *Server, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	return rr
}

func TestManagerRequiresTokenAndValidHost(t *testing.T) {
	server, _ := newTestServer(t)

	missingToken := managerRequest(http.MethodGet, "/api/secrets", "")
	if rr := serveManager(server, missingToken); rr.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	badHost := managerRequest(http.MethodGet, "/api/secrets?token="+testToken, "")
	badHost.Host = "evil.test"
	if rr := serveManager(server, badHost); rr.Code != http.StatusForbidden {
		t.Fatalf("bad host status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestManagerAcceptsLocalhostAndAlternateTokenTransports(t *testing.T) {
	server, _ := newTestServer(t)

	localhost := managerRequest(http.MethodGet, "/api/secrets?token="+testToken, "")
	localhost.Host = "localhost:4321"
	if rr := serveManager(server, localhost); rr.Code != http.StatusOK {
		t.Fatalf("localhost status = %d, want %d", rr.Code, http.StatusOK)
	}

	headerToken := managerRequest(http.MethodGet, "/api/secrets", "")
	headerToken.Header.Set("X-Shelf-Token", testToken)
	if rr := serveManager(server, headerToken); rr.Code != http.StatusOK {
		t.Fatalf("header token status = %d, want %d", rr.Code, http.StatusOK)
	}

	cookieToken := managerRequest(http.MethodGet, "/api/secrets", "")
	cookieToken.AddCookie(&http.Cookie{Name: sessionCookie, Value: testToken})
	if rr := serveManager(server, cookieToken); rr.Code != http.StatusOK {
		t.Fatalf("cookie token status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestManagerQueryTokenSetsStrictCookie(t *testing.T) {
	server, _ := newTestServer(t)
	req := managerRequest(http.MethodGet, "/?token="+testToken, "")
	rr := serveManager(server, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("index status = %d, want %d", rr.Code, http.StatusOK)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != sessionCookie || cookie.Value != testToken {
		t.Fatalf("unexpected cookie: %#v", cookie)
	}
	if !cookie.HttpOnly {
		t.Fatalf("cookie is not HttpOnly")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie SameSite = %v, want Strict", cookie.SameSite)
	}
}

func TestManagerListSearchExcludesSecretValues(t *testing.T) {
	server, _ := newTestServer(t)
	payload := `{"path":"app:token","value":"secret-value","env":"APP_TOKEN","description":"primary api token","tags":["api"]}`
	writeReq := managerRequest(http.MethodPost, "/api/secrets?token="+testToken, payload)
	writeReq.Header.Set("Origin", "http://127.0.0.1:4321")
	if rr := serveManager(server, writeReq); rr.Code != http.StatusOK {
		t.Fatalf("write status = %d body=%s", rr.Code, rr.Body.String())
	}

	listReq := managerRequest(http.MethodGet, "/api/secrets?token="+testToken+"&q=api", "")
	rr := serveManager(server, listReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{"app:token", "APP_TOKEN", "primary api token", "\"value_set\":true"} {
		if !strings.Contains(body, want) {
			t.Fatalf("list response missing %q: %s", want, body)
		}
	}
	if strings.Contains(body, "secret-value") {
		t.Fatalf("list response leaked secret value: %s", body)
	}
}

func TestManagerRevealIsExplicit(t *testing.T) {
	server, _ := newTestServer(t)
	payload := `{"path":"app:token","value":"secret-value","env":"APP_TOKEN"}`
	writeReq := managerRequest(http.MethodPost, "/api/secrets?token="+testToken, payload)
	writeReq.Header.Set("Origin", "http://127.0.0.1:4321")
	if rr := serveManager(server, writeReq); rr.Code != http.StatusOK {
		t.Fatalf("write status = %d body=%s", rr.Code, rr.Body.String())
	}

	withoutToken := managerRequest(http.MethodGet, "/api/reveal?path=app:token", "")
	if rr := serveManager(server, withoutToken); rr.Code != http.StatusUnauthorized {
		t.Fatalf("reveal without token status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	revealReq := managerRequest(http.MethodGet, "/api/reveal?token="+testToken+"&path=app:token", "")
	rr := serveManager(server, revealReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("reveal status = %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "secret-value") {
		t.Fatalf("reveal response missing value: %s", rr.Body.String())
	}
}

func TestManagerWritesUseEncryptedVaultAndRejectBadOrigin(t *testing.T) {
	server, vaultPath := newTestServer(t)
	payload := `{"path":"app:token","value":"secret-value","env":"APP_TOKEN"}`
	badOrigin := managerRequest(http.MethodPost, "/api/secrets?token="+testToken, payload)
	badOrigin.Header.Set("Origin", "http://evil.test")
	if rr := serveManager(server, badOrigin); rr.Code != http.StatusForbidden {
		t.Fatalf("bad origin status = %d, want %d", rr.Code, http.StatusForbidden)
	}

	writeReq := managerRequest(http.MethodPost, "/api/secrets?token="+testToken, payload)
	writeReq.Header.Set("Origin", "http://127.0.0.1:4321")
	if rr := serveManager(server, writeReq); rr.Code != http.StatusOK {
		t.Fatalf("write status = %d body=%s", rr.Code, rr.Body.String())
	}

	updateReq := managerRequest(http.MethodPut, "/api/secrets?token="+testToken, `{"path":"app:token","value":"updated-value","env":"APP_TOKEN"}`)
	updateReq.Header.Set("Origin", "http://127.0.0.1:4321")
	if rr := serveManager(server, updateReq); rr.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", rr.Code, rr.Body.String())
	}

	content, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	for _, forbidden := range [][]byte{[]byte("secret-value"), []byte("updated-value"), []byte("app:token"), []byte("APP_TOKEN")} {
		if bytes.Contains(content, forbidden) {
			t.Fatalf("encrypted vault contains plaintext %q", forbidden)
		}
	}

	revealReq := managerRequest(http.MethodGet, "/api/reveal?token="+testToken+"&path=app:token", "")
	rr := serveManager(server, revealReq)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "updated-value") {
		t.Fatalf("updated reveal failed status=%d body=%s", rr.Code, rr.Body.String())
	}

	deleteReq := managerRequest(http.MethodDelete, "/api/secrets?token="+testToken+"&path=app:token", "")
	deleteReq.Header.Set("Origin", "http://127.0.0.1:4321")
	if rr := serveManager(server, deleteReq); rr.Code != http.StatusOK {
		t.Fatalf("delete status = %d body=%s", rr.Code, rr.Body.String())
	}

	listReq := managerRequest(http.MethodGet, "/api/secrets?token="+testToken, "")
	rr = serveManager(server, listReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rr.Code, rr.Body.String())
	}
	var decoded struct {
		Secrets []SecretInfo `json:"secrets"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(decoded.Secrets) != 0 {
		t.Fatalf("deleted secret still listed: %+v", decoded.Secrets)
	}
}
