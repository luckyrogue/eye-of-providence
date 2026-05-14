package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Mailer interface {
	Send(ctx context.Context, to, subject, html, text string) error
}

var resendURL = "https://api.resend.com/emails"

type ResendMailer struct {
	apiKey string
	from   string
	client *http.Client
	logger *zap.Logger
}

func NewResend(apiKey, from string, logger *zap.Logger) *ResendMailer {
	return &ResendMailer{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{Timeout: 5 * time.Second},
		logger: logger,
	}
}

type resendReq struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

type resendErrResp struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (r *ResendMailer) Send(ctx context.Context, to, subject, html, text string) error {
	body, err := json.Marshal(resendReq{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
		Text:    text,
	})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var er resendErrResp
		_ = json.Unmarshal(raw, &er)

		if r.logger != nil {
			r.logger.Warn("resend send failed",
				zap.String("to", to),
				zap.Int("status", resp.StatusCode),
				zap.String("name", er.Name),
				zap.String("message", er.Message),
			)
		}
		return fmt.Errorf("resend %d: %s", resp.StatusCode, er.Message)
	}
	return nil
}

type NoopMailer struct {
	logger *zap.Logger
}

func Noop(logger *zap.Logger) *NoopMailer {
	return &NoopMailer{logger: logger}
}

func (n *NoopMailer) Send(_ context.Context, to, subject, _, _ string) error {
	if n.logger != nil {
		n.logger.Info("mail.noop send", zap.String("to", to), zap.String("subject", subject))
	}
	return nil
}

var ErrNoMailer = errors.New("mailer not configured")
