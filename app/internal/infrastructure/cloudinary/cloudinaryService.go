package cloudinary

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"server/internal/config"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type UploadResult struct {
	URL      string
	PublicID string
}

type CloudinaryService struct {
	client *cloudinary.Cloudinary
	folder string
}

func NewCloudinaryService() (*CloudinaryService, error) {
	if !config.CloudinaryConfigured() {
		return nil, fmt.Errorf("cloudinary credentials not configured")
	}

	cld, err := cloudinary.NewFromParams(
		config.CloudinaryCloudName(),
		config.CloudinaryAPIKey(),
		config.CloudinaryAPISecret(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudinary client: %w", err)
	}

	return &CloudinaryService{
		client: cld,
		folder: config.CloudinaryFolder(),
	}, nil
}

func (s *CloudinaryService) Upload(ctx context.Context, file multipart.File, filename string) (*UploadResult, error) {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	uploadParams := uploader.UploadParams{
		Folder:   s.folder,
		PublicID: name,
	}

	result, err := s.client.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	return &UploadResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
	}, nil
}

func (s *CloudinaryService) Delete(ctx context.Context, publicId string) error {
	_, err := s.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicId,
	})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}
