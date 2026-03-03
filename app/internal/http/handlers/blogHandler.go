package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"server/internal/application/categories"
	appPosts "server/internal/application/posts"
	"server/internal/domain/posts"
	"server/internal/http/handlers/models"
	"server/util"
	"server/util/httputils"
	"server/web/templates"
)

type BlogHandler struct {
	postService     *appPosts.PostService
	categoryService *categories.CategoryService
}

func NewBlogHandler(
	postService *appPosts.PostService,
	categoryService *categories.CategoryService,
) *BlogHandler {
	return &BlogHandler{
		postService:     postService,
		categoryService: categoryService,
	}
}

func (h *BlogHandler) GetBlogList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	page := 1
	pageSize := 12

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	domainPosts, total, err := h.postService.GetPublished(ctx, page, pageSize)
	if err != nil {
		slog.Error("Error fetching published posts", "error", err)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	allCategories, err := h.categoryService.GetCategories(ctx)
	if err != nil {
		slog.Error("Error fetching categories", "error", err)
		allCategories = nil
	}

	var categoryResources []models.CategoryResponseResource
	for _, cat := range allCategories {
		var resource models.CategoryResponseResource
		categoryResources = append(categoryResources, resource.CreateCategoryResponseFrom(&cat))
	}

	postItems := models.PostListFromDomain(domainPosts)
	totalPages := (total + pageSize - 1) / pageSize

	util.Must(templates.BlogList(postItems, categoryResources, page, totalPages, total, "").Render(r.Context(), w))
}

func (h *BlogHandler) GetBlogPost(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	slug := r.PathValue("slug")
	if slug == "" {
		httputils.SendNotFoundResponse(w, "Post not found")
		return
	}

	post, err := h.postService.GetBySlug(ctx, slug)
	if err != nil {
		slog.Error("Error fetching post by slug", "error", err, "slug", slug)
		httputils.SendNotFoundResponse(w, "Post not found")
		return
	}

	if post.Status != posts.PostStatusPublished {
		httputils.SendNotFoundResponse(w, "Post not found")
		return
	}

	recentPosts, err := h.postService.GetRecent(ctx, 3)
	if err != nil {
		slog.Error("Error fetching recent posts", "error", err)
		recentPosts = []posts.PostWithAuthor{}
	}

	postResponse := models.PostResponseFromDomain(post)
	recentItems := models.PostListFromDomain(recentPosts)

	util.Must(templates.BlogPost(postResponse, recentItems).Render(r.Context(), w))
}

func (h *BlogHandler) GetBlogByCategory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	categorySlug := r.PathValue("slug")
	if categorySlug == "" {
		httputils.SendNotFoundResponse(w, "Category not found")
		return
	}

	page := 1
	pageSize := 12

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	domainPosts, total, err := h.postService.GetByCategory(ctx, categorySlug, page, pageSize)
	if err != nil {
		slog.Error("Error fetching posts by category", "error", err, "categorySlug", categorySlug)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	allCategories, err := h.categoryService.GetCategories(ctx)
	if err != nil {
		slog.Error("Error fetching categories", "error", err)
		allCategories = nil
	}

	var categoryResources []models.CategoryResponseResource
	for _, cat := range allCategories {
		var resource models.CategoryResponseResource
		categoryResources = append(categoryResources, resource.CreateCategoryResponseFrom(&cat))
	}

	postItems := models.PostListFromDomain(domainPosts)
	totalPages := (total + pageSize - 1) / pageSize

	util.Must(templates.BlogList(postItems, categoryResources, page, totalPages, total, categorySlug).Render(r.Context(), w))
}

func (h *BlogHandler) SearchSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	domainPosts, _, err := h.postService.SearchPublished(ctx, q, 1, 5)
	if err != nil {
		slog.Error("Error searching posts", "error", err, "query", q)
		w.WriteHeader(http.StatusOK)
		return
	}

	postItems := models.PostListFromDomain(domainPosts)
	util.Must(templates.SearchSuggestions(postItems, q).Render(r.Context(), w))
}

func (h *BlogHandler) SearchBlogPosts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	page := 1
	pageSize := 12

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	var postItems []models.PostListItem
	var total, totalPages int

	if q != "" {
		domainPosts, t, err := h.postService.SearchPublished(ctx, q, page, pageSize)
		if err != nil {
			slog.Error("Error searching posts", "error", err, "query", q)
			httputils.SendInternalServerResponse(w, r)
			return
		}
		total = t
		totalPages = (total + pageSize - 1) / pageSize
		postItems = models.PostListFromDomain(domainPosts)
	}

	util.Must(templates.BlogSearchResults(postItems, q, page, totalPages, total).Render(r.Context(), w))
}

func (h *BlogHandler) GetRecentPosts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	limit := 6
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 20 {
			limit = parsed
		}
	}

	recentPosts, err := h.postService.GetRecent(ctx, limit)
	if err != nil {
		slog.Error("Error fetching recent posts", "error", err)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	recentItems := models.PostListFromDomain(recentPosts)
	util.Must(templates.RecentPosts(recentItems).Render(r.Context(), w))
}
