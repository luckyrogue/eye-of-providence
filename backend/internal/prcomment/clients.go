package prcomment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PostError struct {
	Status int
	Body   string
}

func (e *PostError) Error() string {
	short := e.Body
	if len(short) > 200 {
		short = short[:200] + "…"
	}
	return fmt.Sprintf("provider %d: %s", e.Status, short)
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func PostGitHub(ctx context.Context, hc HTTPClient, host, repo string, prNumber int, token, body string) error {
	if host == "" {
		host = "https://api.github.com"
	}
	host = strings.TrimRight(host, "/")
	endpoint := fmt.Sprintf("%s/repos/%s/issues/%d/comments", host, repo, prNumber)
	payload, _ := json.Marshal(map[string]string{"body": body})
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")
	return doPost(hc, req)
}

func PostGitLab(ctx context.Context, hc HTTPClient, host, repo string, mrIID int, token, body string) error {
	if host == "" {
		host = "https://gitlab.com"
	}
	host = strings.TrimRight(host, "/")
	endpoint := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/notes",
		host, url.PathEscape(repo), mrIID)
	payload, _ := json.Marshal(map[string]string{"body": body})
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return doPost(hc, req)
}

func doPost(hc HTTPClient, req *http.Request) error {
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	res, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return &PostError{Status: res.StatusCode, Body: string(body)}
	}
	return nil
}
