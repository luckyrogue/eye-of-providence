package mailer

import "strings"

func BaselineTemplate(key string, locale Locale) *Template {
	loc := NormalizeLocale(string(locale))
	switch key {
	case TemplateKeyPasswordReset:
		c := resetCopies[loc]
		return &Template{
			Key:      key,
			Locale:   string(loc),
			Subject:  c.Subject,
			BodyHTML: passwordResetHTMLTpl(c),
			BodyText: passwordResetTextTpl(c),
		}
	case TemplateKeyTeamInvite:
		c := inviteCopies[loc]
		return &Template{
			Key:      key,
			Locale:   string(loc),
			Subject:  formatTpl(c.Subject, "{{.TeamName}}"),
			BodyHTML: teamInviteHTMLTpl(c),
			BodyText: teamInviteTextTpl(c),
		}
	case TemplateKeySubscriptionActivated:

		return subscriptionActivatedBaseline(loc)
	}
	return nil
}

func formatTpl(src, placeholder string) string {
	return replaceFirstSubst(src, placeholder)
}

func replaceFirstSubst(s, placeholder string) string {
	return strings.Replace(s, "%s", placeholder, 1)
}

func passwordResetHTMLTpl(c resetCopy) string {
	return "<!doctype html>\n<html><body style=\"" + baseStyle + "\">\n" +
		"  <h2 style=\"margin:0 0 12px;font-weight:700;\">" + c.Heading + "</h2>\n" +
		"  <p>" + c.Lead + "</p>\n" +
		"  <p style=\"margin:24px 0;\">\n" +
		"    <a href=\"{{.ResetURL}}\" style=\"" + buttonStyle + "\">" + c.Button + "</a>\n" +
		"  </p>\n" +
		"  <p style=\"" + subtleStyle + "\">" + c.ValidNote + "</p>\n" +
		"</body></html>"
}

func passwordResetTextTpl(c resetCopy) string {
	return c.TextHeader + "\n\n" + c.TextLead + "\n{{.ResetURL}}\n\n" + c.TextValid
}

func teamInviteHTMLTpl(c inviteCopy) string {
	heading := replaceFirstSubst(c.Heading, "{{.TeamName}}")

	greetingWith := replaceFirstSubst(replaceFirstSubst(c.Greeting, "{{.InviterName}}"), "{{.TeamName}}")
	greetingNoBy := replaceFirstSubst(c.GreetingNoBy, "{{.TeamName}}")

	return "<!doctype html>\n<html><body style=\"" + baseStyle + "\">\n" +
		"  <h2 style=\"margin:0 0 12px;font-weight:700;\">" + heading + "</h2>\n" +
		"  <p>{{if .InviterName}}" + greetingWith + "{{else}}" + greetingNoBy + "{{end}}</p>\n" +
		"  <p style=\"margin:24px 0;\">\n" +
		"    <a href=\"{{.AcceptURL}}\" style=\"" + buttonStyle + "\">" + c.ButtonAccept + "</a>\n" +
		"  </p>\n" +
		"  <p style=\"" + subtleStyle + "\">" + c.IgnoreNote + "<br>" + c.ValidNote + "</p>\n" +
		"</body></html>"
}

func teamInviteTextTpl(c inviteCopy) string {
	header := replaceFirstSubst(c.TextHeader, "{{.TeamName}}")
	lineWith := replaceFirstSubst(replaceFirstSubst(c.TextLine, "{{.InviterName}}"), "{{.TeamName}}")
	lineNoBy := replaceFirstSubst(c.TextLineNoBy, "{{.TeamName}}")
	return header + "\n\n{{if .InviterName}}" + lineWith + "{{else}}" + lineNoBy + "{{end}}\n\n" +
		c.TextAccept + " {{.AcceptURL}}\n\n" + c.TextIgnore
}

func subscriptionActivatedBaseline(loc Locale) *Template {
	type sub struct{ subj, heading, lead, button, valid, txtHead, txtLead, txtValid string }
	copies := map[Locale]sub{
		LocaleRU: {
			subj:     "Подписка активирована · Eye of Providence",
			heading:  "Подписка активирована",
			lead:     "Привет, {{.Name}}. Тариф <strong>{{.PlanName}}</strong> активен на твоей команде.",
			button:   "Открыть биллинг",
			valid:    "Если есть вопросы — пиши на support@eop.dev.",
			txtHead:  "Подписка активирована · Eye of Providence",
			txtLead:  "Тариф {{.PlanName}} активен на твоей команде.",
			txtValid: "Открыть биллинг: {{.BillingURL}}",
		},
		LocaleEN: {
			subj:     "Subscription activated · Eye of Providence",
			heading:  "Subscription activated",
			lead:     "Hi, {{.Name}}. The <strong>{{.PlanName}}</strong> plan is active on your team.",
			button:   "Open billing",
			valid:    "Questions? Reach us at support@eop.dev.",
			txtHead:  "Subscription activated · Eye of Providence",
			txtLead:  "The {{.PlanName}} plan is active on your team.",
			txtValid: "Open billing: {{.BillingURL}}",
		},
		LocaleKK: {
			subj:     "Жазылым іске қосылды · Eye of Providence",
			heading:  "Жазылым іске қосылды",
			lead:     "Сәлем, {{.Name}}. <strong>{{.PlanName}}</strong> тарифі сіздің командаңызда белсенді.",
			button:   "Биллингті ашу",
			valid:    "Сұрақтар болса — support@eop.dev.",
			txtHead:  "Жазылым іске қосылды · Eye of Providence",
			txtLead:  "{{.PlanName}} тарифі сіздің командаңызда белсенді.",
			txtValid: "Биллингті ашу: {{.BillingURL}}",
		},
		LocaleES: {
			subj:     "Suscripción activada · Eye of Providence",
			heading:  "Suscripción activada",
			lead:     "Hola, {{.Name}}. El plan <strong>{{.PlanName}}</strong> está activo en tu equipo.",
			button:   "Abrir facturación",
			valid:    "¿Preguntas? Escríbenos a support@eop.dev.",
			txtHead:  "Suscripción activada · Eye of Providence",
			txtLead:  "El plan {{.PlanName}} está activo en tu equipo.",
			txtValid: "Abrir facturación: {{.BillingURL}}",
		},
	}
	c := copies[loc]
	html := "<!doctype html>\n<html><body style=\"" + baseStyle + "\">\n" +
		"  <h2 style=\"margin:0 0 12px;font-weight:700;\">" + c.heading + "</h2>\n" +
		"  <p>" + c.lead + "</p>\n" +
		"  <p style=\"margin:24px 0;\">\n" +
		"    <a href=\"{{.BillingURL}}\" style=\"" + buttonStyle + "\">" + c.button + "</a>\n" +
		"  </p>\n" +
		"  <p style=\"" + subtleStyle + "\">" + c.valid + "</p>\n" +
		"</body></html>"
	text := c.txtHead + "\n\n" + c.txtLead + "\n\n" + c.txtValid
	return &Template{
		Key:      TemplateKeySubscriptionActivated,
		Locale:   string(loc),
		Subject:  c.subj,
		BodyHTML: html,
		BodyText: text,
	}
}
