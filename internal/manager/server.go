package manager

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/render"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

const sessionCookie = "shelf_manager_token"

type Server struct {
	vault *store.Vault
	token string
	host  string
}

type SecretInfo struct {
	Path        string   `json:"path"`
	Env         string   `json:"env,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags"`
	ValueSet    bool     `json:"value_set"`
}

type secretPayload struct {
	OldPath     string   `json:"old_path,omitempty"`
	Path        string   `json:"path"`
	Value       *string  `json:"value,omitempty"`
	Env         string   `json:"env,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Force       bool     `json:"force,omitempty"`
}

type pathPayload struct {
	Path string `json:"path"`
}

func NewServer(vault *store.Vault, token, host string) (*Server, error) {
	if vault == nil {
		return nil, fmt.Errorf("vault is required")
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	return &Server{vault: vault, token: token, host: host}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/secret", s.handleSecret)
	mux.HandleFunc("/api/secrets", s.handleSecrets)
	mux.HandleFunc("/api/reveal", s.handleReveal)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setNoStore(w)
		if !s.validHost(r.Host) {
			http.Error(w, "invalid host", http.StatusForbidden)
			return
		}
		if !s.validToken(r) {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		if isUnsafe(r.Method) && !s.validOrigin(r) {
			http.Error(w, "invalid origin", http.StatusForbidden)
			return
		}
		if r.URL.Query().Get("token") == s.token {
			http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: s.token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				cleanURL := *r.URL
				query := cleanURL.Query()
				query.Del("token")
				cleanURL.RawQuery = query.Encode()
				http.Redirect(w, r, cleanURL.String(), http.StatusSeeOther)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = indexTemplate.Execute(w, nil)
}

func (s *Server) handleSecrets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listSecrets(w, r)
	case http.MethodPost, http.MethodPut:
		s.writeSecret(w, r)
	case http.MethodDelete:
		s.deleteSecret(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	var info SecretInfo
	err := s.vault.Read(func(st *store.Store) error {
		secret, ok := st.Get(path)
		if !ok {
			return fmt.Errorf("secret not found: %s", path)
		}
		info = newSecretInfo(path, secret)
		return nil
	})
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) listSecrets(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	var items []SecretInfo
	err := s.vault.Read(func(st *store.Store) error {
		paths := st.List("")
		items = make([]SecretInfo, 0, len(paths))
		for _, path := range paths {
			secret, ok := st.Get(path)
			if !ok || query != "" && !matchesSecret(query, path, secret) {
				continue
			}
			items = append(items, newSecretInfo(path, secret))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
		return nil
	})
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"secrets": items})
}

func (s *Server) handleReveal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var payload pathPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if payload.Path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	var value string
	err := s.vault.Read(func(st *store.Store) error {
		secret, ok := st.Get(payload.Path)
		if !ok {
			return fmt.Errorf("secret not found: %s", payload.Path)
		}
		v, err := render.ValueString(secret.Value)
		if err != nil {
			return err
		}
		value = v
		return nil
	})
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": payload.Path, "value": value})
}

func (s *Server) writeSecret(w http.ResponseWriter, r *http.Request) {
	var payload secretPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if payload.Path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodPost && payload.Value == nil {
		http.Error(w, "value is required", http.StatusBadRequest)
		return
	}
	err := s.vault.Update(func(st *store.Store) error {
		secret := store.Secret{Env: payload.Env, Description: payload.Description, Tags: payload.Tags}
		if r.Method == http.MethodPut {
			oldPath := payload.OldPath
			if oldPath == "" {
				oldPath = payload.Path
			}
			existing, ok := st.Get(oldPath)
			if !ok {
				return fmt.Errorf("secret not found: %s", oldPath)
			}
			secret.Value = existing.Value
			if payload.Value != nil {
				value, err := store.ParseValue(*payload.Value)
				if err != nil {
					return err
				}
				secret.Value = value
			}
			id, err := store.ParseSecretID(payload.Path)
			if err != nil {
				return err
			}
			return st.Update(oldPath, id, secret)
		}
		value, err := store.ParseValue(*payload.Value)
		if err != nil {
			return err
		}
		secret.Value = value
		return st.Set(payload.Path, secret, payload.Force)
	})
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": payload.Path})
}

func (s *Server) deleteSecret(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	err := s.vault.Update(func(st *store.Store) error {
		if !st.Delete(path) {
			return fmt.Errorf("secret not found: %s", path)
		}
		return nil
	})
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": path})
}

func (s *Server) validHost(host string) bool {
	if host == s.host {
		return true
	}
	requestHost, _, err := net.SplitHostPort(host)
	if err != nil {
		requestHost = host
	}
	allowedHost, _, err := net.SplitHostPort(s.host)
	if err != nil {
		allowedHost = s.host
	}
	if requestHost == allowedHost {
		return requestHost == "localhost" || net.ParseIP(requestHost).IsLoopback()
	}
	return (requestHost == "localhost" && net.ParseIP(allowedHost).IsLoopback()) || (allowedHost == "localhost" && net.ParseIP(requestHost).IsLoopback())
}

func (s *Server) validToken(r *http.Request) bool {
	got := r.URL.Query().Get("token")
	if got == "" {
		got = r.Header.Get("X-Shelf-Token")
	}
	if got == "" {
		if cookie, err := r.Cookie(sessionCookie); err == nil {
			got = cookie.Value
		}
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.token)) == 1
}

func (s *Server) validOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "http" {
		return false
	}
	return s.validHost(parsed.Host)
}

func isUnsafe(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}

func newSecretInfo(path string, secret store.Secret) SecretInfo {
	return SecretInfo{Path: path, Env: secret.Env, Description: secret.Description, Tags: append([]string(nil), secret.Tags...), ValueSet: len(secret.Value) > 0}
}

func matchesSecret(query, path string, secret store.Secret) bool {
	if strings.Contains(strings.ToLower(path), query) || strings.Contains(strings.ToLower(secret.Env), query) || strings.Contains(strings.ToLower(secret.Description), query) {
		return true
	}
	for _, tag := range secret.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func serverError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	setNoStore(w)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func setNoStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
}
