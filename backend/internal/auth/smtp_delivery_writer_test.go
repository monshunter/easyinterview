package auth_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func TestSMTPDeliveryWriterSendsLoginCodeThroughSMTP(t *testing.T) {
	secrets := auth.NewDevMailSink(auth.DevMailSinkOptions{})
	if err := secrets.PutDeliverySecret(context.Background(), "auth_challenge:challenge-1", "123456", auth.ChallengeTTL); err != nil {
		t.Fatalf("PutDeliverySecret: %v", err)
	}
	var captured struct {
		addr string
		from string
		to   []string
		msg  string
	}
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "127.0.0.1:1025",
		TLSMode:         auth.SMTPTLSNone,
		FromAddress:     "noreply@easyinterview.local",
		VerifyBaseURL:   "http://127.0.0.1:5173/auth/verify",
		DeliverySecrets: secrets,
		LookupChallengeEmail: func(_ context.Context, challengeID string) (string, error) {
			if challengeID != "challenge-1" {
				t.Fatalf("lookup challenge id = %q", challengeID)
			}
			return "candidate@example.test", nil
		},
		Send: func(_ context.Context, envelope auth.SMTPEnvelope) error {
			captured.addr = envelope.Addr
			captured.from = envelope.From
			captured.to = append([]string(nil), envelope.To...)
			captured.msg = string(envelope.Message)
			return nil
		},
	})

	payload := emailPayload(t, "challenge-1", "auth_challenge:challenge-1")
	if err := writer.Write(context.Background(), payload); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, ok, err := secrets.GetDeliverySecret(context.Background(), "auth_challenge:challenge-1"); err != nil || ok {
		t.Fatalf("success must delete delivery secret: ok=%v err=%v", ok, err)
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
		"Subject: EasyInterview sign-in code",
		"123456",
		"expires in 5 minutes",
		"multipart/alternative",
	} {
		if !strings.Contains(captured.msg, want) {
			t.Fatalf("message missing %q:\n%s", want, captured.msg)
		}
	}
	for _, forbidden := range []string{"auth/verify?token=", "raw-magic-token", "http://127.0.0.1:5173/auth/verify"} {
		if strings.Contains(captured.msg, forbidden) {
			t.Fatalf("message leaked forbidden auth link/token %q:\n%s", forbidden, captured.msg)
		}
	}
	for _, forbidden := range []string{"auth_challenge:challenge-1", "deliverySecretRef"} {
		if strings.Contains(captured.msg, forbidden) {
			t.Fatalf("message leaked internal dispatch field %q:\n%s", forbidden, captured.msg)
		}
	}
}

