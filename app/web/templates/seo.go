package templates

import (
	"strings"
	"time"

	"server/internal/config"
	"server/util/imageutils"
)

const (
	siteName       = "Движи се"
	siteLocale     = "bg_BG"
	defaultOGImage = "/static/img/logo_1024x1024.png"
)

// SEO carries the metadata used to build the page's title, canonical URL,
// Open Graph and Twitter cards, and Schema.org markup.
//
// Only Title and Description are required. Pages that describe a post fill in
// the article fields, which switch the page to og:type=article and emit
// Schema.org Article markup.
type SEO struct {
	Title       string
	Description string

	// Path is the canonical path for this page, such as "/blog/my-post".
	// Empty means no canonical or og:url is emitted.
	Path string

	// ImageURL may be absolute (a Cloudinary URL) or a /static path. Empty
	// falls back to the site logo so shared links still render a card.
	ImageURL string

	// Article metadata. PublishedAt being set is what marks this an article.
	PublishedAt *time.Time
	ModifiedAt  *time.Time
	AuthorName  string
	Section     string
}

// IsArticle reports whether the page describes a published post.
func (s SEO) IsArticle() bool {
	return s.PublishedAt != nil
}

// OGType is the Open Graph object type.
func (s SEO) OGType() string {
	if s.IsArticle() {
		return "article"
	}

	return "website"
}

// PageTitle is what belongs in <title>. Articles already carry the site name
// in the suffix, so it is appended once here rather than at each call site.
func (s SEO) PageTitle() string {
	if s.Title == "" {
		return siteName
	}

	return s.Title + " - " + siteName
}

// CanonicalURL is the absolute URL for this page, or "" when Path is unset.
func (s SEO) CanonicalURL() string {
	if s.Path == "" {
		return ""
	}

	return absoluteURL(s.Path)
}

// ImageAbsoluteURL is the absolute URL of the share image. Social scrapers do
// not resolve relative paths, so this must never be relative.
// ogImageWidth matches what the social networks render a preview at.
const ogImageWidth = 1200

func (s SEO) ImageAbsoluteURL() string {
	if s.ImageURL == "" {
		return absoluteURL(defaultOGImage)
	}

	// Scrapers fetch this once per share and the previews they build are about
	// 1200 wide, so an untouched original would be downloaded in full to be
	// shown small. ogImageWidth is what they need and no more.
	return absoluteURL(imageutils.Resized(s.ImageURL, ogImageWidth))
}

// PublishedISO and ModifiedISO render the article timestamps in RFC 3339,
// which is what article:published_time and Schema.org expect.
func (s SEO) PublishedISO() string {
	if s.PublishedAt == nil {
		return ""
	}

	return s.PublishedAt.UTC().Format(time.RFC3339)
}

func (s SEO) ModifiedISO() string {
	if s.ModifiedAt == nil {
		return s.PublishedISO()
	}

	return s.ModifiedAt.UTC().Format(time.RFC3339)
}

// absoluteURL turns a site relative path into an absolute URL. Values that are
// already absolute (Cloudinary, for instance) are returned unchanged.
func absoluteURL(pathOrURL string) string {
	if pathOrURL == "" {
		return ""
	}

	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		return pathOrURL
	}

	return strings.TrimSuffix(config.BaseURL(), "/") + "/" + strings.TrimPrefix(pathOrURL, "/")
}

// StructuredData builds the Schema.org object for this page. It is rendered
// through templ.JSONScript, which encodes it with HTML escaping on, so post
// content cannot break out of the surrounding <script> tag.
func (s SEO) StructuredData() map[string]any {
	if !s.IsArticle() {
		return s.websiteStructuredData()
	}

	article := map[string]any{
		"@context":      "https://schema.org",
		"@type":         "Article",
		"headline":      s.Title,
		"description":   s.Description,
		"datePublished": s.PublishedISO(),
		"dateModified":  s.ModifiedISO(),
		"image":         []string{s.ImageAbsoluteURL()},
		"inLanguage":    "bg",
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  siteName,
			"logo": map[string]any{
				"@type": "ImageObject",
				"url":   absoluteURL(defaultOGImage),
			},
		},
	}

	if url := s.CanonicalURL(); url != "" {
		article["url"] = url
		article["mainEntityOfPage"] = map[string]any{"@type": "WebPage", "@id": url}
	}

	if s.AuthorName != "" {
		article["author"] = map[string]any{"@type": "Person", "name": s.AuthorName}
	}

	if s.Section != "" {
		article["articleSection"] = s.Section
	}

	return article
}

func (s SEO) websiteStructuredData() map[string]any {
	website := map[string]any{
		"@context":   "https://schema.org",
		"@type":      "WebSite",
		"name":       siteName,
		"inLanguage": "bg",
		"url":        strings.TrimSuffix(config.BaseURL(), "/"),
	}

	if s.Description != "" {
		website["description"] = s.Description
	}

	return website
}
