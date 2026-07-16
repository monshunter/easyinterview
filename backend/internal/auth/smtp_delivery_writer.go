package auth

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
	"unicode"

	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

type ChallengeEmailLookup func(context.Context, string) (string, error)

type SMTPTLSMode string

const (
	SMTPTLSNone     SMTPTLSMode = "none"
	SMTPTLSStartTLS SMTPTLSMode = "starttls"
	SMTPTLSImplicit SMTPTLSMode = "tls"
)

type SMTPEnvelope struct {
	Addr     string
	From     string
	To       []string
	Message  []byte
	Username string
	Password string
	TLSMode  SMTPTLSMode
}

type SMTPSendFunc func(context.Context, SMTPEnvelope) error

const smtpSessionTimeout = 30 * time.Second

type safeDeliveryError struct{ stage string }

func (e safeDeliveryError) Error() string       { return e.SafeMessage() }
func (e safeDeliveryError) SafeMessage() string { return "smtp " + e.stage + " failed" }

func smtpFailure(stage string) error { return safeDeliveryError{stage: stage} }

type SMTPDeliveryWriterOptions struct {
	SMTPAddr             string
	FromAddress          string
	Username             string
	Password             string
	TLSMode              SMTPTLSMode
	VerifyBaseURL        string
	DeliverySecrets      DeliverySecretStore
	LookupChallengeEmail ChallengeEmailLookup
	Send                 SMTPSendFunc
}

type SMTPDeliveryWriter struct {
	smtpAddr             string
	fromAddress          string
	username             string
	password             string
	tlsMode              SMTPTLSMode
	verifyBaseURL        string
	deliverySecrets      DeliverySecretStore
	lookupChallengeEmail ChallengeEmailLookup
	send                 SMTPSendFunc
}

func NewSMTPDeliveryWriter(opts SMTPDeliveryWriterOptions) *SMTPDeliveryWriter {
	send := opts.Send
	if send == nil {
		send = sendSMTPEnvelope
	}
	return &SMTPDeliveryWriter{
		smtpAddr:             strings.TrimSpace(opts.SMTPAddr),
		fromAddress:          strings.TrimSpace(opts.FromAddress),
		username:             strings.TrimSpace(opts.Username),
		password:             opts.Password,
		tlsMode:              opts.TLSMode,
		verifyBaseURL:        strings.TrimSpace(opts.VerifyBaseURL),
		deliverySecrets:      opts.DeliverySecrets,
		lookupChallengeEmail: opts.LookupChallengeEmail,
		send:                 send,
	}
}

func (w *SMTPDeliveryWriter) TLSMode() SMTPTLSMode {
	if w == nil {
		return ""
	}
	return w.tlsMode
}

func (w *SMTPDeliveryWriter) UsesAuthentication() bool {
	return w != nil && w.username != ""
}

func SQLChallengeEmailLookup(db *sql.DB) ChallengeEmailLookup {
	return func(ctx context.Context, challengeID string) (string, error) {
		if db == nil {
			return "", fmt.Errorf("auth challenge email lookup db is nil")
		}
		var email string
		if err := db.QueryRowContext(ctx, `
select email from auth_challenges where id = $1`,
			challengeID).Scan(&email); err != nil {
			return "", err
		}
		return email, nil
	}
}

