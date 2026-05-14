package mailer

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

type Dispatcher struct {
	Mailer Mailer
	Store  TemplateStore
	Logger *zap.Logger
}

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
