package auth

import "context"

// ExternalUser — нормализованный набор полей, который любой OAuth-провайдер
// возвращает после успешного `code → token → userinfo` обмена. Поля subject
// (provider-specific user ID) и email обязательны; name/login опциональны.
//
// Login заполняется только для GitHub (handle, отображается в UI).
type ExternalUser struct {
	Subject string // provider-specific user id (github numeric, google sub, apple sub)
	Email   string // verified email — провайдер обязан гарантировать верификацию
	Name    string // display name (опционально)
	Login   string // github username; для google/apple пустой
}

// OAuthProvider — общий интерфейс для всех OAuth-провайдеров (github/google/apple).
// Позволяет handler.go обрабатывать flow единообразно: redirect → callback → upsert
// identity → JWT. Конкретные имплементации различаются только AuthCodeURL/Exchange.
type OAuthProvider interface {
	// Name — provider key, например "github" | "google" | "apple". Используется
	// в URL'е (`/v1/auth/:name/login`) и как значение колонки `provider` в
	// таблице user_identities.
	Name() string

	// AuthCodeURL — URL для редиректа пользователя на consent screen IdP.
	// state — anti-CSRF nonce, должен совпасть в callback.
	AuthCodeURL(state string) string

	// Exchange — обмен authorization code на ExternalUser. Внутри:
	//   1. token exchange (POST на token endpoint провайдера)
	//   2. user info / id_token verification
	//   3. проверка email_verified (где applicable)
	//
	// Возвращает ошибку если provider не сконфигурирован, токен невалиден,
	// или email не верифицирован.
	Exchange(ctx context.Context, code string) (*ExternalUser, error)
}
