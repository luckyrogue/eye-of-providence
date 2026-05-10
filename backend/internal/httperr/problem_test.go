package httperr

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func TestBadRequest(t *testing.T) {
	app := fiber.New()
	app.Use(requestid.New())
	app.Get("/x", func(c *fiber.Ctx) error {
		return BadRequest(c, "bad_email", "email malformed")
	})
	res, _ := app.Test(httptest.NewRequest("GET", "/x", nil), -1)
	defer res.Body.Close()

	if res.StatusCode != 400 {
		t.Errorf("status=%d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Errorf("content-type=%q", ct)
	}

	body, _ := io.ReadAll(res.Body)
	var p ProblemDetails
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatal(err)
	}
	if p.Status != 400 {
		t.Errorf("status=%d", p.Status)
	}
	if p.Code != "bad_email" {
		t.Errorf("code=%q", p.Code)
	}
	if p.Detail != "email malformed" {
		t.Errorf("detail=%q", p.Detail)
	}
	if p.Error != "email malformed" {
		t.Errorf("error backward-compat alias broken: %q", p.Error)
	}
	if p.Title != "Bad Request" {
		t.Errorf("title=%q", p.Title)
	}
	if !strings.HasPrefix(p.Type, "https://eop.rysdavletov.org/errors/") {
		t.Errorf("type=%q", p.Type)
	}
	if p.Instance != "/x" {
		t.Errorf("instance=%q", p.Instance)
	}
	if p.RequestID == "" {
		t.Error("request_id should be set from middleware")
	}
}

func TestSeveralStatusHelpers(t *testing.T) {
	cases := []struct {
		fn   func(*fiber.Ctx, string, string) error
		code int
	}{
		{Unauthorized, 401},
		{Forbidden, 403},
		{NotFound, 404},
		{Conflict, 409},
		{Gone, 410},
		{TooLarge, 413},
		{TooManyRequests, 429},
		{Unavailable, 503},
		{BadGateway, 502},
	}
	for _, tc := range cases {
		app := fiber.New()
		fn := tc.fn
		app.Get("/x", func(c *fiber.Ctx) error { return fn(c, "test_code", "test detail") })
		res, _ := app.Test(httptest.NewRequest("GET", "/x", nil), -1)
		defer res.Body.Close()
		if res.StatusCode != tc.code {
			t.Errorf("expected %d, got %d", tc.code, res.StatusCode)
		}
	}
}

func TestInternal_GenericDetail(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error { return Internal(c) })
	res, _ := app.Test(httptest.NewRequest("GET", "/x", nil), -1)
	defer res.Body.Close()
	if res.StatusCode != 500 {
		t.Errorf("status=%d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), `"code":"internal_error"`) {
		t.Errorf("missing code: %s", body)
	}
	// Detail генерик, не leak'ает internal info
	if !strings.Contains(string(body), `"detail":"internal error"`) {
		t.Errorf("detail leaked: %s", body)
	}
}

func TestSend_AutoFillsDefaults(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		return Send(c, ProblemDetails{Status: 418, Code: "i_am_teapot"})
	})
	res, _ := app.Test(httptest.NewRequest("GET", "/x", nil), -1)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var p ProblemDetails
	_ = json.Unmarshal(body, &p)
	// Title autofill from net/http.StatusText
	if p.Title != "I'm a teapot" {
		t.Errorf("title=%q", p.Title)
	}
	// Type autofill from code
	if p.Type != "https://eop.rysdavletov.org/errors/i_am_teapot" {
		t.Errorf("type=%q", p.Type)
	}
}

func TestSend_PreservesExplicitFields(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		return Send(c, ProblemDetails{
			Status:    400,
			Code:      "custom",
			Type:      "https://example.com/custom",
			Title:     "Custom Title",
			Detail:    "specific detail",
			Instance:  "/override",
			RequestID: "explicit-rid",
		})
	})
	res, _ := app.Test(httptest.NewRequest("GET", "/x", nil), -1)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var p ProblemDetails
	_ = json.Unmarshal(body, &p)
	if p.Type != "https://example.com/custom" {
		t.Error("type override not preserved")
	}
	if p.Title != "Custom Title" {
		t.Error("title override not preserved")
	}
	if p.Instance != "/override" {
		t.Error("instance override not preserved")
	}
	if p.RequestID != "explicit-rid" {
		t.Error("request_id override not preserved")
	}
}
