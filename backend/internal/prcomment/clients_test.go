package prcomment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostGitHub_Path(t *testing.T) {
	var (
		gotPath   string
		gotAuth   string
		gotBody   string
		gotMethod string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(201)
	}))
	defer srv.Close()

	err := PostGitHub(context.Background(), srv.Client(),
		srv.URL, "luckyrogue/eye-of-providence", 42, "ghp_test", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" {
		t.Errorf("method=%q", gotMethod)
	}
	if gotPath != "/repos/luckyrogue/eye-of-providence/issues/42/comments" {
		t.Errorf("path=%q", gotPath)
	}
	if gotAuth != "Bearer ghp_test" {
		t.Errorf("auth=%q", gotAuth)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["body"] != "hello" {
		t.Errorf("body field=%q", payload["body"])
	}
}

func TestPostGitLab_PathEncoding(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(201)
	}))
	defer srv.Close()

	err := PostGitLab(context.Background(), srv.Client(),
		srv.URL, "team/sub-group/project", 7, "glpat-test", "hi")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(gotPath, "team%2Fsub-group%2Fproject") {
		t.Errorf("path=%q (no URL-encode of slashes)", gotPath)
	}
	if !strings.HasSuffix(gotPath, "/merge_requests/7/notes") {
		t.Errorf("path=%q", gotPath)
	}
}

func TestPost_4xxError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	defer srv.Close()

	err := PostGitHub(context.Background(), srv.Client(),
		srv.URL, "x/y", 1, "bad", "hi")
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*PostError)
	if !ok {
		t.Fatalf("got %T, want *PostError", err)
	}
	if pe.Status != 403 {
		t.Errorf("status=%d", pe.Status)
	}
	if !strings.Contains(pe.Body, "Bad credentials") {
		t.Errorf("body=%q", pe.Body)
	}
}

func TestPost_DefaultHosts(t *testing.T) {

	type roundTripFn func(*http.Request) (*http.Response, error)
	rt := roundTripFn(func(r *http.Request) (*http.Response, error) {

		got := r.URL.String()
		if !strings.HasPrefix(got, "https://api.github.com/") &&
			!strings.HasPrefix(got, "https://gitlab.com/") {
			return nil, &PostError{Status: 999, Body: "unexpected host: " + got}
		}
		return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	hc := &http.Client{Transport: roundTripperFn(rt)}

	if err := PostGitHub(context.Background(), hc, "", "x/y", 1, "t", "b"); err != nil {
		t.Errorf("github default host: %v", err)
	}
	if err := PostGitLab(context.Background(), hc, "", "x/y", 1, "t", "b"); err != nil {
		t.Errorf("gitlab default host: %v", err)
	}
}

type roundTripperFn func(*http.Request) (*http.Response, error)

func (f roundTripperFn) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestPostError_Truncates(t *testing.T) {
	long := strings.Repeat("X", 500)
	pe := &PostError{Status: 500, Body: long}
	if !strings.Contains(pe.Error(), "…") {
		t.Errorf("expected ellipsis truncation: %q", pe.Error())
	}
	if len(pe.Error()) > 250 {
		t.Errorf("error string too long: %d", len(pe.Error()))
	}
}
