package handlers

import (
	"context"
	"encoding/xml"
	"fmt"
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

	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	encoder.Encode(feed)
}

func (h *FeedHandler) GetSitemap(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	baseURL := config.BaseURL()

	domainPosts, _, err := h.postService.GetPublished(ctx, 1, 1000)
	if err != nil {
		domainPosts = nil
	}

	sitemap := models.SitemapFromPosts(domainPosts, baseURL)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	encoder.Encode(sitemap)
}

func (h *FeedHandler) GetRobotsTxt(w http.ResponseWriter, r *http.Request) {
	baseURL := config.BaseURL()

	robotsTxt := fmt.Sprintf(`User-agent: *
Allow: /

Sitemap: %s/sitemap.xml
`, baseURL)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write([]byte(robotsTxt))
}
