package mailer

import (
	"fmt"
	"strings"
)

// Locale — поддерживаемые языки. Синхронизированы с
// dashboard/src/shared/i18n/index.ts SUPPORTED_LOCALES.
type Locale string

const (
	LocaleRU Locale = "ru"
	LocaleEN Locale = "en"
	LocaleKK Locale = "kk"
	LocaleES Locale = "es"
)

// NormalizeLocale — fallback на ru, если локаль пустая или неподдерживаемая.
func NormalizeLocale(s string) Locale {
	switch Locale(s) {
	case LocaleRU, LocaleEN, LocaleKK, LocaleES:
		return Locale(s)
	}
	return LocaleRU
}

// Все шаблоны inline, без template-engine: их пока 3, держим вместе чтобы
// видеть весь outbound surface в одном файле. Стиль — minimal HTML с
// inline CSS (gmail/outlook не любят <style>), text-fallback всегда есть.

const baseStyle = `font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;` +
	`max-width:560px;margin:0 auto;padding:24px;line-height:1.5;color:#0a0a0a;`
const buttonStyle = `display:inline-block;background:#0a0a0a;color:#fff;text-decoration:none;` +
	`padding:10px 16px;border-radius:6px;font-weight:600;font-size:14px;`
const subtleStyle = `color:#666;font-size:13px;margin-top:24px;border-top:1px solid #eee;padding-top:16px;`

// inviteCopy — все строки приглашения по локалям.
type inviteCopy struct {
	Subject       string // %s = team name
	Heading       string // %s = team name
	Greeting      string // %s%s — by-line + team name (e.g. "Иван приглашает в Acme")
	GreetingNoBy  string // %s — only team name when inviter unknown
	ButtonAccept  string
	IgnoreNote    string
	ValidNote     string
	TextHeader    string // %s = team name
	TextLine      string // %s%s = inviter + team
	TextLineNoBy  string // %s = team
	TextAccept    string
	TextIgnore    string
}

