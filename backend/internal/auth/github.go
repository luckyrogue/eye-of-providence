package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GitHubOAuth struct {
	cfg *oauth2.Config
}

func NewGitHubOAuth(clientID, clientSecret, redirectURL string) *GitHubOAuth {
	return &GitHubOAuth{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		},
	}
}

func (g *GitHubOAuth) AuthCodeURL(state string) string {
	return g.cfg.AuthCodeURL(state)
}

func (g *GitHubOAuth) Exchange(ctx context.Context, code string) (*GitHubUser, error) {
	if g.cfg.ClientID == "" {
		return nil, errors.New("github oauth not configured (set EOP_GITHUB_CLIENT_ID)")
	}
	tok, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	client := g.cfg.Client(ctx, tok)
	client.Timeout = 10 * time.Second

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("github user fetch failed")
	}
	var u GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}

	// Заменяем email на primary verified из /user/emails. Поле user.email из /user
	// может быть public-email профиля и не верифицированным — доверять ему нельзя
	// (security: иначе чужак с GitHub-аккаунтом, где email = чужой_email, мог бы
	// присвоить чужую идентичность в нашей системе).
	emailReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	emailResp, err := client.Do(emailReq)
	if err != nil {
		return nil, err
	}
	defer emailResp.Body.Close()
	if emailResp.StatusCode != http.StatusOK {
		return nil, errors.New("github user/emails fetch failed (need user:email scope)")
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
		return nil, err
	}
	verified := ""
	for _, e := range emails {
		if e.Primary && e.Verified {
			verified = e.Email
			break
		}
	}
	if verified == "" {
		// Fallback: любой verified non-primary, если primary не verified.
		for _, e := range emails {
			if e.Verified {
				verified = e.Email
				break
			}
		}
	}
	if verified == "" {
		return nil, errors.New("no verified email on github account")
	}
	u.Email = verified
	return &u, nil
}
