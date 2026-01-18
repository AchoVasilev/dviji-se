package posts

import (
	"context"
	"database/sql"
	"regexp"
	"server/internal/domain/posts"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type postRepository interface {
	Create(ctx context.Context, post posts.Post) (*posts.Post, error)
	Update(ctx context.Context, post posts.Post) (*posts.Post, error)
	Delete(ctx context.Context, id uuid.UUID, deletedBy string) error
	FindById(ctx context.Context, id uuid.UUID) (*posts.Post, error)
	FindBySlug(ctx context.Context, slug string) (*posts.PostWithAuthor, error)
	FindPublished(ctx context.Context, limit, offset int) ([]posts.PostWithAuthor, int, error)
	FindByCategory(ctx context.Context, categorySlug string, limit, offset int) ([]posts.PostWithAuthor, int, error)
	FindAll(ctx context.Context, limit, offset int) ([]posts.PostWithAuthor, int, error)
	FindByStatus(ctx context.Context, status posts.PostStatus, limit, offset int) ([]posts.PostWithAuthor, int, error)
	FindRecent(ctx context.Context, limit int) ([]posts.PostWithAuthor, error)
	ExistsBySlug(ctx context.Context, slug string, excludeId *uuid.UUID) (bool, error)
}

type CreatePostInput struct {
	Title           string
	Content         string
	Excerpt         string
	CoverImageUrl   string
	CategoryId      uuid.UUID
	MetaDescription string
	Status          posts.PostStatus
}

type UpdatePostInput struct {
	Title           string
	Content         string
	Excerpt         string
	CoverImageUrl   string
	CategoryId      uuid.UUID
	MetaDescription string
	Status          posts.PostStatus
}

type PostService struct {
	postRepository postRepository
}

func NewPostService(repo postRepository) *PostService {
	return &PostService{postRepository: repo}
}

func (s *PostService) Create(ctx context.Context, input CreatePostInput, creatorId uuid.UUID) (*posts.Post, error) {
	slug := s.GenerateSlug(input.Title)

	exists, err := s.postRepository.ExistsBySlug(ctx, slug, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	status := input.Status
	if status == "" {
		status = posts.PostStatusCreated
	}

	post := posts.Post{
		Id:                 uuid.New(),
		Title:              input.Title,
		Slug:               slug,
		Content:            input.Content,
		Excerpt:            input.Excerpt,
		CoverImageUrl:      input.CoverImageUrl,
		Status:             status,
		MetaDescription:    input.MetaDescription,
		ReadingTimeMinutes: s.CalculateReadingTime(input.Content),
		CategoryId:         input.CategoryId,
		CreatorUserId:      creatorId,
		CreatedAt:          time.Now().UTC(),
	}

	if status == posts.PostStatusPublished {
		post.PublishedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}

	return s.postRepository.Create(ctx, post)
}

func (s *PostService) Update(ctx context.Context, id uuid.UUID, input UpdatePostInput, updatedBy string) (*posts.Post, error) {
	existing, err := s.postRepository.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	slug := s.GenerateSlug(input.Title)
	if slug != existing.Slug {
		exists, err := s.postRepository.ExistsBySlug(ctx, slug, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			slug = slug + "-" + uuid.New().String()[:8]
		}
	} else {
		slug = existing.Slug
	}

	post := posts.Post{
		Id:                 id,
		Title:              input.Title,
		Slug:               slug,
		Content:            input.Content,
		Excerpt:            input.Excerpt,
		CoverImageUrl:      input.CoverImageUrl,
		Status:             input.Status,
		PublishedAt:        existing.PublishedAt,
		MetaDescription:    input.MetaDescription,
		ReadingTimeMinutes: s.CalculateReadingTime(input.Content),
		CategoryId:         input.CategoryId,
		UpdatedBy:          updatedBy,
	}

	if input.Status == posts.PostStatusPublished && !existing.PublishedAt.Valid {
		post.PublishedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}

	return s.postRepository.Update(ctx, post)
}

func (s *PostService) Delete(ctx context.Context, id uuid.UUID, deletedBy string) error {
	return s.postRepository.Delete(ctx, id, deletedBy)
}

func (s *PostService) GetById(ctx context.Context, id uuid.UUID) (*posts.Post, error) {
	return s.postRepository.FindById(ctx, id)
}

func (s *PostService) GetBySlug(ctx context.Context, slug string) (*posts.PostWithAuthor, error) {
	return s.postRepository.FindBySlug(ctx, slug)
}

func (s *PostService) GetPublished(ctx context.Context, page, pageSize int) ([]posts.PostWithAuthor, int, error) {
	offset := (page - 1) * pageSize
	return s.postRepository.FindPublished(ctx, pageSize, offset)
}

func (s *PostService) GetByCategory(ctx context.Context, categorySlug string, page, pageSize int) ([]posts.PostWithAuthor, int, error) {
	offset := (page - 1) * pageSize
	return s.postRepository.FindByCategory(ctx, categorySlug, pageSize, offset)
}

func (s *PostService) GetAll(ctx context.Context, page, pageSize int) ([]posts.PostWithAuthor, int, error) {
	offset := (page - 1) * pageSize
	return s.postRepository.FindAll(ctx, pageSize, offset)
}

func (s *PostService) GetByStatus(ctx context.Context, status posts.PostStatus, page, pageSize int) ([]posts.PostWithAuthor, int, error) {
	offset := (page - 1) * pageSize
	return s.postRepository.FindByStatus(ctx, status, pageSize, offset)
}

func (s *PostService) GetRecent(ctx context.Context, limit int) ([]posts.PostWithAuthor, error) {
	return s.postRepository.FindRecent(ctx, limit)
}

func (s *PostService) GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = transliterate(slug)

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	if len(slug) > 100 {
		slug = slug[:100]
		if lastDash := strings.LastIndex(slug, "-"); lastDash > 50 {
			slug = slug[:lastDash]
		}
	}

	return slug
}

func (s *PostService) CalculateReadingTime(content string) int {
	words := len(strings.Fields(content))
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

var cyrillicToLatin = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ж': "zh",
	'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
	'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f",
	'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sht", 'ъ': "a", 'ь': "",
	'ю': "yu", 'я': "ya",
	'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ж': "Zh",
	'З': "Z", 'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N",
	'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U", 'Ф': "F",
	'Х': "H", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Sht", 'Ъ': "A", 'Ь': "",
	'Ю': "Yu", 'Я': "Ya",
}

func transliterate(text string) string {
	var result strings.Builder
	for _, r := range text {
		if latin, ok := cyrillicToLatin[r]; ok {
			result.WriteString(latin)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	return result.String()
}
