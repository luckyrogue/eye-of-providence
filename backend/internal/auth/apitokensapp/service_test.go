package apitokensapp_test

import (
	"context"
	"testing"

	"github.com/eye-of-providence/backend/internal/auth/apitokensapp"
)

type fakeStore struct{}

func (fakeStore) List(context.Context, string) ([]apitokensapp.TokenRow, error) {
	return []apitokensapp.TokenRow{{Prefix: "eop_abc"}}, nil
}

func TestList(t *testing.T) {
	out, err := apitokensapp.New(fakeStore{}).List(context.Background(), "u")
	if err != nil || len(out) != 1 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