var inviteCopies = map[Locale]inviteCopy{
	LocaleRU: {
		Subject:      "Приглашение в команду %s · Eye of Providence",
		Heading:      "Тебя пригласили в %s",
		Greeting:     "<strong>%s</strong> приглашает тебя присоединиться к команде <strong>%s</strong> в Eye of Providence — инструменте для отслеживания AI-активности в коде.",
		GreetingNoBy: "Тебя приглашают присоединиться к команде <strong>%s</strong> в Eye of Providence — инструменте для отслеживания AI-активности в коде.",
		ButtonAccept: "Принять приглашение",
		IgnoreNote:   "Если ты не ожидал это письмо — просто игнорируй его, ничего не произойдёт.",
		ValidNote:    "Ссылка действительна 7 дней.",
		TextHeader:   "Тебя пригласили в команду %s.",
		TextLine:     "%s приглашает тебя в %s в Eye of Providence.",
		TextLineNoBy: "Тебя приглашают в %s в Eye of Providence.",
		TextAccept:   "Принять приглашение:",
		TextIgnore:   "Если ты не ожидал это письмо — игнорируй. Ссылка действительна 7 дней.",
	},
	LocaleEN: {
		Subject:      "Invite to %s · Eye of Providence",
		Heading:      "You've been invited to %s",
		Greeting:     "<strong>%s</strong> is inviting you to join <strong>%s</strong> on Eye of Providence — a tool for tracking AI activity in code.",
		GreetingNoBy: "You're invited to join <strong>%s</strong> on Eye of Providence — a tool for tracking AI activity in code.",
		ButtonAccept: "Accept invite",
		IgnoreNote:   "If you didn't expect this email — just ignore it, nothing will happen.",
		ValidNote:    "Link valid for 7 days.",
		TextHeader:   "You've been invited to %s.",
		TextLine:     "%s is inviting you to %s on Eye of Providence.",
		TextLineNoBy: "You're invited to %s on Eye of Providence.",
		TextAccept:   "Accept invite:",
		TextIgnore:   "If you didn't expect this email — ignore. Link valid for 7 days.",
	},
	LocaleKK: {
		Subject:      "%s командасына шақыру · Eye of Providence",
		Heading:      "Сізді %s командасына шақырды",
		Greeting:     "<strong>%s</strong> сізді <strong>%s</strong> командасына Eye of Providence-те қосылуға шақырып жатыр — кодтағы AI-белсенділікті қадағалау құралы.",
		GreetingNoBy: "Сіз <strong>%s</strong> командасына Eye of Providence-те қосылуға шақырылдыңыз — кодтағы AI-белсенділікті қадағалау құралы.",
		ButtonAccept: "Шақыруды қабылдау",
		IgnoreNote:   "Егер бұл хатты күтпеген болсаңыз — назар аудармаңыз, ештеңе болмайды.",
		ValidNote:    "Сілтеме 7 күнге жарамды.",
		TextHeader:   "Сізді %s командасына шақырды.",
		TextLine:     "%s сізді %s командасына Eye of Providence-те шақырып жатыр.",
		TextLineNoBy: "Сіз %s командасына Eye of Providence-те шақырылдыңыз.",
		TextAccept:   "Шақыруды қабылдау:",
		TextIgnore:   "Хатты күтпесеңіз — назар аудармаңыз. Сілтеме 7 күнге жарамды.",
	},
	LocaleES: {
		Subject:      "Invitación a %s · Eye of Providence",
		Heading:      "Te invitaron a %s",
		Greeting:     "<strong>%s</strong> te invita a unirte al equipo <strong>%s</strong> en Eye of Providence — una herramienta para rastrear actividad de IA en el código.",
		GreetingNoBy: "Te invitan a unirte al equipo <strong>%s</strong> en Eye of Providence — una herramienta para rastrear actividad de IA en el código.",
		ButtonAccept: "Aceptar invitación",
		IgnoreNote:   "Si no esperabas este email — ignóralo, no pasará nada.",
		ValidNote:    "Enlace válido por 7 días.",
		TextHeader:   "Te invitaron al equipo %s.",
		TextLine:     "%s te invita al equipo %s en Eye of Providence.",
		TextLineNoBy: "Te invitan al equipo %s en Eye of Providence.",
		TextAccept:   "Aceptar invitación:",
		TextIgnore:   "Si no esperabas este email — ignóralo. Enlace válido por 7 días.",
	},
}

// InviteEmail — приглашение в команду.
//
//	teamName    "Acme Inc."
//	inviteURL   "https://app.example.com/?invite=CODE"
//	inviterName "Темирлан" (опционально)
//	locale      "ru" | "en" | "kk" | "es" (пустой = ru)
func InviteEmail(teamName, inviteURL, inviterName string, locale Locale) (subject, html, text string) {
	c := inviteCopies[NormalizeLocale(string(locale))]

	subject = fmt.Sprintf(c.Subject, teamName)

	greeting := ""
	textGreeting := ""
	if inviterName != "" {
		greeting = fmt.Sprintf(c.Greeting, htmlEscape(inviterName), htmlEscape(teamName))
		textGreeting = fmt.Sprintf(c.TextLine, inviterName, teamName)
	} else {
		greeting = fmt.Sprintf(c.GreetingNoBy, htmlEscape(teamName))
		textGreeting = fmt.Sprintf(c.TextLineNoBy, teamName)
	}

	html = fmt.Sprintf(`<!doctype html>
<html><body style="%s">
  <h2 style="margin:0 0 12px;font-weight:700;">%s</h2>
  <p>%s</p>
  <p style="margin:24px 0;">
    <a href="%s" style="%s">%s</a>
  </p>
  <p style="%s">%s<br>%s</p>
</body></html>`,
		baseStyle,
		fmt.Sprintf(c.Heading, htmlEscape(teamName)),
		greeting,
		inviteURL,
		buttonStyle,
		c.ButtonAccept,
		subtleStyle,
		c.IgnoreNote,
		c.ValidNote,
	)

	text = fmt.Sprintf(`%s

%s

%s %s

%s`,
		fmt.Sprintf(c.TextHeader, teamName),
		textGreeting,
		c.TextAccept, inviteURL,
		c.TextIgnore,
	)

	return
}

