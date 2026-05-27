package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

type ChallengeEmailLookup func(challengeID string) (string, error)

type SMTPSendMailFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

type SMTPDeliveryWriterOptions struct {
	SMTPAddr             string
	FromAddress          string
	VerifyBaseURL        string
	DeliverySecrets      DeliverySecretStore
	LookupChallengeEmail ChallengeEmailLookup
	SendMail             SMTPSendMailFunc
}

type SMTPDeliveryWriter struct {
	smtpAddr             string
	fromAddress          string
	verifyBaseURL        string
	deliverySecrets      DeliverySecretStore
	lookupChallengeEmail ChallengeEmailLookup
	sendMail             SMTPSendMailFunc
}

func NewSMTPDeliveryWriter(opts SMTPDeliveryWriterOptions) *SMTPDeliveryWriter {
	sendMail := opts.SendMail
	if sendMail == nil {
		sendMail = smtp.SendMail
	}
	return &SMTPDeliveryWriter{
		smtpAddr:             strings.TrimSpace(opts.SMTPAddr),
		fromAddress:          strings.TrimSpace(opts.FromAddress),
		verifyBaseURL:        strings.TrimSpace(opts.VerifyBaseURL),
		deliverySecrets:      opts.DeliverySecrets,
		lookupChallengeEmail: opts.LookupChallengeEmail,
		sendMail:             sendMail,
	}
}

func SQLChallengeEmailLookup(db *sql.DB) ChallengeEmailLookup {
	return func(challengeID string) (string, error) {
		if db == nil {
			return "", fmt.Errorf("auth challenge email lookup db is nil")
		}
		var email string
		if err := db.QueryRowContext(context.Background(), `
select email from auth_challenges where id = $1`,
			challengeID).Scan(&email); err != nil {
			return "", err
		}
		return email, nil
	}
}

func (w *SMTPDeliveryWriter) Write(payload jobs.EmailDispatchPayload) error {
	if w == nil {
		return fmt.Errorf("smtp delivery writer is nil")
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
	token, ok := w.deliverySecrets.GetDeliverySecret(secretRef)
	if !ok || token == "" {
		return fmt.Errorf("delivery secret unavailable")
	}
	if w.lookupChallengeEmail == nil {
		return fmt.Errorf("challenge email lookup unavailable")
	}
	recipient, err := w.lookupChallengeEmail(challengeID)
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
	if err := w.sendMail(w.smtpAddr, nil, from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp email delivery failed")
	}
	return nil
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
		{name: "Subject", value: subject},
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
	b.WriteString("--ei_auth_code\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(title)
	b.WriteString("\r\n\r\n")
	b.WriteString(code)
	b.WriteString("\r\n\r\n")
	b.WriteString(instruction)
	b.WriteString("\r\n\r\n")
	b.WriteString("--ei_auth_code\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	b.WriteString(`<!doctype html><html><body style="margin:0;background:#f7f3ea;font-family:Inter,Arial,sans-serif;color:#241f19;">`)
	b.WriteString(`<div style="max-width:520px;margin:0 auto;padding:28px 24px;">`)
	b.WriteString(`<p style="font-size:12px;letter-spacing:.08em;text-transform:uppercase;color:#8a6d3b;margin:0 0 12px;">EasyInterview</p>`)
	b.WriteString(`<h1 style="font-size:22px;line-height:1.3;margin:0 0 18px;font-weight:650;">`)
	b.WriteString(title)
	b.WriteString(`</h1>`)
	b.WriteString(`<div style="display:inline-block;border:1px solid #ded3bd;border-radius:8px;background:#fffaf0;padding:14px 18px;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:32px;letter-spacing:6px;font-weight:700;">`)
	b.WriteString(code)
	b.WriteString(`</div>`)
	b.WriteString(`<p style="font-size:14px;line-height:1.6;color:#665a4a;margin:18px 0 0;">`)
	b.WriteString(instruction)
	b.WriteString(`</p>`)
	b.WriteString(`</div></body></html>`)
	b.WriteString("\r\n--ei_auth_code--\r\n")
	return b.String(), nil
}

var _ DeliveryWriter = (*SMTPDeliveryWriter)(nil)
