package httperr

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type ProblemDetails struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`

	Error string `json:"error,omitempty"`

	Extensions map[string]any `json:"-"`
}

func (p ProblemDetails) MarshalJSON() ([]byte, error) {
	type alias ProblemDetails
	base, err := json.Marshal(alias(p))
	if err != nil {
		return nil, err
	}
	if len(p.Extensions) == 0 {
		return base, nil
	}

	var m map[string]any
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}
	for k, v := range p.Extensions {
		m[k] = v
	}
	return json.Marshal(m)
}

const (
	contentType = "application/problem+json"
	typePrefix  = "https://eop.rysdavletov.org/errors/"
)

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

	if p.Error == "" {
		p.Error = p.Detail
	}

	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	c.Status(p.Status)
	c.Set(fiber.HeaderContentType, contentType)
	return c.Send(body)
}

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

func Internal(c *fiber.Ctx) error {
	return Send(c, ProblemDetails{
		Status: fiber.StatusInternalServerError,
		Code:   "internal_error",
		Detail: "internal error",
	})
}
