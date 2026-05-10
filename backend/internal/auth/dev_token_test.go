package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

// Smoke-test: с EnableDevToken=false роут /v1/auth/dev-token не должен
// существовать (404). Критично для прода — иначе любой получит JWT.
func TestRegisterRoutes_DevTokenDisabledIsNotRouted(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, Service{
		JWTSecret:      "test-secret-of-sufficient-length-32",
		Logger:         zap.NewNop(),
		EnableDevToken: false,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/dev-token", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 with dev-token disabled, got %d", resp.StatusCode)
	}
}

// Зеркальная проверка: при EnableDevToken=true роут зарегистрирован.
// Хэндлер требует Pool/Users; в этом тесте ставим recover-middleware и
// проверяем только что роут найден (не 404). 500 после nil-pool — норма.
func TestRegisterRoutes_DevTokenEnabledIsRouted(t *testing.T) {
	app := fiber.New()
	app.Use(recover.New())
	RegisterRoutes(app, Service{
		JWTSecret:      "test-secret-of-sufficient-length-32",
		Logger:         zap.NewNop(),
		EnableDevToken: true,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/dev-token", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		t.Fatalf("expected dev-token route to exist when enabled, got 404")
	}
}