func (w *SMTPDeliveryWriter) Write(ctx context.Context, payload jobs.EmailDispatchPayload) error {
	if w == nil {
		return fmt.Errorf("smtp delivery writer is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	challengeID := strings.TrimSpace(payload["authChallengeId"])
	if challengeID == "" {
		return fmt.Errorf("%s payload missing authChallengeId", jobs.JobTypeEmailDispatch)
	}
	secretRef := strings.TrimSpace(payload["deliverySecretRef"])
	if secretRef == "" {
		return fmt.Errorf("%s payload missing deliverySecretRef", jobs.JobTypeEmailDispatch)
	}
	if w.deliverySecrets == nil {
		return fmt.Errorf("delivery secret store unavailable")
	}
	token, ok, err := w.deliverySecrets.GetDeliverySecret(ctx, secretRef)
	if err != nil {
		return fmt.Errorf("delivery secret lookup failed")
	}
	if !ok || token == "" {
		return fmt.Errorf("delivery secret unavailable")
	}
	if w.lookupChallengeEmail == nil {
		return fmt.Errorf("challenge email lookup unavailable")
	}
	recipient, err := w.lookupChallengeEmail(ctx, challengeID)
	if err != nil {
		return fmt.Errorf("challenge recipient lookup failed")
	}
	to, err := parseEmailAddress(recipient)
	if err != nil {
		return fmt.Errorf("challenge recipient lookup failed")
	}
	from, err := parseEmailAddress(w.fromAddress)
	if err != nil {
		return fmt.Errorf("smtp from address invalid")
	}
	msg, err := buildLoginCodeMessage(from, to, token, payload["locale"])
	if err != nil {
		return fmt.Errorf("email message build failed")
	}
	if w.smtpAddr == "" {
		return fmt.Errorf("smtp address missing")
	}
	if err := w.send(ctx, SMTPEnvelope{
		Addr: w.smtpAddr, From: from, To: []string{to}, Message: []byte(msg),
		Username: w.username, Password: w.password, TLSMode: w.tlsMode,
	}); err != nil {
		var safe interface{ SafeMessage() string }
		if errors.As(err, &safe) {
			return err
		}
		return smtpFailure("delivery")
	}
	// Delivery succeeded. Cleanup is best-effort: a delete failure must not
	// retry SMTP and send the same code twice; the store TTL is the fallback.
	_ = w.deliverySecrets.DeleteDeliverySecret(ctx, secretRef)
	return nil
}

func sendSMTPEnvelope(ctx context.Context, envelope SMTPEnvelope) error {
	host, _, err := net.SplitHostPort(envelope.Addr)
	if err != nil {
		return smtpFailure("connect")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	sessionCtx, cancel := context.WithTimeout(ctx, smtpSessionTimeout)
	defer cancel()
	if envelope.TLSMode != SMTPTLSNone && envelope.TLSMode != SMTPTLSStartTLS && envelope.TLSMode != SMTPTLSImplicit {
		return smtpFailure("tls-mode")
	}
	tlsConfig := newSMTPTLSConfig(host)
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(sessionCtx, "tcp", envelope.Addr)
	if err != nil {
		return smtpFailure("connect")
	}
	defer conn.Close()
	if deadline, ok := sessionCtx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return smtpFailure("connect")
		}
	}
	stopCancellation := context.AfterFunc(sessionCtx, func() {
		_ = conn.SetDeadline(time.Now())
	})
	defer stopCancellation()

	if envelope.TLSMode == SMTPTLSImplicit {
		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.HandshakeContext(sessionCtx); err != nil {
			return smtpFailure("connect")
		}
		conn = tlsConn
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return smtpFailure("greeting")
	}
	defer client.Close()
	if envelope.TLSMode == SMTPTLSStartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return smtpFailure("starttls")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return smtpFailure("starttls")
		}
	}
	if envelope.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", envelope.Username, envelope.Password, host)); err != nil {
			return smtpFailure("auth")
		}
	}
	if err := client.Mail(envelope.From); err != nil {
		return smtpFailure("mail-from")
	}
	for _, recipient := range envelope.To {
		if err := client.Rcpt(recipient); err != nil {
			return smtpFailure("recipient")
		}
	}
	w, err := client.Data()
	if err != nil {
		return smtpFailure("data")
	}
	if _, err := w.Write(envelope.Message); err != nil {
		w.Close()
		return smtpFailure("data")
	}
	if err := w.Close(); err != nil {
		return smtpFailure("data")
	}
	// DATA has been accepted. QUIT is only connection cleanup; retrying after
	// a QUIT response failure can send the same code twice.
	_ = client.Quit()
	return nil
}

