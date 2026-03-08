package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	storage "github.com/KOMKZ/go-yogan-domain-storage"
	storageerrors "github.com/KOMKZ/go-yogan-domain-storage/errors"
	"github.com/KOMKZ/go-yogan-domain-storage/model"
	"github.com/KOMKZ/go-yogan-domain-storage/repository"
	"github.com/KOMKZ/go-yogan-framework/logger"
	"go.uber.org/zap"
)

// StorageService 文件存储服务
type StorageService struct {
	driver storage.Driver
	repo   repository.StorageRepository
	config *storage.Config
	logger *logger.CtxZapLogger
}

// NewStorageService 创建存储服务
func NewStorageService(driver storage.Driver, repo repository.StorageRepository, config *storage.Config, log *logger.CtxZapLogger) *StorageService {
	return &StorageService{driver: driver, repo: repo, config: config, logger: log}
}

// UploadInput 上传输入
type UploadInput struct {
	Reader       io.Reader
	OriginalName string
	FileSize     int64
	ContentType  string
	Storage      string
	Category     string
	BusinessID   string
	BusinessType string
}

// FileInfo 上传结果
type FileInfo struct {
	ID               uint   `json:"file_id"`
	StorageID        string `json:"storage_id"`
	URL              string `json:"url"`
	OriginalFilename string `json:"original_filename"`
	FileSize         int64  `json:"file_size"`
	ContentType      string `json:"content_type"`
}

// Upload 上传文件
func (s *StorageService) Upload(ctx context.Context, input *UploadInput) (*FileInfo, error) {
	storageType := s.config.GetStorageType(input.Storage)
	if storageType == nil {
		return nil, storageerrors.ErrBadRequest.WithMsgf("unsupported storage type: %s", input.Storage)
	}

	maxBytes := storageType.MaxSize * 1024 * 1024
	if input.FileSize > maxBytes {
		return nil, storageerrors.ErrBadRequest.WithMsgf("file size exceeds limit: max %dMB", storageType.MaxSize)
	}

	ext := strings.ToLower(filepath.Ext(input.OriginalName))
	if len(storageType.AllowedTypes) > 0 && !contains(storageType.AllowedTypes, ext) {
		return nil, storageerrors.ErrBadRequest.WithMsgf("unsupported file type: %s", ext)
	}

	result, err := s.driver.Upload(&storage.UploadRequest{
		Reader:       input.Reader,
		OriginalName: input.OriginalName,
		FileSize:     input.FileSize,
		ContentType:  input.ContentType,
		Storage:      input.Storage,
		Category:     input.Category,
		BusinessID:   input.BusinessID,
		BusinessType: input.BusinessType,
	})
	if err != nil {
		s.logger.ErrorCtx(ctx, "file upload failed", zap.Error(err))
		return nil, storageerrors.ErrUploadFailed.WithMsg("file upload failed")
	}

	url := s.GetURLByStorageID(result.StorageID)

	metadata := &model.FileMetadata{
		StorageID:        result.StorageID,
		OriginalFilename: input.OriginalName,
		StoredFilename:   result.StoredFilename,
		FilePath:         result.FilePath,
		URL:              url,
		FileSize:         input.FileSize,
		ContentType:      input.ContentType,
		Driver:           s.driver.Name(),
		Visibility:       storageType.Visibility,
		Category:         input.Category,
		BusinessID:       input.BusinessID,
		BusinessType:     input.BusinessType,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.Create(ctx, metadata); err != nil {
		s.logger.ErrorCtx(ctx, "save file metadata failed", zap.Error(err))
		_ = s.driver.Delete(result.FilePath)
		return nil, storageerrors.ErrDatabaseError.Wrap(err)
	}

	s.logger.InfoCtx(ctx, "file uploaded",
		zap.Uint("file_id", metadata.ID),
		zap.String("storage_id", result.StorageID))

	return &FileInfo{
		ID:               metadata.ID,
		StorageID:        result.StorageID,
		URL:              url,
		OriginalFilename: input.OriginalName,
		FileSize:         input.FileSize,
		ContentType:      input.ContentType,
	}, nil
}

// Delete 按 ID 删除文件
func (s *StorageService) Delete(ctx context.Context, id uint) error {
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return storageerrors.ErrDatabaseError.Wrap(err)
	}
	if file == nil {
		return storageerrors.ErrNotFound.WithMsg("file not found")
	}

	if err := s.driver.Delete(file.FilePath); err != nil {
		s.logger.WarnCtx(ctx, "delete physical file failed",
			zap.String("path", file.FilePath), zap.Error(err))
	}

	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return storageerrors.ErrDatabaseError.Wrap(err)
	}

	s.logger.InfoCtx(ctx, "file deleted", zap.Uint("file_id", id))
	return nil
}