// resetCopy — строки сброса пароля.
type resetCopy struct {
	Subject     string
	Heading     string
	Lead        string
	Button      string
	ValidNote   string
	TextHeader  string
	TextLead    string
	TextValid   string
}

var resetCopies = map[Locale]resetCopy{
	LocaleRU: {
		Subject:    "Сброс пароля · Eye of Providence",
		Heading:    "Сброс пароля",
		Lead:       "Кто-то (надеюсь, ты) запросил сброс пароля для твоего аккаунта в Eye of Providence.",
		Button:     "Сбросить пароль",
		ValidNote:  "Ссылка действительна 1 час. Если это был не ты — просто игнорируй; пароль не изменится без перехода по ссылке.",
		TextHeader: "Сброс пароля · Eye of Providence",
		TextLead:   "Если ты запрашивал сброс пароля, перейди по ссылке:",
		TextValid:  "Ссылка действительна 1 час. Если это был не ты — игнорируй.",
	},
	LocaleEN: {
		Subject:    "Password reset · Eye of Providence",
		Heading:    "Password reset",
		Lead:       "Someone (hopefully you) requested a password reset for your Eye of Providence account.",
		Button:     "Reset password",
		ValidNote:  "Link valid for 1 hour. If it wasn't you — just ignore; password won't change without clicking the link.",
		TextHeader: "Password reset · Eye of Providence",
		TextLead:   "If you requested a password reset, follow the link:",
		TextValid:  "Link valid for 1 hour. If it wasn't you — ignore.",
	},
	LocaleKK: {
		Subject:    "Құпиясөзді ауыстыру · Eye of Providence",
		Heading:    "Құпиясөзді ауыстыру",
		Lead:       "Біреу (өзіңіз болса жақсы) сіздің Eye of Providence аккаунтыңыздың құпиясөзін ауыстыруды сұрады.",
		Button:     "Құпиясөзді ауыстыру",
		ValidNote:  "Сілтеме 1 сағатқа жарамды. Егер бұл сіз болмасаңыз — назар аудармаңыз; сілтемеге өтпейінше құпиясөз өзгермейді.",
		TextHeader: "Құпиясөзді ауыстыру · Eye of Providence",
		TextLead:   "Егер құпиясөзді ауыстыруды сұраған болсаңыз, сілтемеге өтіңіз:",
		TextValid:  "Сілтеме 1 сағатқа жарамды. Сіз болмасаңыз — назар аудармаңыз.",
	},
	LocaleES: {
		Subject:    "Restablecer contraseña · Eye of Providence",
		Heading:    "Restablecer contraseña",
		Lead:       "Alguien (esperamos que tú) solicitó restablecer la contraseña de tu cuenta de Eye of Providence.",
		Button:     "Restablecer contraseña",
		ValidNote:  "Enlace válido por 1 hora. Si no fuiste tú — ignóralo; la contraseña no cambiará sin pasar por el enlace.",
		TextHeader: "Restablecer contraseña · Eye of Providence",
		TextLead:   "Si solicitaste restablecer la contraseña, abre el enlace:",
		TextValid:  "Enlace válido por 1 hora. Si no fuiste tú — ignóralo.",
	},
}

// PasswordResetEmail — ссылка для сброса пароля.
//
//	resetURL "https://app.example.com/reset?token=..."
//	locale   "ru" | "en" | "kk" | "es"
func PasswordResetEmail(resetURL string, locale Locale) (subject, html, text string) {
	c := resetCopies[NormalizeLocale(string(locale))]
	subject = c.Subject

	html = fmt.Sprintf(`<!doctype html>
<html><body style="%s">
  <h2 style="margin:0 0 12px;font-weight:700;">%s</h2>
  <p>%s</p>
  <p style="margin:24px 0;">
    <a href="%s" style="%s">%s</a>
  </p>
  <p style="%s">%s</p>
</body></html>`,
		baseStyle, c.Heading, c.Lead, resetURL, buttonStyle, c.Button, subtleStyle, c.ValidNote)

	text = fmt.Sprintf(`%s

%s
%s

%s`,
		c.TextHeader,
		c.TextLead,
		resetURL,
		c.TextValid,
	)

	return
}

