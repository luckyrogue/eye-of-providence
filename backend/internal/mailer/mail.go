// Package mailer — minimal transactional email клиент через Resend HTTP API.
//
// Названо `mailer` (не `mail`), чтобы не конфликтовать с stdlib `net/mail`,
// который используется в teams-handler'е для validateEmail.
//
// Зачем не SDK: Resend Go-SDK тащит лишние deps (otel, jwt, etc.) ради
// одного POST /emails. Свой клиент — 60 LoC и testable через httptest.
//
// Использование:
//
//	m := mailer.NewResend(cfg.ResendAPIKey, cfg.MailFrom)
//	// или, если API key пустой (dev) — Noop, который только логирует:
//	if cfg.ResendAPIKey == "" {
//	    m = mailer.Noop(logger)
//	}
//	m.Send(ctx, "user@example.com", "Subject", htmlBody, textBody)
//
// Шаблоны держим inline — пока их 3, sprintf'ом проще чем добавлять
// templating engine.
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

// Mailer — интерфейс. Используется handler'ами; Resend / Noop — реализации.
type Mailer interface {
	Send(ctx context.Context, to, subject, html, text string) error
}

// resendURL — endpoint Resend API. Var, чтобы тесты могли подменить.
var resendURL = "https://api.resend.com/emails"

// ResendMailer — отправка через api.resend.com.
type ResendMailer struct {
	apiKey string
	from   string
	client *http.Client
	logger *zap.Logger
}

// NewResend — клиент с разумными дефолтами (5s timeout).
// from — RFC-5322 строка типа `Eye of Providence <noreply@eop.dev>`.
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
		// Логируем — отдадим caller'у только короткую ошибку, чтобы он мог
		// решить fail/swallow.
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

// NoopMailer — fallback когда EOP_RESEND_API_KEY пустой. Не падает, не шлёт,
// логирует «бы отправили». Для dev и для feature-flag'а в проде если
// провайдер ляжет.
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

// ErrNoMailer — когда code-path требует Mailer но его нет (caller сам решает,
// fail или swallow).
var ErrNoMailer = errors.New("mailer not configured")
