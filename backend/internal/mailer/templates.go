package mailer

import (
	"fmt"
	"strings"
)

// Все шаблоны inline, без template-engine: их пока 3, держим вместе чтобы
// видеть весь outbound surface в одном файле. Стиль — minimal HTML с
// inline CSS (gmail/outlook не любят <style>), text-fallback всегда есть.

const baseStyle = `font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;` +
	`max-width:560px;margin:0 auto;padding:24px;line-height:1.5;color:#0a0a0a;`
const buttonStyle = `display:inline-block;background:#0a0a0a;color:#fff;text-decoration:none;` +
	`padding:10px 16px;border-radius:6px;font-weight:600;font-size:14px;`
const subtleStyle = `color:#666;font-size:13px;margin-top:24px;border-top:1px solid #eee;padding-top:16px;`

// InviteEmail — приглашение в команду.
//
//	teamName    "Acme Inc."
//	inviteURL   "https://app.example.com/?invite=CODE"
//	inviterName "Темирлан" (опционально)
func InviteEmail(teamName, inviteURL, inviterName string) (subject, html, text string) {
	subject = fmt.Sprintf("Приглашение в команду %s · Eye of Providence", teamName)

	by := ""
	if inviterName != "" {
		by = fmt.Sprintf("<strong>%s</strong> ", inviterName)
	}

	html = fmt.Sprintf(`<!doctype html>
<html><body style="%s">
  <h2 style="margin:0 0 12px;font-weight:700;">Тебя пригласили в %s</h2>
  <p>%sприглашает тебя присоединиться к команде <strong>%s</strong> в Eye of Providence — инструменте для отслеживания AI-активности в коде.</p>
  <p style="margin:24px 0;">
    <a href="%s" style="%s">Принять приглашение</a>
  </p>
  <p style="%s">Если ты не ожидал это письмо — просто игнорируй его, ничего не произойдёт.<br>
  Ссылка действительна 7 дней.</p>
</body></html>`,
		baseStyle, teamName, by, teamName, inviteURL, buttonStyle, subtleStyle)

	text = fmt.Sprintf(`Тебя пригласили в команду %s.

%sприглашает тебя присоединиться в Eye of Providence.

Принять приглашение: %s

Если ты не ожидал это письмо — игнорируй. Ссылка действительна 7 дней.`,
		teamName, strings.TrimSpace(plainText(by)), inviteURL)

	return
}

// PasswordResetEmail — ссылка для сброса пароля.
//
//	resetURL "https://app.example.com/reset?token=..."
func PasswordResetEmail(resetURL string) (subject, html, text string) {
	subject = "Сброс пароля · Eye of Providence"

	html = fmt.Sprintf(`<!doctype html>
<html><body style="%s">
  <h2 style="margin:0 0 12px;font-weight:700;">Сброс пароля</h2>
  <p>Кто-то (надеюсь, ты) запросил сброс пароля для твоего аккаунта в Eye of Providence.</p>
  <p style="margin:24px 0;">
    <a href="%s" style="%s">Сбросить пароль</a>
  </p>
  <p style="%s">Ссылка действительна 1 час. Если это был не ты — просто игнорируй; пароль не изменится без перехода по ссылке.</p>
</body></html>`,
		baseStyle, resetURL, buttonStyle, subtleStyle)

	text = fmt.Sprintf(`Сброс пароля · Eye of Providence

Если ты запрашивал сброс пароля, перейди по ссылке:
%s

Ссылка действительна 1 час. Если это был не ты — игнорируй.`, resetURL)

	return
}

// WeeklyDigestEmail — короткий push о готовом еженедельном AI-отчёте.
// Сам отчёт находится в дашборде; в письме только ссылка + summary.
//
//	displayName "Темирлан"
//	dashboardURL "https://app.example.com/dashboard"
//	summaryLine "На прошлой неделе: 42% AI, 28% manual, 30% review"
func WeeklyDigestEmail(displayName, dashboardURL, summaryLine string) (subject, html, text string) {
	subject = "Еженедельный отчёт · Eye of Providence"

	html = fmt.Sprintf(`<!doctype html>
<html><body style="%s">
  <h2 style="margin:0 0 12px;font-weight:700;">Привет, %s.</h2>
  <p>Отчёт за прошлую неделю готов.</p>
  <p style="background:#f5f5f5;padding:12px;border-radius:6px;font-family:monospace;font-size:13px;">%s</p>
  <p style="margin:24px 0;">
    <a href="%s" style="%s">Открыть дашборд</a>
  </p>
  <p style="%s">Отписаться от еженедельных отчётов: <a href="%s/settings">в настройках</a>.</p>
</body></html>`,
		baseStyle, htmlEscape(displayName), htmlEscape(summaryLine), dashboardURL, buttonStyle, subtleStyle, dashboardURL)

	text = fmt.Sprintf(`Привет, %s.

Отчёт за прошлую неделю готов.

%s

Открыть дашборд: %s

Отписаться: %s/settings`, displayName, summaryLine, dashboardURL, dashboardURL)

	return
}

// plainText — для text-fallback'а; убираем самые частые HTML-теги.
// Для аккуратности достаточно: subject/html — мы сами генерим, не user-content.
func plainText(s string) string {
	r := strings.NewReplacer("<strong>", "", "</strong>", "", "<br>", "\n")
	return r.Replace(s)
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
