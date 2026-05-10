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

// PostError — wrap провайдерных HTTP error'ов. Status и raw body — для UI
// debugging (frontend показывает code + первые 200 chars).
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

// HTTPClient — interface для testability. http.Client satisfies it.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// PostGitHub — POST https://{host}/repos/{repo}/issues/{pr}/comments.
// Default host = api.github.com (для Cloud); enterprise users указывают свой
// (например api.github.example.com).
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

// PostGitLab — POST https://{host}/api/v4/projects/{repo-encoded}/merge_requests/{iid}/notes.
// Default host = https://gitlab.com; self-hosted users указывают свой URL.
// repo URL-encoded целиком ("namespace/project" → "namespace%2Fproject").
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
	// GitLab принимает либо "PRIVATE-TOKEN: xxx" header, либо "Authorization:
	// Bearer". Bearer работает и с PAT, и с OAuth — выбираем его.
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
