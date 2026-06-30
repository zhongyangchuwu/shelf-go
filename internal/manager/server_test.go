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
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultfile"
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
	v, err := vaultfile.NewVault(vaultPath, vaultfile.VaultOptions{Recipients: []string{identity.Recipient().String()}, IdentityPaths: []string{identityPath}})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	service, err := app.NewSecretService(v)
	if err != nil {
		t.Fatalf("new manager service: %v", err)
	}
	server, err := NewServer(service, testToken, "127.0.0.1:4321")
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

func authorizedManagerRequest(method, target, body string) *http.Request {
	req := managerRequest(method, target, body)
	req.Header.Set("X-Shelf-Token", testToken)
	return req
}

func unsafeManagerRequest(method, target, body string) *http.Request {
	req := authorizedManagerRequest(method, target, body)
	req.Header.Set("Origin", "http://127.0.0.1:4321")
	return req
}

func serveManager(server *Server, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	return rr
}

func requireNoStore(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	if got := rr.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestManagerRequiresTokenAndValidHost(t *testing.T) {
	server, _ := newTestServer(t)

	missingToken := managerRequest(http.MethodGet, "/api/secrets", "")
	if rr := serveManager(server, missingToken); rr.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	badHost := authorizedManagerRequest(http.MethodGet, "/api/secrets", "")
	badHost.Host = "evil.test"
	if rr := serveManager(server, badHost); rr.Code != http.StatusForbidden {
		t.Fatalf("bad host status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestManagerAcceptsLocalhostAndAlternateTokenTransports(t *testing.T) {
	server, _ := newTestServer(t)

	localhost := authorizedManagerRequest(http.MethodGet, "/api/secrets", "")
	localhost.Host = "localhost:4321"
	if rr := serveManager(server, localhost); rr.Code != http.StatusOK {
		t.Fatalf("localhost status = %d, want %d", rr.Code, http.StatusOK)
	}

	headerToken := authorizedManagerRequest(http.MethodGet, "/api/secrets", "")
	if rr := serveManager(server, headerToken); rr.Code != http.StatusOK {
		t.Fatalf("header token status = %d, want %d", rr.Code, http.StatusOK)
	}

	cookieToken := managerRequest(http.MethodGet, "/api/secrets", "")
	cookieToken.AddCookie(&http.Cookie{Name: sessionCookie, Value: testToken})
	if rr := serveManager(server, cookieToken); rr.Code != http.StatusOK {
		t.Fatalf("cookie token status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestManagerQueryTokenSetsStrictCookieAndRedirects(t *testing.T) {
	server, _ := newTestServer(t)
	req := managerRequest(http.MethodGet, "/?token="+testToken, "")
	rr := serveManager(server, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("index status = %d, want %d", rr.Code, http.StatusSeeOther)
	}
	if got := rr.Header().Get("Location"); got != "/" {
		t.Fatalf("Location = %q, want /", got)
	}
	if strings.Contains(rr.Header().Get("Location"), testToken) {
		t.Fatalf("redirect leaked token in location: %s", rr.Header().Get("Location"))
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
	requireNoStore(t, rr)
}
func TestManagerIndexServesEmbeddedWorkbench(t *testing.T) {
	server, _ := newTestServer(t)
	req := authorizedManagerRequest(http.MethodGet, "/", "")
	rr := serveManager(server, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("index status = %d, want %d", rr.Code, http.StatusOK)
	}
	requireNoStore(t, rr)
	body := rr.Body.String()
	for _, want := range []string{"Shelf manager", "Add secret", "Reveal value", "Copy value", "Delete secret", "Local encrypted vault"} {
		if !strings.Contains(body, want) {
			t.Fatalf("index missing %q", want)
		}
	}
	for _, forbidden := range []string{"https://", "http://", "localStorage", "sessionStorage"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("index contains forbidden browser dependency/storage %q", forbidden)
		}
	}
}

func TestManagerListSearchExcludesSecretValues(t *testing.T) {
	server, _ := newTestServer(t)
	payload := `{"path":"app:token","value":"secret-value","env":"APP_TOKEN","description":"primary api token","tags":["api"]}`
	writeReq := unsafeManagerRequest(http.MethodPost, "/api/secrets", payload)
	if rr := serveManager(server, writeReq); rr.Code != http.StatusOK {
		t.Fatalf("write status = %d body=%s", rr.Code, rr.Body.String())
	}

	listReq := authorizedManagerRequest(http.MethodGet, "/api/secrets?q=api", "")
	rr := serveManager(server, listReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rr.Code, rr.Body.String())
	}
	requireNoStore(t, rr)
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

func TestManagerSecretDetailExcludesSecretValue(t *testing.T) {
	server, _ := newTestServer(t)
	payload := `{"path":"app:token","value":"secret-value","env":"APP_TOKEN","description":"primary api token","tags":["api"]}`
	writeReq := unsafeManagerRequest(http.MethodPost, "/api/secrets", payload)
	if rr := serveManager(server, writeReq); rr.Code != http.StatusOK {
		t.Fatalf("write status = %d body=%s", rr.Code, rr.Body.String())
	}

	detailReq := authorizedManagerRequest(http.MethodGet, "/api/secret?path=app:token", "")
	rr := serveManager(server, detailReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("detail status = %d body=%s", rr.Code, rr.Body.String())
	}
	requireNoStore(t, rr)
	body := rr.Body.String()
	for _, want := range []string{"app:token", "APP_TOKEN", "primary api token", "\"value_set\":true"} {
		if !strings.Contains(body, want) {
			t.Fatalf("detail response missing %q: %s", want, body)
		}
	}
	if strings.Contains(body, "secret-value") {
		t.Fatalf("detail response leaked secret value: %s", body)
	}
}

func TestManagerRevealRequiresPostAndIsExplicit(t *testing.T) {
	server, _ := newTestServer(t)
	payload := `{"path":"app:token","value":"secret-value","env":"APP_TOKEN"}`
	writeReq := unsafeManagerRequest(http.MethodPost, "/api/secrets", payload)
	if rr := serveManager(server, writeReq); rr.Code != http.StatusOK {
		t.Fatalf("write status = %d body=%s", rr.Code, rr.Body.String())
	}

	withoutToken := managerRequest(http.MethodPost, "/api/reveal", `{"path":"app:token"}`)
	withoutToken.Header.Set("Origin", "http://127.0.0.1:4321")
	if rr := serveManager(server, withoutToken); rr.Code != http.StatusUnauthorized {
		t.Fatalf("reveal without token status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	getReveal := authorizedManagerRequest(http.MethodGet, "/api/reveal?path=app:token", "")
	if rr := serveManager(server, getReveal); rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET reveal status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}

	revealReq := unsafeManagerRequest(http.MethodPost, "/api/reveal", `{"path":"app:token"}`)
	rr := serveManager(server, revealReq)
	if rr.Code != http.StatusOK {
		t.Fatalf("reveal status = %d body=%s", rr.Code, rr.Body.String())
	}
	requireNoStore(t, rr)
	if !strings.Contains(rr.Body.String(), "secret-value") {
		t.Fatalf("reveal response missing value: %s", rr.Body.String())
	}
}

func TestManagerWritesUseEncryptedVaultAndRejectBadOrigin(t *testing.T) {
	server, vaultPath := newTestServer(t)
	payload := `{"path":"app:token","value":"secret-value","env":"APP_TOKEN"}`
	badOrigin := authorizedManagerRequest(http.MethodPost, "/api/secrets", payload)
	badOrigin.Header.Set("Origin", "http://evil.test")
	if rr := serveManager(server, badOrigin); rr.Code != http.StatusForbidden {
		t.Fatalf("bad origin status = %d, want %d", rr.Code, http.StatusForbidden)
	}

	writeReq := unsafeManagerRequest(http.MethodPost, "/api/secrets", payload)
	if rr := serveManager(server, writeReq); rr.Code != http.StatusOK {
		t.Fatalf("write status = %d body=%s", rr.Code, rr.Body.String())
	}

	updateReq := unsafeManagerRequest(http.MethodPut, "/api/secrets", `{"path":"app:token","value":"updated-value","env":"APP_TOKEN"}`)
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

	revealReq := unsafeManagerRequest(http.MethodPost, "/api/reveal", `{"path":"app:token"}`)
	rr := serveManager(server, revealReq)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "updated-value") {
		t.Fatalf("updated reveal failed status=%d body=%s", rr.Code, rr.Body.String())
	}

	deleteReq := unsafeManagerRequest(http.MethodDelete, "/api/secrets?path=app:token", "")
	if rr := serveManager(server, deleteReq); rr.Code != http.StatusOK {
		t.Fatalf("delete status = %d body=%s", rr.Code, rr.Body.String())
	}

	listReq := authorizedManagerRequest(http.MethodGet, "/api/secrets", "")
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

func TestManagerPutPreservesValueAndRenamesExplicitly(t *testing.T) {
	server, _ := newTestServer(t)
	writeReq := unsafeManagerRequest(http.MethodPost, "/api/secrets", `{"path":"app:token","value":"secret-value","env":"APP_TOKEN"}`)
	if rr := serveManager(server, writeReq); rr.Code != http.StatusOK {
		t.Fatalf("write status = %d body=%s", rr.Code, rr.Body.String())
	}

	updateReq := unsafeManagerRequest(http.MethodPut, "/api/secrets", `{"old_path":"app:token","path":"app:renamed","env":"RENAMED_TOKEN","description":"metadata only","tags":["api","local"]}`)
	if rr := serveManager(server, updateReq); rr.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", rr.Code, rr.Body.String())
	}

	revealReq := unsafeManagerRequest(http.MethodPost, "/api/reveal", `{"path":"app:renamed"}`)
	rr := serveManager(server, revealReq)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "secret-value") {
		t.Fatalf("renamed reveal failed status=%d body=%s", rr.Code, rr.Body.String())
	}

	oldReveal := unsafeManagerRequest(http.MethodPost, "/api/reveal", `{"path":"app:token"}`)
	if rr := serveManager(server, oldReveal); rr.Code != http.StatusBadRequest {
		t.Fatalf("old path reveal status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
