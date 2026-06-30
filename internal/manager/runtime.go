package manager

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Runtime struct {
	listener   net.Listener
	token      string
	httpServer *http.Server
}

func Open(service SecretService, addr string) (*Runtime, error) {
	listener, err := ListenLoopback(addr)
	if err != nil {
		return nil, err
	}
	token, err := Token()
	if err != nil {
		listener.Close()
		return nil, err
	}
	server, err := NewServer(service, token, listener.Addr().String())
	if err != nil {
		listener.Close()
		return nil, err
	}
	return &Runtime{listener: listener, token: token, httpServer: &http.Server{Handler: server.Handler()}}, nil
}

func (r *Runtime) Addr() string {
	return r.listener.Addr().String()
}

func (r *Runtime) Token() string {
	return r.token
}

func (r *Runtime) Close() error {
	return r.listener.Close()
}

func (r *Runtime) ServeUntilSignal(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := r.httpServer.Serve(r.listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		return r.httpServer.Close()
	case <-ctx.Done():
		return r.httpServer.Close()
	}
}

func ListenLoopback(addr string) (net.Listener, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, fmt.Errorf("manager address must be loopback: %s", addr)
	}
	return net.Listen("tcp", addr)
}

func Token() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}
