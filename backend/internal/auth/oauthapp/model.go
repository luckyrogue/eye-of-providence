package oauthapp

// ExternalUser — нормализованный профиль после OAuth code exchange.
type ExternalUser struct {
	Subject string
	Email   string
	Name    string
	Login   string
}
