// Package httperr — typed error responses (RFC 7807 Problem Details + extensions).
//
// Goal: stable machine-readable error contract для frontend и API consumers.
// Заменяет ad-hoc `c.Status(400).JSON(fiber.Map{"error": "..."})`.
//
// Schema:
//
//	{
//	  "type":       "https://eop.rysdavletov.org/errors/{code}",  // dereferenceable URI
//	  "title":      "Bad Request",                                 // human-readable HTTP status
//	  "status":     400,                                           // HTTP status int
//	  "detail":     "name must be 1..100 chars",                   // specific message
//	  "code":       "invalid_team_name",                           // machine-readable identifier
//	  "instance":   "/v1/teams",                                   // path that produced error
//	  "request_id": "abc-123",                                     // correlation w/ logs
//	  "error":      "name must be 1..100 chars"                    // backward-compat alias of detail
//	}
//
// Backward-compat: keep `error` field = `detail` чтобы старые frontend
// consumers продолжали работать без миграции. Новый код может читать
// `code` для structured error-handling (e.g. показать конкретный i18n).
//
// Content-Type: application/problem+json (per RFC 7807). Fiber клиенты
// которые проверяют content-type — корректно расценивают как error.
package httperr

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// ProblemDetails — JSON-serializable struct. Fields per RFC 7807 + EoP-specific
// extensions (code, request_id, error).
type ProblemDetails struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	// Error — alias of Detail для backward-compat со старыми consumers
	// которые читали `error` поле. Новый код предпочитает Detail+Code.
	Error string `json:"error,omitempty"`
}

const (
	contentType = "application/problem+json"
	typePrefix  = "https://eop.rysdavletov.org/errors/"
)

// Send — записывает problem в response с правильным content-type. Возвращает
// nil чтобы handler могли просто `return httperr.Send(c, p)`.
func Send(c *fiber.Ctx, p ProblemDetails) error {
	if p.Title == "" {
		p.Title = http.StatusText(p.Status)
	}
	if p.Type == "" && p.Code != "" {
		p.Type = typePrefix + p.Code
	}
	if p.Instance == "" {
		p.Instance = c.Path()
	}
	if p.RequestID == "" {
		if rid, ok := c.Locals("requestid").(string); ok {
			p.RequestID = rid
		}
	}
	// Backward-compat: error = detail.
	if p.Error == "" {
		p.Error = p.Detail
	}
	// Используем Send + manual marshal вместо c.JSON чтобы set content-type
	// до отправки тела (c.JSON forc'ит application/json).
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	c.Status(p.Status)
	c.Set(fiber.HeaderContentType, contentType)
	return c.Send(body)
}

// --- Standard helpers ---

func BadRequest(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusBadRequest, Code: code, Detail: detail})
}

func Unauthorized(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusUnauthorized, Code: code, Detail: detail})
}

func Forbidden(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusForbidden, Code: code, Detail: detail})
}

func NotFound(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusNotFound, Code: code, Detail: detail})
}

func Conflict(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusConflict, Code: code, Detail: detail})
}

func Gone(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusGone, Code: code, Detail: detail})
}

func TooLarge(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusRequestEntityTooLarge, Code: code, Detail: detail})
}

func TooManyRequests(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusTooManyRequests, Code: code, Detail: detail})
}

func Unavailable(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusServiceUnavailable, Code: code, Detail: detail})
}

func BadGateway(c *fiber.Ctx, code, detail string) error {
	return Send(c, ProblemDetails{Status: fiber.StatusBadGateway, Code: code, Detail: detail})
}

// Internal — 500 с auto request_id correlation. Detail генерик ("internal error")
// чтобы не светить implementation details наружу. Caller должен log.Error
// собственно ошибку до этого вызова.
func Internal(c *fiber.Ctx) error {
	return Send(c, ProblemDetails{
		Status: fiber.StatusInternalServerError,
		Code:   "internal_error",
		Detail: "internal error",
	})
}
