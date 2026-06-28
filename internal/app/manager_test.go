package app

import "testing"

func TestListenLoopbackRejectsNonLoopback(t *testing.T) {
	listener, err := ListenLoopback("192.0.2.10:0")
	if err == nil {
		listener.Close()
		t.Fatalf("expected non-loopback manager address to fail")
	}
}

func TestListenLoopbackAcceptsLoopback(t *testing.T) {
	listener, err := ListenLoopback("127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen loopback: %v", err)
	}
	listener.Close()
}

func TestManagerTokenIsGenerated(t *testing.T) {
	first, err := ManagerToken()
	if err != nil {
		t.Fatalf("manager token: %v", err)
	}
	second, err := ManagerToken()
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
