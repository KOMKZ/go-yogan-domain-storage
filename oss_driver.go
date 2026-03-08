package storage

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
)

// OSSConfig 阿里云 OSS 配置
type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name"`
	BaseURL         string `mapstructure:"base_url"`
	UploadPath      string `mapstructure:"upload_path"`
}

// OSSDriver 阿里云 OSS 存储驱动
type OSSDriver struct {
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
	baseURL    string
	uploadPath string
}

// NewOSSDriver 创建 OSS 存储驱动
func NewOSSDriver(cfg *OSSConfig) (*OSSDriver, error) {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.BucketName == "" {
		return nil, fmt.Errorf("OSS configuration incomplete")
	}

	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("create OSS client: %w", err)
	}

	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}

	return &OSSDriver{
		client:     client,
		bucket:     bucket,
		bucketName: cfg.BucketName,
		baseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
		uploadPath: strings.Trim(cfg.UploadPath, "/"),
	}, nil
}

func (d *OSSDriver) Name() string { return "oss" }

func (d *OSSDriver) Upload(req *UploadRequest) (*UploadResult, error) {
	if req.Storage == "" {
		return nil, fmt.Errorf("storage type cannot be empty")
	}

	ext := filepath.Ext(req.OriginalName)
	storedFilename := generateOSSShortUUID() + ext

	storageID, err := BuildStorageID(d.Name(), req.Storage, storedFilename)
	if err != nil {
		return nil, fmt.Errorf("build storage id: %w", err)
	}

	objectKey := d.buildObjectKey(req.Storage, storedFilename)

	if err := d.bucket.PutObject(objectKey, req.Reader); err != nil {
		return nil, fmt.Errorf("upload to OSS: %w", err)
	}

	_ = d.bucket.SetObjectACL(objectKey, oss.ACLPublicRead)

	return &UploadResult{
		StorageID:      storageID,
		FilePath:       objectKey,
		URL:            d.GetURL(objectKey),
		StoredFilename: storedFilename,
	}, nil
}

func (d *OSSDriver) Delete(filePath string) error {
	return d.bucket.DeleteObject(filePath)
}

func (d *OSSDriver) GetURL(filePath string) string {
	if d.baseURL != "" {
		return fmt.Sprintf("%s/%s", d.baseURL, filePath)
	}
	return fmt.Sprintf("https://%s.%s/%s",
		d.bucketName,
		strings.TrimPrefix(d.client.Config.Endpoint, "https://"),
		filePath,
	)
}

func (d *OSSDriver) Exists(filePath string) bool {
	exist, _ := d.bucket.IsObjectExist(filePath)
	return exist
}

func (d *OSSDriver) GetReader(filePath string) (io.ReadCloser, error) {
	return d.bucket.GetObject(filePath)
}

// UploadFromURL 从URL下载并上传到OSS（实现 URLUploader 接口）
func (d *OSSDriver) UploadFromURL(sourceURL string, storage string, ext string) (string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(sourceURL)
	if err != nil {
		return "", fmt.Errorf("download from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if ext == "" {
		ext = ".png"
	}
	storedFilename := generateOSSShortUUID() + ext
	objectKey := d.buildObjectKey(storage, storedFilename)

	if err := d.bucket.PutObject(objectKey, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("upload to OSS: %w", err)
	}

	_ = d.bucket.SetObjectACL(objectKey, oss.ACLPublicRead)
	return d.GetURL(objectKey), nil
}

// UploadFromReader 从Reader上传到OSS（实现 ReaderUploader 接口）
func (d *OSSDriver) UploadFromReader(reader io.Reader, storage string, ext string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read data: %w", err)
	}

	if ext == "" {
		ext = ".png"
	}
	storedFilename := generateOSSShortUUID() + ext
	objectKey := d.buildObjectKey(storage, storedFilename)

	if err := d.bucket.PutObject(objectKey, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("upload to OSS: %w", err)
	}

	_ = d.bucket.SetObjectACL(objectKey, oss.ACLPublicRead)
	return d.GetURL(objectKey), nil
}

func (d *OSSDriver) buildObjectKey(storage, filename string) string {
	if d.uploadPath != "" {
		return fmt.Sprintf("%s/%s/%s", d.uploadPath, storage, filename)
	}
	return fmt.Sprintf("%s/%s", storage, filename)
}

func generateOSSShortUUID() string {
	u := uuid.New()
	hash := md5.Sum(u[:])
	return hex.EncodeToString(hash[:8])
}

// compile-time interface check
var (
	_ Driver         = (*OSSDriver)(nil)
	_ URLUploader    = (*OSSDriver)(nil)
	_ ReaderUploader = (*OSSDriver)(nil)
)
