package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"server/internal/application/categories"
	appPosts "server/internal/application/posts"
	"server/internal/domain/posts"
	"server/internal/http/handlers/models"
	"server/internal/infrastructure/cloudinary"
	"server/util"
	"server/util/ctxutils"
	"server/util/httputils"
	"server/web/templates/admin"

	"github.com/google/uuid"
)

type AdminHandler struct {
	postService       *appPosts.PostService
	categoryService   *categories.CategoryService
	cloudinaryService *cloudinary.CloudinaryService
}

func NewAdminHandler(
	postService *appPosts.PostService,
	categoryService *categories.CategoryService,
	cloudinaryService *cloudinary.CloudinaryService,
) *AdminHandler {
	return &AdminHandler{
		postService:       postService,
		categoryService:   categoryService,
		cloudinaryService: cloudinaryService,
	}
}

func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	recentPosts, err := h.postService.GetRecent(ctx, 5)
	if err != nil {
		slog.Error("Error fetching recent posts", "error", err)
		recentPosts = []posts.PostWithAuthor{}
	}

	allPosts, total, err := h.postService.GetAll(ctx, 1, 1)
	if err != nil {
		slog.Error("Error fetching posts count", "error", err)
		total = 0
	}
	_ = allPosts

	publishedPosts, publishedCount, err := h.postService.GetByStatus(ctx, posts.PostStatusPublished, 1, 1)
	if err != nil {
		slog.Error("Error fetching published posts count", "error", err)
		publishedCount = 0
	}
	_ = publishedPosts

	draftPosts, draftCount, err := h.postService.GetByStatus(ctx, posts.PostStatusDraft, 1, 1)
	if err != nil {
		slog.Error("Error fetching draft posts count", "error", err)
		draftCount = 0
	}
	_ = draftPosts

	recentItems := models.PostListFromDomain(recentPosts)
	util.Must(admin.Dashboard(total, publishedCount, draftCount, recentItems).Render(r.Context(), w))
}

func (h *AdminHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	status := r.URL.Query().Get("status")

	var domainPosts []posts.PostWithAuthor
	var total int
	var err error

	if status != "" && posts.PostStatus(status) != "" {
		domainPosts, total, err = h.postService.GetByStatus(ctx, posts.PostStatus(status), page, pageSize)
	} else {
		domainPosts, total, err = h.postService.GetAll(ctx, page, pageSize)
	}

	if err != nil {
		slog.Error("Error fetching posts", "error", err)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	postItems := models.PostListFromDomain(domainPosts)
	totalPages := (total + pageSize - 1) / pageSize

	util.Must(admin.PostsList(postItems, page, totalPages, total, status).Render(r.Context(), w))
}

func (h *AdminHandler) GetPostForm(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	allCategories, err := h.categoryService.GetCategories(ctx)
	if err != nil {
		slog.Error("Error fetching categories", "error", err)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	var categoryResources []models.CategoryResponseResource
	for _, cat := range allCategories {
		var resource models.CategoryResponseResource
		categoryResources = append(categoryResources, resource.CreateCategoryResponseFrom(&cat))
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		util.Must(admin.PostForm(nil, categoryResources).Render(r.Context(), w))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		httputils.SendBadRequestResponse(w, "Invalid post ID")
		return
	}

	post, err := h.postService.GetById(ctx, id)
	if err != nil {
		slog.Error("Error fetching post", "error", err, "id", id)
		httputils.SendNotFoundResponse(w, "Post not found")
		return
	}

	util.Must(admin.PostForm(post, categoryResources).Render(r.Context(), w))
}

func (h *AdminHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	var input models.CreatePostResource
	if !httputils.ProcessRequestBody(w, r, &input) {
		return
	}

	user, err := ctxutils.GetUser(r.Context())
	if err != nil {
		httputils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	creatorId, err := uuid.Parse(user.Id)
	if err != nil {
		httputils.SendBadRequestResponse(w, "Invalid user ID")
		return
	}

	categoryId, err := uuid.Parse(input.CategoryId)
	if err != nil {
		httputils.SendBadRequestResponse(w, "Invalid category ID")
		return
	}

	status := posts.PostStatusCreated
	if input.Status != "" {
		status = posts.PostStatus(input.Status)
	}

	createInput := appPosts.CreatePostInput{
		Title:           input.Title,
		Content:         input.Content,
		Excerpt:         input.Excerpt,
		CoverImageUrl:   input.CoverImageUrl,
		CategoryId:      categoryId,
		MetaDescription: input.MetaDescription,
		Status:          status,
	}

	post, err := h.postService.Create(ctx, createInput, creatorId)
	if err != nil {
		slog.Error("Error creating post", "error", err)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	slog.Info(fmt.Sprintf("Successfully created post [id=%s]", post.Id.String()))
	httputils.SendSuccessResponse(w, "Post created successfully", map[string]string{"id": post.Id.String()}, http.StatusCreated)
}

func (h *AdminHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputils.SendBadRequestResponse(w, "Invalid post ID")
		return
	}

	var input models.UpdatePostResource
	if !httputils.ProcessRequestBody(w, r, &input) {
		return
	}

	user, err := ctxutils.GetUser(r.Context())
	if err != nil {
		httputils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	categoryId, err := uuid.Parse(input.CategoryId)
	if err != nil {
		httputils.SendBadRequestResponse(w, "Invalid category ID")
		return
	}

	updateInput := appPosts.UpdatePostInput{
		Title:           input.Title,
		Content:         input.Content,
		Excerpt:         input.Excerpt,
		CoverImageUrl:   input.CoverImageUrl,
		CategoryId:      categoryId,
		MetaDescription: input.MetaDescription,
		Status:          posts.PostStatus(input.Status),
	}

	post, err := h.postService.Update(ctx, id, updateInput, user.Username)
	if err != nil {
		slog.Error("Error updating post", "error", err, "id", id)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	slog.Info(fmt.Sprintf("Successfully updated post [id=%s]", post.Id.String()))
	httputils.SendSuccessResponse(w, "Post updated successfully", map[string]string{"id": post.Id.String()}, http.StatusOK)
}

func (h *AdminHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputils.SendBadRequestResponse(w, "Invalid post ID")
		return
	}

	user, err := ctxutils.GetUser(r.Context())
	if err != nil {
		httputils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.postService.Delete(ctx, id, user.Username)
	if err != nil {
		slog.Error("Error deleting post", "error", err, "id", id)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	slog.Info(fmt.Sprintf("Successfully deleted post [id=%s]", id.String()))
	httputils.SendSuccessResponse(w, "Post deleted successfully", nil, http.StatusOK)
}

func (h *AdminHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), cancelTime)
	defer cancel()

	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		httputils.SendBadRequestResponse(w, "Failed to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httputils.SendBadRequestResponse(w, "No file uploaded")
		return
	}
	defer file.Close()

	if h.cloudinaryService == nil {
		httputils.SendErrorResponse(w, "Image upload not configured", http.StatusServiceUnavailable)
		return
	}

	result, err := h.cloudinaryService.Upload(ctx, file, header.Filename)
	if err != nil {
		slog.Error("Error uploading image", "error", err)
		httputils.SendInternalServerResponse(w, r)
		return
	}

	httputils.SendSuccessResponse(w, "Image uploaded successfully", map[string]string{
		"location": result.URL,
	}, http.StatusOK)
}
