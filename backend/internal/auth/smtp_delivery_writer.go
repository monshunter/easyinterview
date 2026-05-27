package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/mail"
	"net/smtp"
	"net/url"
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
	link, err := buildMagicLink(w.verifyBaseURL, token)
	if err != nil {
		return fmt.Errorf("magic link build failed")
	}
	msg, err := buildMagicLinkMessage(from, to, link)
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

func buildMagicLink(base string, token string) (string, error) {
	if strings.TrimSpace(base) == "" || strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("base url or token missing")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func buildMagicLinkMessage(from string, to string, link string) (string, error) {
	headers := []struct {
		name  string
		value string
	}{
		{name: "From", value: from},
		{name: "To", value: to},
		{name: "Subject", value: "EasyInterview sign-in link"},
		{name: "MIME-Version", value: "1.0"},
		{name: "Content-Type", value: "text/plain; charset=UTF-8"},
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
	b.WriteString("Open this EasyInterview sign-in link in the same browser to return to the app and finish sign-in:\r\n\r\n")
	b.WriteString(link)
	b.WriteString("\r\n")
	return b.String(), nil
}

var _ DeliveryWriter = (*SMTPDeliveryWriter)(nil)