// DeleteByStorageID 按 StorageID 删除文件
func (s *StorageService) DeleteByStorageID(ctx context.Context, storageID string) error {
	file, err := s.repo.FindByStorageID(ctx, storageID)
	if err != nil {
		return storageerrors.ErrDatabaseError.Wrap(err)
	}
	if file == nil {
		return storageerrors.ErrNotFound.WithMsg("file not found")
	}

	if err := s.driver.Delete(file.FilePath); err != nil {
		s.logger.WarnCtx(ctx, "delete physical file failed",
			zap.String("path", file.FilePath), zap.Error(err))
	}

	if err := s.repo.DeleteByStorageID(ctx, storageID); err != nil {
		return storageerrors.ErrDatabaseError.Wrap(err)
	}

	s.logger.InfoCtx(ctx, "file deleted", zap.String("storage_id", storageID))
	return nil
}

// GetFileInfo 获取文件信息
func (s *StorageService) GetFileInfo(ctx context.Context, id uint) (*model.FileMetadata, error) {
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, storageerrors.ErrDatabaseError.Wrap(err)
	}
	if file == nil {
		return nil, storageerrors.ErrNotFound.WithMsg("file not found")
	}
	return file, nil
}

// GetFileInfoByStorageID 根据存储ID获取文件信息
func (s *StorageService) GetFileInfoByStorageID(ctx context.Context, storageID string) (*model.FileMetadata, error) {
	file, err := s.repo.FindByStorageID(ctx, storageID)
	if err != nil {
		return nil, storageerrors.ErrDatabaseError.Wrap(err)
	}
	if file == nil {
		return nil, storageerrors.ErrNotFound.WithMsg("file not found")
	}
	return file, nil
}

// GetReader 获取文件读取流
func (s *StorageService) GetReader(ctx context.Context, storageID string) (io.ReadCloser, *model.FileMetadata, error) {
	file, err := s.repo.FindByStorageID(ctx, storageID)
	if err != nil {
		return nil, nil, storageerrors.ErrDatabaseError.Wrap(err)
	}
	if file == nil {
		return nil, nil, storageerrors.ErrNotFound.WithMsg("file not found")
	}

	reader, err := s.driver.GetReader(file.FilePath)
	if err != nil {
		return nil, nil, storageerrors.ErrReadFailed.WithMsg("read file failed")
	}

	return reader, file, nil
}

// GetReaderByPath 根据StorageID直接获取文件读取流（无需数据库查询）
func (s *StorageService) GetReaderByPath(ctx context.Context, storageID string) (io.ReadCloser, string, error) {
	sid, err := storage.ParseStorageID(storageID)
	if err != nil {
		return nil, "", storageerrors.ErrBadRequest.WithMsg("invalid storage id")
	}

	storageType := s.config.GetStorageType(sid.Storage)
	if storageType == nil {
		return nil, "", storageerrors.ErrBadRequest.WithMsgf("unknown storage type: %s", sid.Storage)
	}

	var filePath string
	if s.config.Driver == "local" {
		filePath = fmt.Sprintf("%s/%s/%s", s.config.Local.BasePath, sid.Storage, sid.Filename)
	} else {
		filePath = fmt.Sprintf("%s/%s", sid.Storage, sid.Filename)
	}

	reader, err := s.driver.GetReader(filePath)
	if err != nil {
		return nil, "", storageerrors.ErrNotFound.WithMsg("file not found")
	}

	ext := filepath.Ext(sid.Filename)
	contentType := getContentType(ext)

	return reader, contentType, nil
}

// GetURLByStorageID 根据存储ID生成URL
func (s *StorageService) GetURLByStorageID(storageID string) string {
	return s.config.Local.URLPrefix + storageID
}

// UploadFromURL 从URL转存文件到永久存储
func (s *StorageService) UploadFromURL(ctx context.Context, sourceURL string, ext string) (string, error) {
	uploader, ok := s.driver.(storage.URLUploader)
	if !ok {
		return "", storageerrors.ErrBadRequest.WithMsg("driver does not support URL upload")
	}

	permanentURL, err := uploader.UploadFromURL(sourceURL, "ai-generated", ext)
	if err != nil {
		s.logger.ErrorCtx(ctx, "upload from URL failed",
			zap.String("source_url", sourceURL), zap.Error(err))
		return "", storageerrors.ErrUploadFailed.Wrap(err)
	}

	s.logger.InfoCtx(ctx, "file uploaded from URL",
		zap.String("source_url", sourceURL),
		zap.String("permanent_url", permanentURL))

	return permanentURL, nil
}

// UploadFromReader 从Reader上传到永久存储
func (s *StorageService) UploadFromReader(ctx context.Context, reader io.Reader, ext string) (string, error) {
	uploader, ok := s.driver.(storage.ReaderUploader)
	if !ok {
		return "", storageerrors.ErrBadRequest.WithMsg("driver does not support reader upload")
	}

	permanentURL, err := uploader.UploadFromReader(reader, "merged-assets", ext)
	if err != nil {
		s.logger.ErrorCtx(ctx, "upload from reader failed", zap.Error(err))
		return "", storageerrors.ErrUploadFailed.Wrap(err)
	}

	s.logger.InfoCtx(ctx, "file uploaded from reader",
		zap.String("permanent_url", permanentURL))

	return permanentURL, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getContentType(ext string) string {
	types := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".mp4":  "video/mp4",
	}
	if ct, ok := types[strings.ToLower(ext)]; ok {
		return ct
	}
	return "application/octet-stream"
}