func TestSMTPDeliveryWriterEncodesLocalizedMessageAsStandardsCompliantMIME(t *testing.T) {
	secrets := auth.NewDevMailSink(auth.DevMailSinkOptions{})
	if err := secrets.PutDeliverySecret(context.Background(), "auth_challenge:challenge-1", "123456", auth.ChallengeTTL); err != nil {
		t.Fatalf("PutDeliverySecret: %v", err)
	}
	var message []byte
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "smtp.example.test:587",
		FromAddress:     "noreply@example.test",
		DeliverySecrets: secrets,
		LookupChallengeEmail: func(context.Context, string) (string, error) {
			return "candidate@example.test", nil
		},
		Send: func(_ context.Context, envelope auth.SMTPEnvelope) error {
			message = append([]byte(nil), envelope.Message...)
			return nil
		},
	})
	payload, err := jobs.BuildEmailDispatchPayload(map[string]string{
		"authChallengeId":   "challenge-1",
		"templateKey":       "auth_login_code",
		"locale":            "zh-CN",
		"deliverySecretRef": "auth_challenge:challenge-1",
		"dedupeKey":         "dedupe-hash",
	})
	if err != nil {
		t.Fatalf("BuildEmailDispatchPayload: %v", err)
	}

	if err := writer.Write(context.Background(), payload); err != nil {
		t.Fatalf("Write: %v", err)
	}
	parsed, err := mail.ReadMessage(bytes.NewReader(message))
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	rawSubject := parsed.Header.Get("Subject")
	if !strings.HasPrefix(rawSubject, "=?UTF-8?") {
		t.Fatalf("localized Subject must use RFC 2047 encoded-word, got %q", rawSubject)
	}
	decodedSubject, err := new(mime.WordDecoder).DecodeHeader(rawSubject)
	if err != nil {
		t.Fatalf("DecodeHeader: %v", err)
	}
	if decodedSubject != "EasyInterview 登录验证码" {
		t.Fatalf("decoded Subject = %q", decodedSubject)
	}

	mediaType, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("ParseMediaType: %v", err)
	}
	if mediaType != "multipart/alternative" {
		t.Fatalf("Content-Type = %q", mediaType)
	}
	reader := multipart.NewReader(parsed.Body, params["boundary"])
	decodedParts := make(map[string]string)
	for {
		part, err := reader.NextRawPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("NextRawPart: %v", err)
		}
		partType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			t.Fatalf("parse part Content-Type: %v", err)
		}
		if got := part.Header.Get("Content-Transfer-Encoding"); got != "quoted-printable" {
			t.Fatalf("%s Content-Transfer-Encoding = %q", partType, got)
		}
		body, err := io.ReadAll(quotedprintable.NewReader(part))
		if err != nil {
			t.Fatalf("decode %s: %v", partType, err)
		}
		decodedParts[partType] = string(body)
	}
	for _, partType := range []string{"text/plain", "text/html"} {
		body, ok := decodedParts[partType]
		if !ok {
			t.Fatalf("missing MIME part %s", partType)
		}
		for _, want := range []string{"你的 EasyInterview 登录验证码", "123456", "请在 EasyInterview 输入这 6 位验证码。验证码 5 分钟内有效。"} {
			if !strings.Contains(body, want) {
				t.Fatalf("decoded %s missing %q: %s", partType, want, body)
			}
		}
	}
}

func TestSMTPDeliveryWriterRetainsSecretWhenSendFails(t *testing.T) {
	secrets := auth.NewDevMailSink(auth.DevMailSinkOptions{})
	if err := secrets.PutDeliverySecret(context.Background(), "auth_challenge:challenge-1", "123456", auth.ChallengeTTL); err != nil {
		t.Fatalf("PutDeliverySecret: %v", err)
	}
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "smtp.example.test:587",
		FromAddress:     "noreply@example.test",
		DeliverySecrets: secrets,
		LookupChallengeEmail: func(context.Context, string) (string, error) {
			return "candidate@example.test", nil
		},
		Send: func(context.Context, auth.SMTPEnvelope) error { return errors.New("provider unavailable") },
	})

	if err := writer.Write(context.Background(), emailPayload(t, "challenge-1", "auth_challenge:challenge-1")); err == nil {
		t.Fatal("expected send failure")
	}
	if code, ok, err := secrets.GetDeliverySecret(context.Background(), "auth_challenge:challenge-1"); err != nil || !ok || code != "123456" {
		t.Fatalf("send failure must retain delivery secret: code=%q ok=%v err=%v", code, ok, err)
	}
}

func TestSMTPDeliveryWriterIgnoresDeleteFailureAfterSuccessfulSend(t *testing.T) {
	secrets := &lifecycleSecretStore{
		secrets: map[string]string{"auth_challenge:challenge-1": "123456"},
		delErr:  errors.New("redis://user:password@private-host:6379 123456"),
	}
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "smtp.example.test:587",
		FromAddress:     "noreply@example.test",
		DeliverySecrets: secrets,
		LookupChallengeEmail: func(context.Context, string) (string, error) {
			return "candidate@example.test", nil
		},
		Send: func(context.Context, auth.SMTPEnvelope) error { return nil },
	})

	if err := writer.Write(context.Background(), emailPayload(t, "challenge-1", "auth_challenge:challenge-1")); err != nil {
		t.Fatalf("successful SMTP send must not retry when cleanup fails: %v", err)
	}
	if len(secrets.deleted) != 1 {
		t.Fatalf("delete attempts = %d, want 1", len(secrets.deleted))
	}
}

