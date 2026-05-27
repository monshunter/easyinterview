package auth_test

import (
	"errors"
	"net/smtp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func TestSMTPDeliveryWriterSendsMagicLinkThroughSMTP(t *testing.T) {
	secrets := auth.NewDevMailSink(auth.DevMailSinkOptions{})
	secrets.PutDeliverySecret("auth_challenge:challenge-1", "raw-magic-token")
	var captured struct {
		addr string
		from string
		to   []string
		msg  string
	}
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "127.0.0.1:1025",
		FromAddress:     "noreply@easyinterview.local",
		VerifyBaseURL:   "http://127.0.0.1:5173/auth/verify",
		DeliverySecrets: secrets,
		LookupChallengeEmail: func(challengeID string) (string, error) {
			if challengeID != "challenge-1" {
				t.Fatalf("lookup challenge id = %q", challengeID)
			}
			return "candidate@example.test", nil
		},
		SendMail: func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
			captured.addr = addr
			captured.from = from
			captured.to = append([]string(nil), to...)
			captured.msg = string(msg)
			return nil
		},
	})

	payload := emailPayload(t, "challenge-1", "auth_challenge:challenge-1")
	if err := writer.Write(payload); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if captured.addr != "127.0.0.1:1025" {
		t.Fatalf("smtp addr = %q", captured.addr)
	}
	if captured.from != "noreply@easyinterview.local" {
		t.Fatalf("smtp from = %q", captured.from)
	}
	if len(captured.to) != 1 || captured.to[0] != "candidate@example.test" {
		t.Fatalf("smtp recipients = %#v", captured.to)
	}
	for _, want := range []string{
			"To: candidate@example.test",
			"Subject: EasyInterview sign-in link",
			"return to the app and finish sign-in",
			"http://127.0.0.1:5173/auth/verify?token=raw-magic-token",
	} {
		if !strings.Contains(captured.msg, want) {
			t.Fatalf("message missing %q:\n%s", want, captured.msg)
		}
	}
	for _, forbidden := range []string{"auth_challenge:challenge-1", "deliverySecretRef"} {
		if strings.Contains(captured.msg, forbidden) {
			t.Fatalf("message leaked internal dispatch field %q:\n%s", forbidden, captured.msg)
		}
	}
}

func TestSMTPDeliveryWriterRequiresStoredDeliverySecret(t *testing.T) {
	called := false
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "127.0.0.1:1025",
		FromAddress:     "noreply@easyinterview.local",
			VerifyBaseURL:   "http://127.0.0.1:5173/auth/verify",
		DeliverySecrets: auth.NewDevMailSink(auth.DevMailSinkOptions{}),
		LookupChallengeEmail: func(string) (string, error) {
			return "candidate@example.test", nil
		},
		SendMail: func(string, smtp.Auth, string, []string, []byte) error {
			called = true
			return nil
		},
	})

	err := writer.Write(emailPayload(t, "challenge-1", "missing-secret-ref"))
	if err == nil {
		t.Fatal("expected missing delivery secret to fail")
	}
	if called {
		t.Fatal("SMTP send must not run without a stored delivery secret")
	}
}

func TestSMTPDeliveryWriterDoesNotExposeLookupErrorDetails(t *testing.T) {
	secrets := auth.NewDevMailSink(auth.DevMailSinkOptions{})
	secrets.PutDeliverySecret("auth_challenge:challenge-1", "raw-magic-token")
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "127.0.0.1:1025",
		FromAddress:     "noreply@easyinterview.local",
		VerifyBaseURL:   "http://127.0.0.1:5173/auth/verify",
		DeliverySecrets: secrets,
		LookupChallengeEmail: func(string) (string, error) {
			return "", errors.New("candidate@example.test raw-magic-token")
		},
		SendMail: func(string, smtp.Auth, string, []string, []byte) error {
			t.Fatal("SMTP send must not run when recipient lookup fails")
			return nil
		},
	})

	err := writer.Write(emailPayload(t, "challenge-1", "auth_challenge:challenge-1"))
	if err == nil {
		t.Fatal("expected lookup failure")
	}
	for _, forbidden := range []string{"candidate@example.test", "raw-magic-token"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("error leaked %q: %v", forbidden, err)
		}
	}
}

func TestSQLChallengeEmailLookupReturnsAuthChallengeEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("select email from auth_challenges").
		WithArgs("challenge-1").
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("candidate@example.test"))

	email, err := auth.SQLChallengeEmailLookup(db)("challenge-1")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if email != "candidate@example.test" {
		t.Fatalf("email = %q", email)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func emailPayload(t *testing.T, challengeID string, secretRef string) jobs.EmailDispatchPayload {
	t.Helper()
	payload, err := jobs.BuildEmailDispatchPayload(map[string]string{
		"authChallengeId":   challengeID,
		"templateKey":       "auth_magic_link",
		"locale":            "en",
		"deliverySecretRef": secretRef,
		"dedupeKey":         "dedupe-hash",
	})
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
