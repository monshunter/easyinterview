package auth

import (
	"bufio"
	"crypto/tls"
	"net"
	"strings"
	"testing"
)

func TestSMTPTLSConfigRequiresTLS12AndServerName(t *testing.T) {
	cfg := newSMTPTLSConfig("smtp.example.test")
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("MinVersion = %d, want TLS 1.2", cfg.MinVersion)
	}
	if cfg.ServerName != "smtp.example.test" {
		t.Fatalf("ServerName = %q", cfg.ServerName)
	}
}

func TestSMTPStartTLSFailsWhenServerDoesNotAdvertiseExtension(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("220 fake smtp\r\n"))
		_, _ = bufio.NewReader(conn).ReadString('\n')
		_, _ = conn.Write([]byte("250-fake\r\n250 AUTH PLAIN\r\n"))
	}()

	err = sendSMTPEnvelope(SMTPEnvelope{Addr: listener.Addr().String(), TLSMode: SMTPTLSStartTLS})
	if err == nil || !strings.Contains(err.Error(), "starttls") {
		t.Fatalf("send error = %v", err)
	}
	<-done
}

func TestSMTPTransportRejectsInvalidAddress(t *testing.T) {
	err := sendSMTPEnvelope(SMTPEnvelope{Addr: "missing-port", TLSMode: SMTPTLSImplicit})
	if err == nil {
		t.Fatal("expected invalid SMTP address to fail")
	}
	if safe, ok := err.(interface{ SafeMessage() string }); !ok || safe.SafeMessage() != "smtp connect failed" {
		t.Fatalf("safe error = %#v", err)
	}
}