func TestSMTPDeliveryWriterRequiresStoredDeliverySecret(t *testing.T) {
	called := false
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "127.0.0.1:1025",
		FromAddress:     "noreply@easyinterview.local",
		VerifyBaseURL:   "http://127.0.0.1:5173/auth/verify",
		DeliverySecrets: auth.NewDevMailSink(auth.DevMailSinkOptions{}),
		LookupChallengeEmail: func(context.Context, string) (string, error) {
			return "candidate@example.test", nil
		},
		Send: func(context.Context, auth.SMTPEnvelope) error {
			called = true
			return nil
		},
	})

	err := writer.Write(context.Background(), emailPayload(t, "challenge-1", "missing-secret-ref"))
	if err == nil {
		t.Fatal("expected missing delivery secret to fail")
	}
	if called {
		t.Fatal("SMTP send must not run without a stored delivery secret")
	}
}

func TestSMTPDeliveryWriterDoesNotExposeLookupErrorDetails(t *testing.T) {
	secrets := auth.NewDevMailSink(auth.DevMailSinkOptions{})
	if err := secrets.PutDeliverySecret(context.Background(), "auth_challenge:challenge-1", "123456", auth.ChallengeTTL); err != nil {
		t.Fatalf("PutDeliverySecret: %v", err)
	}
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "127.0.0.1:1025",
		FromAddress:     "noreply@easyinterview.local",
		VerifyBaseURL:   "http://127.0.0.1:5173/auth/verify",
		DeliverySecrets: secrets,
		LookupChallengeEmail: func(context.Context, string) (string, error) {
			return "", errors.New("candidate@example.test 123456")
		},
		Send: func(context.Context, auth.SMTPEnvelope) error {
			t.Fatal("SMTP send must not run when recipient lookup fails")
			return nil
		},
	})

	err := writer.Write(context.Background(), emailPayload(t, "challenge-1", "auth_challenge:challenge-1"))
	if err == nil {
		t.Fatal("expected lookup failure")
	}
	for _, forbidden := range []string{"candidate@example.test", "123456"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("error leaked %q: %v", forbidden, err)
		}
	}
}

func TestSMTPDeliveryWriterPassesTLSAndAuthenticationWithoutLeakingSecrets(t *testing.T) {
	for _, tt := range []struct {
		name string
		mode auth.SMTPTLSMode
	}{
		{name: "starttls", mode: auth.SMTPTLSStartTLS},
		{name: "implicit tls", mode: auth.SMTPTLSImplicit},
	} {
		t.Run(tt.name, func(t *testing.T) {
			secrets := auth.NewDevMailSink(auth.DevMailSinkOptions{})
			if err := secrets.PutDeliverySecret(context.Background(), "auth_challenge:challenge-1", "123456", auth.ChallengeTTL); err != nil {
				t.Fatalf("PutDeliverySecret: %v", err)
			}
			var captured auth.SMTPEnvelope
			writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
				SMTPAddr:        "smtp.example.test:587",
				FromAddress:     "noreply@example.test",
				Username:        "mailer",
				Password:        "smtp-secret",
				TLSMode:         tt.mode,
				DeliverySecrets: secrets,
				LookupChallengeEmail: func(context.Context, string) (string, error) {
					return "candidate@example.test", nil
				},
				Send: func(_ context.Context, envelope auth.SMTPEnvelope) error {
					captured = envelope
					return errors.New("provider rejected credentials smtp-secret candidate@example.test 123456")
				},
			})

			err := writer.Write(context.Background(), emailPayload(t, "challenge-1", "auth_challenge:challenge-1"))
			if err == nil {
				t.Fatal("expected delivery failure")
			}
			if captured.TLSMode != tt.mode || captured.Username != "mailer" || captured.Password != "smtp-secret" {
				t.Fatalf("SMTP envelope config = %#v", captured)
			}
			for _, forbidden := range []string{"smtp-secret", "candidate@example.test", "123456"} {
				if strings.Contains(err.Error(), forbidden) {
					t.Fatalf("delivery error leaked %q: %v", forbidden, err)
				}
			}
		})
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

	email, err := auth.SQLChallengeEmailLookup(db)(context.Background(), "challenge-1")
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
		"templateKey":       "auth_login_code",
		"locale":            "en",
		"deliverySecretRef": secretRef,
		"dedupeKey":         "dedupe-hash",
	})
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
