package cli

import "testing"

func TestListenLoopbackRejectsNonLoopback(t *testing.T) {
	listener, err := listenLoopback("192.0.2.10:0")
	if err == nil {
		listener.Close()
		t.Fatalf("expected non-loopback manager address to fail")
	}
}

func TestManagerTokenIsGenerated(t *testing.T) {
	first, err := managerToken()
	if err != nil {
		t.Fatalf("manager token: %v", err)
	}
	second, err := managerToken()
	if err != nil {
		t.Fatalf("manager token: %v", err)
	}
	if first == "" || second == "" {
		t.Fatalf("manager token should not be empty")
	}
	if first == second {
		t.Fatalf("manager token should be random")
	}
}

func TestRootIncludesManagerCommand(t *testing.T) {
	cmd := NewRootCmd()
	found := false
	for _, child := range cmd.Commands() {
		if child.Name() == "manager" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("root command missing manager subcommand")
	}
}
