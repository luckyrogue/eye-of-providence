package reports

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eye-of-providence/backend/internal/metrics"
)

//go:embed prompts/system.md
var systemPrompt string

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

type GeminiClient struct {
	APIKey string
	Model  string
	HTTP   *http.Client
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &GeminiClient{
		APIKey: apiKey,
		Model:  model,
		HTTP:   &http.Client{Timeout: 60 * time.Second},
	}
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  map[string]any  `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (g *GeminiClient) Generate(ctx context.Context, nc *NumericContext) (string, error) {
	if g.APIKey == "" {
		metrics.ReportsGenerated.Inc()
		return mockReport(nc), nil
	}

	userText, err := buildUserPrompt(nc)
	if err != nil {
		return "", err
	}

	body := geminiRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: systemPrompt}}},
		Contents:          []geminiContent{{Role: "user", Parts: []geminiPart{{Text: userText}}}},
		GenerationConfig: map[string]any{
			"temperature":     0.3,
			"maxOutputTokens": 2048,
		},
	}
	raw, _ := json.Marshal(body)

	url := fmt.Sprintf(geminiEndpoint, g.Model, g.APIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var parsed geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("gemini error %d: %s", parsed.Error.Code, parsed.Error.Message)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini returned no content")
	}
	out := parsed.Candidates[0].Content.Parts[0].Text
	metrics.GeminiCallsTotal.Inc()
	metrics.ReportsGenerated.Inc()

	metrics.GeminiInputTokens.Add(uint64(len(userText) / 4))
	metrics.GeminiOutputTokens.Add(uint64(len(out) / 4))
	return out, nil
}

func buildUserPrompt(nc *NumericContext) (string, error) {
	stripped := struct {
		Period            string                       `json:"period"`
		FromISO           string                       `json:"from"`
		ToISO             string                       `json:"to"`
		TotalActiveSec    uint64                       `json:"total_active_sec"`
		IdleSec           uint64                       `json:"idle_sec"`
		ByCategorySec     map[string]uint64            `json:"by_category_sec"`
		AICharsByProvider map[string]uint64            `json:"ai_chars_by_provider"`
		AISecByChannel    map[string]uint64            `json:"ai_sec_by_channel"`
		ByLanguage        map[string]LanguageBreakdown `json:"by_language_chars"`
		EventsTotal       int                          `json:"events_total"`
	}{
		Period:            nc.Period,
		FromISO:           nc.From.Format(time.RFC3339),
		ToISO:             nc.To.Format(time.RFC3339),
		TotalActiveSec:    nc.TotalActiveMS / 1000,
		IdleSec:           nc.IdleMS / 1000,
		ByCategorySec:     scaleMap(nc.ByCategoryMS, 1000),
		AICharsByProvider: nc.AICharsByProvider,
		AISecByChannel:    scaleMap(nc.AIMSByChannel, 1000),
		ByLanguage:        nc.ByLanguageChars,
		EventsTotal:       nc.EventsTotal,
	}
	raw, err := json.MarshalIndent(stripped, "", "  ")
	if err != nil {
		return "", err
	}
	return "Сгенерируй weekly-отчёт по этим данным:\n\n```json\n" + string(raw) + "\n```\n", nil
}

func scaleMap(m map[string]uint64, divisor uint64) map[string]uint64 {
	out := make(map[string]uint64, len(m))
	for k, v := range m {
		out[k] = v / divisor
	}
	return out
}
