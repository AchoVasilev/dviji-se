package handlers

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"

	appPosts "server/internal/application/posts"
	"server/internal/config"
	"server/internal/http/handlers/models"
)

type FeedHandler struct {
	postService *appPosts.PostService
}

func NewFeedHandler(postService *appPosts.PostService) *FeedHandler {
	return &FeedHandler{
		postService: postService,
	}
}

func (h *FeedHandler) GetRSSFeed(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	domainPosts, _, err := h.postService.GetPublished(ctx, 1, 20)
	if err != nil {
		http.Error(w, "Failed to generate feed", http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL()
	feedURL := baseURL + "/feed.xml"

	feed := models.RSSFromPosts(domainPosts, baseURL, feedURL)

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if _, err := w.Write([]byte(xml.Header)); err != nil {
		slog.WarnContext(ctx, "Failed to write RSS feed", "error", err)
		return
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(feed); err != nil {
		slog.WarnContext(ctx, "Failed to encode RSS feed", "error", err)
	}
}

func (h *FeedHandler) GetSitemap(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	baseURL := config.BaseURL()

	domainPosts, err := h.postService.GetSitemapEntries(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Error fetching sitemap entries", "error", err)
		domainPosts = nil
	}

	sitemap := models.SitemapFromPosts(domainPosts, baseURL)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if _, err := w.Write([]byte(xml.Header)); err != nil {
		slog.WarnContext(ctx, "Failed to write sitemap", "error", err)
		return
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(sitemap); err != nil {
		slog.WarnContext(ctx, "Failed to encode sitemap", "error", err)
	}
}

func (h *FeedHandler) GetRobotsTxt(w http.ResponseWriter, r *http.Request) {
	baseURL := config.BaseURL()

	robotsTxt := fmt.Sprintf(`User-agent: *
Allow: /

Sitemap: %s/sitemap.xml
`, baseURL)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := w.Write([]byte(robotsTxt)); err != nil {
		slog.WarnContext(r.Context(), "Failed to write robots.txt", "error", err)
	}
}
