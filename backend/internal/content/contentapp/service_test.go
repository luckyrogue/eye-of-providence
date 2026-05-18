package contentapp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/content/domain"
)

type fakeRepo struct {
	blocks map[string]*domain.Block
}

func (f *fakeRepo) key(slug, locale string) string { return slug + ":" + locale }

func (f *fakeRepo) Lookup(_ context.Context, slug, locale string, includeDraft bool) (*domain.Block, error) {
	b, ok := f.blocks[f.key(slug, locale)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if !includeDraft {
		out := *b
		out.DraftContent = nil
		return &out, nil
	}
	cp := *b
	return &cp, nil
}

func (f *fakeRepo) Upsert(context.Context, domain.UpsertParams) (*domain.Block, error) {
	return nil, nil
}
func (f *fakeRepo) Delete(context.Context, string, string) (*domain.Block, error) {
	return nil, nil
}
func (f *fakeRepo) ListMatrix(context.Context) ([]domain.MatrixEntry, error) {
	return nil, nil
}

type fakeCat struct{}

func (fakeCat) IsKnownSlug(slug string) bool { return slug == "landing.hero.headline" }
func (fakeCat) IsSupportedLocale(loc string) bool {
	return loc == "en" || loc == "ru"
}
func (fakeCat) LookupSlug(slug string) (domain.SlugDescriptor, bool) {
	return domain.LookupSlug(slug)
}
func (fakeCat) FallbackLocales(string) []string { return []string{"en"} }
func (fakeCat) AllowedSlugs() []domain.SlugDescriptor {
	return []domain.SlugDescriptor{{Slug: "landing.hero.headline", SchemaVersion: 1}}
}
func (fakeCat) SupportedLocales() []string { return []string{"en", "ru"} }
func (fakeCat) ValidateContent(slug string, raw []byte) error {
	return domain.Validate(slug, raw)
}

func TestGetPublished_localeFallback(t *testing.T) {
	now := time.Now()
	pub := now.Add(-time.Hour)
	repo := &fakeRepo{blocks: map[string]*domain.Block{
		"landing.hero.headline:en": {
			Slug: "landing.hero.headline", Locale: "en",
			Content: json.RawMessage(`{"text":"hello"}`), SchemaVersion: 1,
			PublishedAt: &pub, UpdatedAt: now,
		},
	}}
	svc := New(Deps{Repo: repo, Cat: fakeCat{}})
	view, err := svc.GetPublished(context.Background(), "landing.hero.headline", "ru", "")
	if err != nil {
		t.Fatal(err)
	}
	if view.Locale != "en" {
		t.Fatalf("locale fallback: got %q want en", view.Locale)
	}
	if view.Source != "locale_fallback:en" {
		t.Fatalf("source: %q", view.Source)
	}
}

func TestGetPublished_notFound(t *testing.T) {
	svc := New(Deps{Repo: &fakeRepo{blocks: map[string]*domain.Block{}}, Cat: fakeCat{}})
	_, err := svc.GetPublished(context.Background(), "landing.hero.headline", "ru", "")
	if err != domain.ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestCheckSuperAdmin(t *testing.T) {
	uid := uuid.New()
	svc := New(Deps{Admins: superAdminGate{uid: uid}})
	if !svc.CheckSuperAdmin(context.Background(), uid) {
		t.Fatal("expected super admin")
	}
}

type superAdminGate struct{ uid uuid.UUID }

func (g superAdminGate) IsSuperAdmin(_ context.Context, id uuid.UUID) bool { return id == g.uid }
