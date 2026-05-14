package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

var ErrUnknownSlug = errors.New("unknown content slug")

type SchemaError struct {
	Code       string
	Field      string
	Detail     string
	SchemaPath string
}

func (e *SchemaError) Error() string {
	if e == nil {
		return ""
	}
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field=%s)", e.Code, e.Detail, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Detail)
}

type ValidatorFunc func(raw json.RawMessage) *SchemaError

var hrefPattern = regexp.MustCompile(`^(https://|/)`)

var anchorPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validateString(m map[string]json.RawMessage, key string, maxLen int, fieldPath string) *SchemaError {
	const minLen = 1
	raw, present := m[key]
	if !present {
		return &SchemaError{
			Code:       "schema_violation",
			Field:      fieldPath + key,
			Detail:     "required field missing",
			SchemaPath: "/required/" + key,
		}
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return &SchemaError{
			Code:       "schema_violation",
			Field:      fieldPath + key,
			Detail:     "must be a string",
			SchemaPath: "/properties/" + key + "/type",
		}
	}
	if len(s) < minLen {
		return &SchemaError{
			Code:       "schema_violation",
			Field:      fieldPath + key,
			Detail:     fmt.Sprintf("must be at least %d characters", minLen),
			SchemaPath: "/properties/" + key + "/minLength",
		}
	}
	if len(s) > maxLen {
		return &SchemaError{
			Code:       "schema_violation",
			Field:      fieldPath + key,
			Detail:     fmt.Sprintf("must be at most %d characters", maxLen),
			SchemaPath: "/properties/" + key + "/maxLength",
		}
	}
	return nil
}

func validateTextOnly(maxLen int) ValidatorFunc {
	return func(raw json.RawMessage) *SchemaError {
		if len(raw) == 0 {
			return &SchemaError{Code: "invalid_json", Detail: "empty body"}
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			return &SchemaError{Code: "invalid_json", Detail: err.Error()}
		}
		if verr := validateString(m, "text", maxLen, "/"); verr != nil {
			return verr
		}
		for k := range m {
			if k != "text" {
				return &SchemaError{
					Code:       "schema_violation",
					Field:      "/" + k,
					Detail:     "unknown property; expected only `text`",
					SchemaPath: "/additionalProperties",
				}
			}
		}
		return nil
	}
}

func validateCTAButton() ValidatorFunc {
	return func(raw json.RawMessage) *SchemaError {
		if len(raw) == 0 {
			return &SchemaError{Code: "invalid_json", Detail: "empty body"}
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			return &SchemaError{Code: "invalid_json", Detail: err.Error()}
		}
		if verr := validateString(m, "label", 40, "/"); verr != nil {
			return verr
		}
		if verr := validateString(m, "href", 500, "/"); verr != nil {
			return verr
		}
		var href string
		_ = json.Unmarshal(m["href"], &href)
		if !hrefPattern.MatchString(href) {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/href",
				Detail:     "must start with https:// or / (relative path)",
				SchemaPath: "/properties/href/pattern",
			}
		}
		if raw, ok := m["external"]; ok {
			var b bool
			if err := json.Unmarshal(raw, &b); err != nil {
				return &SchemaError{
					Code:       "schema_violation",
					Field:      "/external",
					Detail:     "must be a boolean",
					SchemaPath: "/properties/external/type",
				}
			}
		}
		for k := range m {
			if k != "label" && k != "href" && k != "external" {
				return &SchemaError{
					Code:       "schema_violation",
					Field:      "/" + k,
					Detail:     "unknown property; expected label, href, external",
					SchemaPath: "/additionalProperties",
				}
			}
		}
		return nil
	}
}

func validatePricingTier() ValidatorFunc {
	return func(raw json.RawMessage) *SchemaError {
		if len(raw) == 0 {
			return &SchemaError{Code: "invalid_json", Detail: "empty body"}
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			return &SchemaError{Code: "invalid_json", Detail: err.Error()}
		}
		if verr := validateString(m, "tagline", 80, "/"); verr != nil {
			return verr
		}
		bulletsRaw, ok := m["bullets"]
		if !ok {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/bullets",
				Detail:     "required field missing",
				SchemaPath: "/required/bullets",
			}
		}
		var bullets []string
		if err := json.Unmarshal(bulletsRaw, &bullets); err != nil {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/bullets",
				Detail:     "must be an array of strings",
				SchemaPath: "/properties/bullets/type",
			}
		}
		if len(bullets) < 3 {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/bullets",
				Detail:     "must have at least 3 items",
				SchemaPath: "/properties/bullets/minItems",
			}
		}
		if len(bullets) > 8 {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/bullets",
				Detail:     "must have at most 8 items",
				SchemaPath: "/properties/bullets/maxItems",
			}
		}
		for i, b := range bullets {
			if len(b) < 1 {
				return &SchemaError{
					Code:       "schema_violation",
					Field:      fmt.Sprintf("/bullets[%d]", i),
					Detail:     "must be at least 1 character",
					SchemaPath: "/properties/bullets/items/minLength",
				}
			}
			if len(b) > 80 {
				return &SchemaError{
					Code:       "schema_violation",
					Field:      fmt.Sprintf("/bullets[%d]", i),
					Detail:     "must be at most 80 characters",
					SchemaPath: "/properties/bullets/items/maxLength",
				}
			}
		}
		for k := range m {
			if k != "tagline" && k != "bullets" {
				return &SchemaError{
					Code:       "schema_violation",
					Field:      "/" + k,
					Detail:     "unknown property; expected tagline, bullets",
					SchemaPath: "/additionalProperties",
				}
			}
		}
		return nil
	}
}

