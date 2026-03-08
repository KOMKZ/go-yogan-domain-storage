package storage

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// LocalDriver 本地文件系统存储驱动
type LocalDriver struct {
	basePath  string
	urlPrefix string
}

// NewLocalDriver 创建本地存储驱动
func NewLocalDriver(basePath, urlPrefix string) *LocalDriver {
	return &LocalDriver{basePath: basePath, urlPrefix: urlPrefix}
}

func (d *LocalDriver) Name() string { return "local" }

func (d *LocalDriver) Upload(req *UploadRequest) (*UploadResult, error) {
	if req.Storage == "" {
		return nil, fmt.Errorf("storage identifier is required")
	}

	ext := filepath.Ext(req.OriginalName)
	storedFilename := generateShortUUID() + ext

	storageID, err := BuildStorageID(d.Name(), req.Storage, storedFilename)
	if err != nil {
		return nil, fmt.Errorf("build storage id: %w", err)
	}

	dirPath := filepath.Join(d.basePath, req.Storage)
	filePath := filepath.Join(dirPath, storedFilename)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, req.Reader); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return &UploadResult{
		StorageID:      storageID,
		FilePath:       filePath,
		StoredFilename: storedFilename,
	}, nil
}

func (d *LocalDriver) Delete(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(filePath)
}

func (d *LocalDriver) GetURL(filePath string) string {
	return d.urlPrefix + filePath
}

func (d *LocalDriver) Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func (d *LocalDriver) GetReader(filePath string) (io.ReadCloser, error) {
	return os.Open(filePath)
}

func generateShortUUID() string {
	u := uuid.New()
	hash := md5.Sum(u[:])
	return hex.EncodeToString(hash[:8])
}
