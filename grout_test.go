package grout

import (
	"errors"
	"net"
	"testing"
)

func TestConnect_Success(t *testing.T) {
	s := newMockServer(t, helloHandler("grout 0.16.0"))
	defer s.close()

	client, err := Connect(WithSocketPath(s.sockPath))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer client.Close()

	if v := client.APIVersion(); v != CurrentAPIVersion {
		t.Errorf("APIVersion() = %d, want %d", v, CurrentAPIVersion)
	}
	if v := client.ServerVersion(); v != "grout 0.16.0" {
		t.Errorf("ServerVersion() = %q, want %q", v, "grout 0.16.0")
	}
}

func TestConnect_ConnectionRefused(t *testing.T) {
	_, err := Connect(WithSocketPath("/tmp/nonexistent-grout-test.sock"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrConnection) {
		t.Errorf("expected ErrConnection, got: %v", err)
	}
}

func TestConnect_DefaultSocketPath(t *testing.T) {
	cfg := &clientConfig{socketPath: DefaultSocketPath}
	if cfg.socketPath != "/run/grout.sock" {
		t.Errorf("default socket path = %q, want %q", cfg.socketPath, "/run/grout.sock")
	}
}

func TestConnect_WithSocketPath(t *testing.T) {
	cfg := &clientConfig{socketPath: DefaultSocketPath}
	WithSocketPath("/custom/path.sock")(cfg)
	if cfg.socketPath != "/custom/path.sock" {
		t.Errorf("socket path = %q, want %q", cfg.socketPath, "/custom/path.sock")
	}
}

func TestClose_Idempotent(t *testing.T) {
	s := newMockServer(t, helloHandler("grout 0.16.0"))
	defer s.close()

	client, err := Connect(WithSocketPath(s.sockPath))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	if err := client.Close(); err != nil {
		t.Errorf("first Close() error: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Errorf("second Close() error: %v", err)
	}
}

func TestConnect_HandshakeFailure(t *testing.T) {
	s := newMockServer(t, func(conn net.Conn) {
		conn.Close()
	})
	defer s.close()

	_, err := Connect(WithSocketPath(s.sockPath))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrConnection) {
		t.Errorf("expected ErrConnection, got: %v", err)
	}
}
