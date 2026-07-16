package auth

import (
	"bufio"
	"context"
	"crypto/tls"
	"net"
	"strings"
	"testing"
	"time"
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

	err = sendSMTPEnvelope(context.Background(), SMTPEnvelope{Addr: listener.Addr().String(), TLSMode: SMTPTLSStartTLS})
	if err == nil || !strings.Contains(err.Error(), "starttls") {
		t.Fatalf("send error = %v", err)
	}
	<-done
}

func TestSMTPTransportRejectsInvalidAddress(t *testing.T) {
	err := sendSMTPEnvelope(context.Background(), SMTPEnvelope{Addr: "missing-port", TLSMode: SMTPTLSImplicit})
	if err == nil {
		t.Fatal("expected invalid SMTP address to fail")
	}
	if safe, ok := err.(interface{ SafeMessage() string }); !ok || safe.SafeMessage() != "smtp connect failed" {
		t.Fatalf("safe error = %#v", err)
	}
}

func TestSMTPTransportHonorsCancellationAfterConnect(t *testing.T) {
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
		_, _ = bufio.NewReader(conn).ReadString('\n')
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	err = sendSMTPEnvelope(ctx, SMTPEnvelope{Addr: listener.Addr().String(), TLSMode: SMTPTLSNone})
	if err == nil {
		t.Fatal("expected canceled SMTP session to fail")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("SMTP cancellation took %s, want less than 1s", elapsed)
	}
	<-done
}

func TestSMTPTransportTreatsQuitFailureAfterAcceptedDataAsSuccess(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	serverErr := make(chan error, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverErr <- acceptErr
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		write := func(response string) error {
			_, writeErr := conn.Write([]byte(response))
			return writeErr
		}
		readCommand := func(prefix string) error {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return readErr
			}
			if !strings.HasPrefix(line, prefix) {
				return &smtpTestProtocolError{got: line, wantPrefix: prefix}
			}
			return nil
		}
		if err := write("220 fake smtp\r\n"); err != nil {
			serverErr <- err
			return
		}
		if err := readCommand("EHLO "); err != nil {
			serverErr <- err
			return
		}
		if err := write("250 fake smtp\r\n"); err != nil {
			serverErr <- err
			return
		}
		for _, command := range []string{"MAIL FROM:", "RCPT TO:"} {
			if err := readCommand(command); err != nil {
				serverErr <- err
				return
			}
			if err := write("250 ok\r\n"); err != nil {
				serverErr <- err
				return
			}
		}
		if err := readCommand("DATA"); err != nil {
			serverErr <- err
			return
		}
		if err := write("354 send data\r\n"); err != nil {
			serverErr <- err
			return
		}
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				serverErr <- readErr
				return
			}
			if line == ".\r\n" {
				break
			}
		}
		if err := write("250 queued\r\n"); err != nil {
			serverErr <- err
			return
		}
		if err := readCommand("QUIT"); err != nil {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	err = sendSMTPEnvelope(context.Background(), SMTPEnvelope{
		Addr:    listener.Addr().String(),
		From:    "noreply@example.test",
		To:      []string{"candidate@example.test"},
		Message: []byte("Subject: test\r\n\r\nbody\r\n"),
		TLSMode: SMTPTLSNone,
	})
	if err != nil {
		t.Fatalf("accepted DATA must not be retried after QUIT failure: %v", err)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("fake SMTP server: %v", err)
	}
}

type smtpTestProtocolError struct {
	got        string
	wantPrefix string
}

func (e *smtpTestProtocolError) Error() string {
	return "SMTP command " + strings.TrimSpace(e.got) + " does not start with " + e.wantPrefix
}
