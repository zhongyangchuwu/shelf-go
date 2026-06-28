package manager

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

const sessionCookie = "shelf_manager_token"

type Server struct {
	service *app.SecretService
	token   string
	host    string
}

type SecretInfo = app.SecretSummary

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

func NewServer(service *app.SecretService, token, host string) (*Server, error) {
	if service == nil {
		return nil, fmt.Errorf("manager service is required")
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	return &Server{service: service, token: token, host: host}, nil
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
	info, err := s.service.SecretInfo(path)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) listSecrets(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.ListSecrets(r.URL.Query().Get("q"))
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
	value, err := s.service.RevealSecret(payload.Path)
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
	err := s.service.WriteSecret(r.Method == http.MethodPut, app.WriteSecretRequest{
		OldPath:     payload.OldPath,
		Path:        payload.Path,
		Value:       payload.Value,
		Env:         payload.Env,
		Description: payload.Description,
		Tags:        payload.Tags,
		Force:       payload.Force,
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
	err := s.service.DeleteSecret(path)
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
