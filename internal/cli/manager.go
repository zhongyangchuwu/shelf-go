package cli

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/manager"
)

func newManagerCmd() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "manager",
		Short: "Start a localhost vault manager",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, vault, err := loadVault(cmd)
			if err != nil {
				return err
			}
			listener, err := listenLoopback(addr)
			if err != nil {
				return err
			}
			defer listener.Close()
			token, err := managerToken()
			if err != nil {
				return err
			}
			server, err := manager.NewServer(vault, token, listener.Addr().String())
			if err != nil {
				return err
			}
			httpServer := &http.Server{Handler: server.Handler()}
			errCh := make(chan error, 1)
			go func() {
				if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
					errCh <- err
					return
				}
				errCh <- nil
			}()
			fmt.Fprintf(cmd.OutOrStdout(), "manager: http://%s/?token=%s\n", listener.Addr().String(), token)
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(sigCh)
			select {
			case err := <-errCh:
				return err
			case <-sigCh:
				return httpServer.Close()
			}
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:0", "Loopback address to listen on")
	return cmd
}

func listenLoopback(addr string) (net.Listener, error) {
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

func managerToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}
