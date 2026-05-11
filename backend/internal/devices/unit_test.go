package devices

import (
	"strings"
	"testing"
)

// Pure-Go тесты без БД — для быстрой обратной связи и CI без integration tag.

func TestRandomCode_AlphabetAndLength(t *testing.T) {
	for i := 0; i < 50; i++ {
		c, err := randomCode(codeLen)
		if err != nil {
			t.Fatalf("randomCode: %v", err)
		}
		if len(c) != codeLen {
			t.Errorf("len=%d, want %d", len(c), codeLen)
		}
		for _, r := range c {
			if !strings.ContainsRune(codeAlphabet, r) {
				t.Errorf("rune %q not in alphabet", r)
			}
		}
		// Гарантия отсутствия наиболее visually-confusable символов
		// (0/O, 1/I). L оставляем — pair-code UI рендерится моноширинно,
		// плюс L != 1 в большинстве шрифтов.
		if strings.ContainsAny(c, "0O1I") {
			t.Errorf("code %q contains confusable char", c)
		}
	}
}

func TestRandomToken_Format(t *testing.T) {
	tok, err := randomToken()
	if err != nil {
		t.Fatalf("randomToken: %v", err)
	}
	if !strings.HasPrefix(tok, tokenPrefix) {
		t.Errorf("token %q missing prefix %q", tok, tokenPrefix)
	}
	// "eop_" + 24 random bytes hex = 4 + 48 = 52 символа.
	if len(tok) != len(tokenPrefix)+tokenRandHex {
		t.Errorf("len=%d, want %d", len(tok), len(tokenPrefix)+tokenRandHex)
	}
}

func TestDefaultName(t *testing.T) {
	cases := map[string]string{
		"ext":     "Browser extension",
		"agent":   "Desktop agent",
		"ide":     "VS Code",
		"unknown": "Device",
	}
	for in, want := range cases {
		if got := defaultName(in); got != want {
			t.Errorf("defaultName(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestValidKinds(t *testing.T) {
	for _, k := range []string{"ext", "agent", "ide"} {
		if _, ok := validKinds[k]; !ok {
			t.Errorf("kind %q missing from validKinds", k)
		}
	}
	if _, ok := validKinds["windows"]; ok {
		t.Error("invalid kind 'windows' should not be in validKinds")
	}
}
