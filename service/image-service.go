package service

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/gin-gonic/gin"
)

var allowedFormats = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func UploadImage(c *gin.Context, entityType string, entityID uint, backend internal.StorageBackend) (*model.Image, error) {
	db, err := internal.GetGormInstance()
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	// 1. making sure entities exist
	switch entityType {
	case "user":
		var user model.User
		if err := db.First(&user, entityID).Error; err != nil {
			return nil, errors.New("user not found")
		}
	case "recipe":
		var recipe model.Recipe
		if err := db.First(&recipe, entityID).Error; err != nil {
			return nil, errors.New("recipe not found")
		}
	default:
		return nil, errors.New("unsupported entity type")
	}

	// 2. Enforce only one image for users
	if entityType == "user" {
		var count int64
		if err := db.Model(&model.Image{}).
			Where("entity_type = ? AND entity_id = ?", entityType, entityID).
			Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to check existing user images: %w", err)
		}
		if count > 0 {
			return nil, fmt.Errorf("a profile image already exists for user ID %d", entityID)
		}
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return nil, fmt.Errorf("image file is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %v", err)
	}
	contentType := http.DetectContentType(buf[:n])

	if !allowedFormats[contentType] {
		return nil, fmt.Errorf("unsupported image format: %s", contentType)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		}
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dstPath := filepath.Join(entityType, fmt.Sprintf("%d", entityID), filename)

	if err := backend.Save(dstPath, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	img := model.Image{
		EntityType: entityType,
		EntityID:   entityID,
		Path:       dstPath,
		Format:     contentType,
		Size:       fileHeader.Size,
		CreatedAt:  time.Now(),
	}

	if err := db.Create(&img).Error; err != nil {
		_ = backend.Delete(dstPath)
		return nil, err
	}

	return &img, nil
}

func ServeImage(c *gin.Context, entityType string, entityID, imageID uint, backend internal.StorageBackend) error {
	db, err := internal.GetGormInstance()
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}

	var img model.Image
	err = db.Where("id = ? AND entity_type = ? AND entity_id = ?", imageID, entityType, entityID).First(&img).Error
	if err != nil {
		return err
	}

	c.Header("Content-Type", img.Format)

	localBackend, ok := backend.(*internal.LocalStorage)
	if ok {
		fullPath := filepath.Join(localBackend.BasePath, img.Path)
		c.File(fullPath)
		return nil
	}

	return fmt.Errorf("ServeImage not implemented for this backend")
}

func DeleteImage(entityType string, entityID, imageID uint, backend internal.StorageBackend) error {
	db, err := internal.GetGormInstance()
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}

	var img model.Image
	err = db.Where("id = ? AND entity_type = ? AND entity_id = ?", imageID, entityType, entityID).First(&img).Error
	if err != nil {
		return err
	}

	if err := backend.Delete(img.Path); err != nil {
		return fmt.Errorf("failed to delete image file: %v", err)
	}

	if err := db.Delete(&img).Error; err != nil {
		return fmt.Errorf("failed to delete image record: %v", err)
	}

	return nil
}

func GetImagesForEntity(entityType string, entityID uint) ([]model.Image, error) {
	db, err := internal.GetGormInstance()
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	var images []model.Image
	err = db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).Find(&images).Error
	return images, err
}

func DeleteImagesForEntity(entityType string, entityID uint, backend internal.StorageBackend) error {
	db, err := internal.GetGormInstance()
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}

	var images []model.Image
	if err := db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).Find(&images).Error; err != nil {
		return err
	}

	for _, img := range images {
		if err := backend.Delete(img.Path); err != nil {
			return fmt.Errorf("failed to delete image file %s: %v", img.Path, err)
		}
	}

	if err := db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).Delete(&model.Image{}).Error; err != nil {
		return err
	}

	return nil
}