// digestCopy — еженедельный отчёт.
type digestCopy struct {
	Subject     string
	Heading     string // %s = displayName
	Lead        string
	Button      string
	Unsubscribe string // %s = settings URL
	TextHello   string // %s = displayName
	TextLead    string
	TextOpen    string
	TextUnsub   string // %s = settings URL
}

var digestCopies = map[Locale]digestCopy{
	LocaleRU: {
		Subject:     "Еженедельный отчёт · Eye of Providence",
		Heading:     "Привет, %s.",
		Lead:        "Отчёт за прошлую неделю готов.",
		Button:      "Открыть дашборд",
		Unsubscribe: `Отписаться от еженедельных отчётов: <a href="%s/settings">в настройках</a>.`,
		TextHello:   "Привет, %s.",
		TextLead:    "Отчёт за прошлую неделю готов.",
		TextOpen:    "Открыть дашборд:",
		TextUnsub:   "Отписаться: %s/settings",
	},
	LocaleEN: {
		Subject:     "Weekly report · Eye of Providence",
		Heading:     "Hi %s.",
		Lead:        "Last week's report is ready.",
		Button:      "Open dashboard",
		Unsubscribe: `Unsubscribe from weekly reports: <a href="%s/settings">in settings</a>.`,
		TextHello:   "Hi %s.",
		TextLead:    "Last week's report is ready.",
		TextOpen:    "Open dashboard:",
		TextUnsub:   "Unsubscribe: %s/settings",
	},
	LocaleKK: {
		Subject:     "Апталық есеп · Eye of Providence",
		Heading:     "Сәлем, %s.",
		Lead:        "Өткен аптаның есебі дайын.",
		Button:      "Дашбордты ашу",
		Unsubscribe: `Апталық есептерден бас тарту: <a href="%s/settings">баптауларда</a>.`,
		TextHello:   "Сәлем, %s.",
		TextLead:    "Өткен аптаның есебі дайын.",
		TextOpen:    "Дашбордты ашу:",
		TextUnsub:   "Бас тарту: %s/settings",
	},
	LocaleES: {
		Subject:     "Reporte semanal · Eye of Providence",
		Heading:     "Hola, %s.",
		Lead:        "El reporte de la semana pasada está listo.",
		Button:      "Abrir panel",
		Unsubscribe: `Darse de baja de reportes semanales: <a href="%s/settings">en configuración</a>.`,
		TextHello:   "Hola, %s.",
		TextLead:    "El reporte de la semana pasada está listo.",
		TextOpen:    "Abrir panel:",
		TextUnsub:   "Darse de baja: %s/settings",
	},
}

// WeeklyDigestEmail — короткий push о готовом еженедельном AI-отчёте.
// Сам отчёт находится в дашборде; в письме только ссылка + summary.
func WeeklyDigestEmail(displayName, dashboardURL, summaryLine string, locale Locale) (subject, html, text string) {
	c := digestCopies[NormalizeLocale(string(locale))]
	subject = c.Subject

	html = fmt.Sprintf(`<!doctype html>
<html><body style="%s">
  <h2 style="margin:0 0 12px;font-weight:700;">%s</h2>
  <p>%s</p>
  <p style="background:#f5f5f5;padding:12px;border-radius:6px;font-family:monospace;font-size:13px;">%s</p>
  <p style="margin:24px 0;">
    <a href="%s" style="%s">%s</a>
  </p>
  <p style="%s">%s</p>
</body></html>`,
		baseStyle,
		fmt.Sprintf(c.Heading, htmlEscape(displayName)),
		c.Lead,
		htmlEscape(summaryLine),
		dashboardURL, buttonStyle, c.Button,
		subtleStyle,
		fmt.Sprintf(c.Unsubscribe, dashboardURL),
	)

	text = fmt.Sprintf(`%s

%s

%s

%s %s

%s`,
		fmt.Sprintf(c.TextHello, displayName),
		c.TextLead,
		summaryLine,
		c.TextOpen, dashboardURL,
		fmt.Sprintf(c.TextUnsub, dashboardURL),
	)

	return
}

// htmlEscape — точечный escape для значений из user-input в HTML body.
// Используется только для display_name (нет полного template engine).
func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}
