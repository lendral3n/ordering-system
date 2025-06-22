package media

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudinaryURL string) (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudinary client: %w", err)
	}
	
	return &CloudinaryService{
		cld: cld,
	}, nil
}

type UploadResult struct {
	PublicID     string
	SecureURL    string
	ThumbnailURL string
	Format       string
	ResourceType string
	FileSize     int64
	Width        int
	Height       int
}

func (s *CloudinaryService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// Determine folder based on file type
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	folder := "restaurant/menu"
	resourceType := "image"
	
	if isVideo(fileExt) {
		folder = "restaurant/videos"
		resourceType = "video"
	}
	
	// Upload to Cloudinary
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: resourceType,
		Format:       "auto",
	}
	
	// Add transformations for images
	if resourceType == "image" {
		uploadParams.Transformation = "q_auto,f_auto"
	}
	
	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to cloudinary: %w", err)
	}
	
	// Generate thumbnail URL for images
	thumbnailURL := ""
	if resourceType == "image" {
		thumbnailURL, _ = s.cld.Image(result.PublicID).Transform("w_200,h_200,c_fill,g_auto").URL()
	}
	
	return &UploadResult{
		PublicID:     result.PublicID,
		SecureURL:    result.SecureURL,
		ThumbnailURL: thumbnailURL,
		Format:       result.Format,
		ResourceType: result.ResourceType,
		FileSize:     int64(result.Bytes),
		Width:        result.Width,
		Height:       result.Height,
	}, nil
}

func (s *CloudinaryService) DeleteFile(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	
	if err != nil {
		return fmt.Errorf("failed to delete from cloudinary: %w", err)
	}
	
	return nil
}

func (s *CloudinaryService) Generate360URL(publicID string) string {
	// For 360 images, we can add special transformations
	url, _ := s.cld.Image(publicID).Transform("w_1024,h_512,c_limit,q_auto").URL()
	return url
}

// Helper functions
func isVideo(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm"}
	for _, vExt := range videoExts {
		if ext == vExt {
			return true
		}
	}
	return false
}

func isImage(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp"}
	for _, iExt := range imageExts {
		if ext == iExt {
			return true
		}
	}
	return false
}