func newSMTPTLSConfig(host string) *tls.Config {
	return &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
}

func parseEmailAddress(raw string) (string, error) {
	if strings.ContainsAny(raw, "\r\n") {
		return "", fmt.Errorf("email address contains newline")
	}
	parsed, err := mail.ParseAddress(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	if parsed.Address == "" || strings.ContainsAny(parsed.Address, "\r\n") {
		return "", fmt.Errorf("email address invalid")
	}
	return parsed.Address, nil
}

func buildLoginCodeMessage(from string, to string, code string, locale string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "", fmt.Errorf("login code missing")
	}
	subject := "EasyInterview sign-in code"
	title := "Your EasyInterview sign-in code"
	instruction := "Enter this 6-digit code in EasyInterview. It expires in 5 minutes."
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(locale)), "zh") {
		subject = "EasyInterview 登录验证码"
		title = "你的 EasyInterview 登录验证码"
		instruction = "请在 EasyInterview 输入这 6 位验证码。验证码 5 分钟内有效。"
	}
	headers := []struct {
		name  string
		value string
	}{
		{name: "From", value: from},
		{name: "To", value: to},
		{name: "Subject", value: encodeMIMEHeader(subject)},
		{name: "MIME-Version", value: "1.0"},
		{name: "Content-Type", value: `multipart/alternative; boundary="ei_auth_code"`},
	}
	var b strings.Builder
	for _, h := range headers {
		if strings.ContainsAny(h.value, "\r\n") {
			return "", fmt.Errorf("header %s contains newline", h.name)
		}
		b.WriteString(h.name)
		b.WriteString(": ")
		b.WriteString(h.value)
		b.WriteString("\r\n")
	}
	b.WriteString("\r\n")
	plainBody := title + "\r\n\r\n" + code + "\r\n\r\n" + instruction + "\r\n"
	if err := writeQuotedPrintablePart(&b, "text/plain", plainBody); err != nil {
		return "", err
	}
	htmlBody := `<!doctype html><html><body style="margin:0;background:#f7f3ea;font-family:Inter,Arial,sans-serif;color:#241f19;">` +
		`<div style="max-width:520px;margin:0 auto;padding:28px 24px;">` +
		`<p style="font-size:12px;letter-spacing:.08em;text-transform:uppercase;color:#8a6d3b;margin:0 0 12px;">EasyInterview</p>` +
		`<h1 style="font-size:22px;line-height:1.3;margin:0 0 18px;font-weight:650;">` + title + `</h1>` +
		`<div style="display:inline-block;border:1px solid #ded3bd;border-radius:8px;background:#fffaf0;padding:14px 18px;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:32px;letter-spacing:6px;font-weight:700;">` + code + `</div>` +
		`<p style="font-size:14px;line-height:1.6;color:#665a4a;margin:18px 0 0;">` + instruction + `</p>` +
		`</div></body></html>`
	if err := writeQuotedPrintablePart(&b, "text/html", htmlBody); err != nil {
		return "", err
	}
	b.WriteString("\r\n--ei_auth_code--\r\n")
	return b.String(), nil
}

func encodeMIMEHeader(value string) string {
	if strings.IndexFunc(value, func(r rune) bool { return r > unicode.MaxASCII }) < 0 {
		return value
	}
	return mime.QEncoding.Encode("UTF-8", value)
}

func writeQuotedPrintablePart(b *strings.Builder, contentType string, body string) error {
	b.WriteString("--ei_auth_code\r\n")
	b.WriteString("Content-Type: ")
	b.WriteString(contentType)
	b.WriteString("; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	encoder := quotedprintable.NewWriter(b)
	if _, err := encoder.Write([]byte(body)); err != nil {
		return fmt.Errorf("encode %s body: %w", contentType, err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("close %s body encoder: %w", contentType, err)
	}
	return nil
}

var _ DeliveryWriter = (*SMTPDeliveryWriter)(nil)
