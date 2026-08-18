package templates

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
)

func TestMain(m *testing.M) {
	os.Setenv("APP_BASE_URL", "https://dviji.se")
	os.Setenv("JWT_KEY", "test-jwt-key")
	os.Setenv("JWT_REFRESH_KEY", "test-jwt-refresh-key")
	os.Setenv("XSRF", "test-xsrf-key")

	os.Exit(m.Run())
}

func TestSEO_PageTitle(t *testing.T) {
	if got := (SEO{Title: "Блог"}).PageTitle(); got != "Блог - Движи се" {
		t.Errorf("PageTitle() = %q", got)
	}

	if got := (SEO{}).PageTitle(); got != "Движи се" {
		t.Errorf("PageTitle() with no title = %q, want the bare site name", got)
	}
}

func TestSEO_CanonicalURL(t *testing.T) {
	if got := (SEO{Path: "/blog/my-post"}).CanonicalURL(); got != "https://dviji.se/blog/my-post" {
		t.Errorf("CanonicalURL() = %q", got)
	}

	// No path means no canonical, rather than one pointing at the site root.
	if got := (SEO{}).CanonicalURL(); got != "" {
		t.Errorf("CanonicalURL() with no path = %q, want empty", got)
	}
}

// Social scrapers do not resolve relative image paths, so the share image must
// always come out absolute.
func TestSEO_ImageAbsoluteURL(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{"static path is made absolute", "/static/img/a.png", "https://dviji.se/static/img/a.png"},
		{"absolute urls are left alone", "https://res.cloudinary.com/x/a.jpg", "https://res.cloudinary.com/x/a.jpg"},
		{"empty falls back to the logo", "", "https://dviji.se/static/img/logo_1024x1024.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (SEO{ImageURL: tt.image}).ImageAbsoluteURL(); got != tt.want {
				t.Errorf("ImageAbsoluteURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSEO_ArticleDetection(t *testing.T) {
	published := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	article := SEO{PublishedAt: &published}
	if !article.IsArticle() || article.OGType() != "article" {
		t.Error("a page with a publish date should be an article")
	}

	page := SEO{}
	if page.IsArticle() || page.OGType() != "website" {
		t.Error("a page without a publish date should be a website")
	}
}

// dateModified must always be present for Schema.org, so it falls back to the
// publish date when a post has never been edited.
func TestSEO_ModifiedFallsBackToPublished(t *testing.T) {
	published := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	seo := SEO{PublishedAt: &published}

	if got := seo.ModifiedISO(); got != seo.PublishedISO() {
		t.Errorf("ModifiedISO() = %q, want the published time %q", got, seo.PublishedISO())
	}
}

func TestSEO_StructuredDataForArticle(t *testing.T) {
	published := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	modified := time.Date(2026, 3, 2, 9, 30, 0, 0, time.UTC)

	seo := SEO{
		Title:       "Фитнес за начинаещи",
		Description: "Как да започнете",
		Path:        "/blog/fitnes",
		ImageURL:    "https://res.cloudinary.com/x/cover.jpg",
		PublishedAt: &published,
		ModifiedAt:  &modified,
		AuthorName:  "Иван Петров",
		Section:     "Тренировки",
	}

	data := seo.StructuredData()

	checks := map[string]string{
		"@type":          "Article",
		"headline":       "Фитнес за начинаещи",
		"datePublished":  "2026-03-01T12:00:00Z",
		"dateModified":   "2026-03-02T09:30:00Z",
		"url":            "https://dviji.se/blog/fitnes",
		"articleSection": "Тренировки",
	}

	for key, want := range checks {
		if got, _ := data[key].(string); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}

	author, _ := data["author"].(map[string]any)
	if name, _ := author["name"].(string); name != "Иван Петров" {
		t.Errorf("author name = %q", name)
	}
}

func TestSEO_StructuredDataForNonArticle(t *testing.T) {
	data := SEO{Title: "Начало"}.StructuredData()

	if got, _ := data["@type"].(string); got != "WebSite" {
		t.Errorf("@type = %q, want WebSite", got)
	}
}

// renderHead renders a page through the real layout and returns the HTML. The
// earlier tests only exercised the Go helpers, which missed the JSON-LD being
// emitted as literal template source rather than rendered.
func renderHead(t *testing.T, seo SEO) string {
	t.Helper()

	var buf bytes.Buffer
	if err := LayoutSEO(templ.NopComponent, seo, "/", "csrf-token", false).Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render the layout: %v", err)
	}

	return buf.String()
}

func structuredDataFrom(t *testing.T, html string) map[string]any {
	t.Helper()

	match := regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`).FindStringSubmatch(html)
	if match == nil {
		t.Fatal("no JSON-LD block was rendered")
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
		t.Fatalf("rendered JSON-LD is not valid JSON: %v\n%s", err, match[1])
	}

	return data
}

// The rendered page must carry real JSON, not the template expression.
func TestLayoutSEO_RendersStructuredData(t *testing.T) {
	published := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	html := renderHead(t, SEO{
		Title:       "Фитнес",
		Description: "Описание",
		Path:        "/blog/fitnes",
		PublishedAt: &published,
	})

	data := structuredDataFrom(t, html)
	if got, _ := data["@type"].(string); got != "Article" {
		t.Errorf("@type = %q, want Article", got)
	}

	if strings.Contains(html, "templ.Raw") || strings.Contains(html, "StructuredData()") {
		t.Error("the template expression leaked into the rendered output")
	}
}

// Post content must not be able to close the surrounding script tag.
func TestLayoutSEO_StructuredDataEscapesScriptTags(t *testing.T) {
	html := renderHead(t, SEO{Title: `</script><script>alert(1)</script>`, Description: "x"})

	data := structuredDataFrom(t, html)
	if got, _ := data["name"].(string); got == "" {
		t.Error("expected the website object to still render")
	}

	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Error("an unescaped script tag was rendered into the page")
	}
}

// Social scrapers read these; a missing card is the failure mode.
func TestLayoutSEO_RendersSocialTags(t *testing.T) {
	html := renderHead(t, SEO{Title: "Заглавие", Description: "Описание", Path: "/blog"})

	for _, want := range []string{
		`<meta property="og:type" content="website">`,
		`<meta property="og:title" content="Заглавие">`,
		`<meta property="og:url" content="https://dviji.se/blog">`,
		`<meta property="og:image" content="https://dviji.se/static/img/logo_1024x1024.png">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<link rel="canonical" href="https://dviji.se/blog">`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered page is missing %s", want)
		}
	}
}
