package manager

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
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
	Path        string   `json:"path"`
	Value       string   `json:"value"`
	Env         string   `json:"env,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Force       bool     `json:"force,omitempty"`
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
	mux.HandleFunc("/api/secrets", s.handleSecrets)
	mux.HandleFunc("/api/reveal", s.handleReveal)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			items = append(items, SecretInfo{Path: path, Env: secret.Env, Description: secret.Description, Tags: append([]string(nil), secret.Tags...), ValueSet: len(secret.Value) > 0})
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
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	var value string
	err := s.vault.Read(func(st *store.Store) error {
		secret, ok := st.Get(path)
		if !ok {
			return fmt.Errorf("secret not found: %s", path)
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
	writeJSON(w, http.StatusOK, map[string]string{"path": path, "value": value})
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
	value, err := store.ParseValue(payload.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	secret := store.Secret{Value: value, Env: payload.Env, Description: payload.Description, Tags: payload.Tags}
	err = s.vault.Update(func(st *store.Store) error {
		force := payload.Force || r.Method == http.MethodPut
		return st.Set(payload.Path, secret, force)
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
	return origin == "http://"+s.host || origin == "http://localhost" || strings.HasPrefix(origin, "http://127.0.0.1:") || strings.HasPrefix(origin, "http://[::1]:")
}

func isUnsafe(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
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
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html><head><meta charset="utf-8"><title>Shelf Vault Manager</title></head>
<body>
<h1>Shelf Vault Manager</h1>
<form id="search"><input name="q" placeholder="Search paths, env, description, tags"><button>Search</button></form>
<ul id="results"></ul>
<script>
async function load(q='') {
  const res = await fetch('/api/secrets?q=' + encodeURIComponent(q));
  const data = await res.json();
  results.innerHTML = '';
  for (const item of data.secrets || []) {
    const li = document.createElement('li');
    li.textContent = item.path + (item.env ? ' -> ' + item.env : '') + (item.description ? ' - ' + item.description : '');
    const button = document.createElement('button');
    button.textContent = 'Reveal';
    button.onclick = async () => {
      const r = await fetch('/api/reveal?path=' + encodeURIComponent(item.path));
      const revealed = await r.json();
      alert(revealed.value);
    };
    li.appendChild(button);
    results.appendChild(li);
  }
}
search.onsubmit = event => { event.preventDefault(); load(new FormData(search).get('q') || ''); };
load();
</script>
</body></html>`))
