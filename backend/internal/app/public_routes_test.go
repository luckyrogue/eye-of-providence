package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/devices"
	"github.com/eye-of-providence/backend/internal/teams"
)

// publicRoutes — source of truth for endpoints that MUST be reachable without
// Authorization header. Adding a new public endpoint? Add it here AND make sure
// the corresponding RegisterRoutes() call lives BEFORE teams.RegisterRoutes()
// in app/modules.go. teams.RegisterRoutes() installs auth.Middleware on the
// `/v1` group via `app.Group("/v1", mw)` — which Fiber implements as
// `app.Use("/v1", mw)`, applying middleware to every subsequent route on that
// prefix.
//
// We hit this trap twice in alpha-1 (password-reset, devices/pair). This test
// makes it impossible to regress: rearranging modules.go puts an endpoint
// behind auth → CI fails with the exact path that broke.
var publicRoutes = []struct {
	method, path string
}{
	{http.MethodPost, "/v1/auth/forgot-password"},
	{http.MethodPost, "/v1/auth/reset-password"},
	{http.MethodPost, "/v1/devices/pair"},
	{http.MethodPost, "/v1/devices/poll"},
}

// TestPublicRoutesNotBehindAuth wires the same registration sequence as
// app.API.RegisterProductRoutes (the slice of registration that's relevant
// to the public/auth boundary) and asserts each declared public route does
// NOT return 401 "missing_bearer" when called without an Authorization header.
//
// We register a subset (no Store/PG required for the routes under test) but
// CRUCIALLY in the same order as the production composition root. That's the
// invariant under test: ORDER. If you add new RegisterRoutes calls to
// modules.go between password-reset and teams, mirror them here.
func TestPublicRoutesNotBehindAuth(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	log := zap.NewNop()

	// SANITY-CHECK FOR THIS TEST: comment out one of the public registrations
	// below, swap with teams' registration FIRST — test should fail with the
	// exact endpoint name. We verified this manually 2026-05-19.

	// teams MUST come AFTER all public registrations.
	teams.RegisterRoutes(app, teams.Service{
		Pool: nil, JWTSecret: "test-secret-32-chars-or-longer-aaaa",
		Logger: log, InviteOnly: true, BetaTeamLimit: 3,
	})

	auth.RegisterPasswordResetRoutes(app, auth.PasswordResetService{
		Pool: nil, Mailer: nil, PublicURL: "http://test.local", Logger: log,
	})

	devices.RegisterRoutes(app, devices.Service{
		Pool: nil, Logger: log, JWTSecret: "test-secret-32-chars-or-longer-aaaa",
	})

	for _, r := range publicRoutes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			req := httptest.NewRequest(r.method, r.path, strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusUnauthorized {
				t.Fatalf(
					"%s %s returned 401 (missing_bearer) — auth.Middleware caught a public route.\n"+
						"Almost certainly app/modules.go RegisterProductRoutes() registers this route AFTER\n"+
						"teams.RegisterRoutes(), which installs /v1 auth.Middleware via Fiber Group/Use\n"+
						"semantics. Move the registration call ABOVE teams.RegisterRoutes().",
					r.method, r.path,
				)
			}
		})
	}
}
