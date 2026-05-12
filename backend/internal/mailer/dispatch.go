// dispatch.go — high-level `SendTemplate` API that combines (key, locale,
// vars) с DB override lookup + embedded baseline fallback + render +
// underlying Mailer.Send.
//
// Поведение (см. .team/product-acceptance/admin-email-templates.md):
//   1. Если TemplateStore != nil, попытаться Lookup(key, locale).
//   2. Если row найден — Render через html/template + Mailer.Send. Done.
//   3. Если row nil или Lookup error → log warn и use embedded baseline
//      (BaselineTemplate). Это match'ит "fallback на DB outage" сценарий
//      (Scenario 3) и "fallback на missing row" (Scenario 2).
//   4. Если embedded baseline тоже nil (неизвестный key) — return error.
//
// Caller (teams/invites.go etc.) может by-passнуть этот pathway и
// продолжать вызывать `mailer.InviteEmail(...)` + `Mailer.Send(...)` если
// нужен старый sprintf-стиль (backward compat). Phase 3 backfill optional.

package mailer

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

// Dispatcher — облегчённая обёртка вокруг Mailer + TemplateStore. Caller'у
// нужен один объект, через который шлёт `(to, key, locale, vars)`.
type Dispatcher struct {
	Mailer Mailer
	Store  TemplateStore
	Logger *zap.Logger
}

// SendTemplate — основной entry point. Возвращает error только если ВСЕ
// fallback пути исчерпаны (i.e., нет baseline для key). DB outage НЕ
// возвращает error — silent fallback на embedded.
func (d Dispatcher) SendTemplate(ctx context.Context, to, key string, locale Locale, vars map[string]any) error {
	if d.Mailer == nil {
		return ErrNoMailer
	}
	if !IsSupportedTemplateKey(key) {
		return errors.New("mailer: unknown template key: " + key)
	}
	loc := NormalizeLocale(string(locale))

	tpl, err := LookupOrFallback(ctx, d.Store, key, loc)
	if err != nil {
		// DB ошибка — log + fallback на baseline.
		if d.Logger != nil {
			d.Logger.Warn("email_template_db_unavailable",
				zap.String("key", key),
				zap.String("locale", string(loc)),
				zap.Error(err),
			)
		}
		tpl = nil
	}
	if tpl == nil {
		tpl = BaselineTemplate(key, loc)
		if tpl == nil {
			return errors.New("mailer: no baseline for " + key + ":" + string(loc))
		}
	}
	rendered, err := Render(*tpl, vars)
	if err != nil {
		return err
	}
	return d.Mailer.Send(ctx, to, rendered.Subject, rendered.HTML, rendered.Text)
}