func validateFAQItems() ValidatorFunc {
	return func(raw json.RawMessage) *SchemaError {
		if len(raw) == 0 {
			return &SchemaError{Code: "invalid_json", Detail: "empty body"}
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			return &SchemaError{Code: "invalid_json", Detail: err.Error()}
		}
		itemsRaw, ok := m["items"]
		if !ok {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/items",
				Detail:     "required field missing",
				SchemaPath: "/required/items",
			}
		}
		var rawItems []map[string]json.RawMessage
		if err := json.Unmarshal(itemsRaw, &rawItems); err != nil {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/items",
				Detail:     "must be an array of objects",
				SchemaPath: "/properties/items/type",
			}
		}
		if len(rawItems) < 3 {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/items",
				Detail:     "must have at least 3 items",
				SchemaPath: "/properties/items/minItems",
			}
		}
		if len(rawItems) > 10 {
			return &SchemaError{
				Code:       "schema_violation",
				Field:      "/items",
				Detail:     "must have at most 10 items",
				SchemaPath: "/properties/items/maxItems",
			}
		}
		for i, item := range rawItems {
			prefix := fmt.Sprintf("/items[%d]/", i)
			if verr := validateString(item, "q", 200, prefix); verr != nil {
				return verr
			}
			if verr := validateString(item, "a", 800, prefix); verr != nil {
				return verr
			}
			if raw, ok := item["anchor"]; ok {
				var s string
				if err := json.Unmarshal(raw, &s); err != nil {
					return &SchemaError{
						Code:       "schema_violation",
						Field:      prefix + "anchor",
						Detail:     "must be a string",
						SchemaPath: "/properties/items/items/properties/anchor/type",
					}
				}
				if s != "" {
					if len(s) > 60 {
						return &SchemaError{
							Code:       "schema_violation",
							Field:      prefix + "anchor",
							Detail:     "must be at most 60 characters",
							SchemaPath: "/properties/items/items/properties/anchor/maxLength",
						}
					}
					if !anchorPattern.MatchString(s) {
						return &SchemaError{
							Code:       "schema_violation",
							Field:      prefix + "anchor",
							Detail:     "must be kebab-case (a-z, 0-9, -)",
							SchemaPath: "/properties/items/items/properties/anchor/pattern",
						}
					}
				}
			}
			for k := range item {
				if k != "q" && k != "a" && k != "anchor" {
					return &SchemaError{
						Code:       "schema_violation",
						Field:      prefix + k,
						Detail:     "unknown property; expected q, a, anchor",
						SchemaPath: "/properties/items/items/additionalProperties",
					}
				}
			}
		}
		for k := range m {
			if k != "items" {
				return &SchemaError{
					Code:       "schema_violation",
					Field:      "/" + k,
					Detail:     "unknown property; expected only `items`",
					SchemaPath: "/additionalProperties",
				}
			}
		}
		return nil
	}
}

type SlugDescriptor struct {
	Slug          string
	SchemaVersion int
	Validate      ValidatorFunc
}

var AllowedSlugs = []SlugDescriptor{
	{Slug: "landing.hero.headline", SchemaVersion: 1, Validate: validateTextOnly(200)},
	{Slug: "landing.hero.subhead", SchemaVersion: 1, Validate: validateTextOnly(400)},
	{Slug: "landing.hero.cta_primary", SchemaVersion: 1, Validate: validateCTAButton()},
	{Slug: "landing.hero.cta_secondary", SchemaVersion: 1, Validate: validateCTAButton()},
	{Slug: "landing.pricing.free.description", SchemaVersion: 1, Validate: validatePricingTier()},
	{Slug: "landing.pricing.pro.description", SchemaVersion: 1, Validate: validatePricingTier()},
	{Slug: "landing.pricing.business.description", SchemaVersion: 1, Validate: validatePricingTier()},
	{Slug: "landing.faq.items", SchemaVersion: 1, Validate: validateFAQItems()},
}

var slugIndex map[string]SlugDescriptor

func init() {
	slugIndex = make(map[string]SlugDescriptor, len(AllowedSlugs))
	for _, d := range AllowedSlugs {
		slugIndex[d.Slug] = d
	}
}

func LookupSlug(slug string) (SlugDescriptor, bool) {
	d, ok := slugIndex[slug]
	return d, ok
}

func IsKnownSlug(slug string) bool {
	_, ok := slugIndex[slug]
	return ok
}

func IsAllowedSlug(slug string) bool { return IsKnownSlug(slug) }

func SlugList() []string {
	out := make([]string, len(AllowedSlugs))
	for i, d := range AllowedSlugs {
		out[i] = d.Slug
	}
	return out
}

var SupportedLocales = []string{"en", "ru", "kk", "es"}

func IsSupportedLocale(loc string) bool {
	for _, l := range SupportedLocales {
		if l == loc {
			return true
		}
	}
	return false
}

func FallbackLocales(requested string) []string {
	if requested == "en" {
		return nil
	}
	return []string{"en"}
}

func Validate(slug string, raw json.RawMessage) error {
	desc, ok := LookupSlug(slug)
	if !ok {
		return ErrUnknownSlug
	}
	if verr := desc.Validate(raw); verr != nil {
		return verr
	}
	return nil
}
