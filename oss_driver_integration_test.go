package storage

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func getOSSConfig() *OSSConfig {
	return &OSSConfig{
		Endpoint:        os.Getenv("ALIYUN_OSS_ENDPOINT"),
		AccessKeyID:     os.Getenv("ALIYUN_OSS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ALIYUN_OSS_ACCESS_KEY_SECRET"),
		BucketName:      os.Getenv("ALIYUN_OSS_BUCKET_NAME"),
		BaseURL:         "https://" + os.Getenv("ALIYUN_OSS_BUCKET_NAME") + "." + os.Getenv("ALIYUN_OSS_ENDPOINT"),
		UploadPath:      "test-integration",
	}
}

func skipIfNoOSSEnv(t *testing.T) *OSSConfig {
	t.Helper()
	cfg := getOSSConfig()
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.BucketName == "" {
		t.Skip("skipping OSS integration test: missing OSS environment variables")
	}
	return cfg
}

func TestOSSDriver_Integration_NewDriver(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	if driver.Name() != "oss" {
		t.Errorf("Name() = %q, want oss", driver.Name())
	}
}

func TestOSSDriver_Integration_Upload(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	content := []byte("test content for OSS upload integration test")
	req := &UploadRequest{
		OriginalName: "test-upload.txt",
		Reader:       bytes.NewReader(content),
		Storage:      "image",
	}

	result, err := driver.Upload(req)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	if result.StorageID == "" {
		t.Error("Upload() StorageID is empty")
	}
	if result.FilePath == "" {
		t.Error("Upload() FilePath is empty")
	}
	if result.URL == "" {
		t.Error("Upload() URL is empty")
	}
	if !strings.HasPrefix(result.StorageID, "oss:image@") {
		t.Errorf("Upload() StorageID = %q, should start with oss:image@", result.StorageID)
	}

	t.Cleanup(func() {
		_ = driver.Delete(result.FilePath)
	})
}

func TestOSSDriver_Integration_Exists(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	content := []byte("test exists check")
	req := &UploadRequest{
		OriginalName: "test-exists.txt",
		Reader:       bytes.NewReader(content),
		Storage:      "image",
	}

	result, err := driver.Upload(req)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	t.Cleanup(func() {
		_ = driver.Delete(result.FilePath)
	})

	if !driver.Exists(result.FilePath) {
		t.Error("Exists() = false after upload, want true")
	}
}

func TestOSSDriver_Integration_Delete(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	content := []byte("test delete file")
	req := &UploadRequest{
		OriginalName: "test-delete.txt",
		Reader:       bytes.NewReader(content),
		Storage:      "image",
	}

	result, err := driver.Upload(req)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	if err := driver.Delete(result.FilePath); err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	if driver.Exists(result.FilePath) {
		t.Error("Exists() = true after delete, want false")
	}
}

func TestOSSDriver_Integration_GetReader(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	content := []byte("test read content back")
	req := &UploadRequest{
		OriginalName: "test-read.txt",
		Reader:       bytes.NewReader(content),
		Storage:      "image",
	}

	result, err := driver.Upload(req)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	t.Cleanup(func() {
		_ = driver.Delete(result.FilePath)
	})

	reader, err := driver.GetReader(result.FilePath)
	if err != nil {
		t.Fatalf("GetReader() error = %v", err)
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	if buf.String() != string(content) {
		t.Errorf("GetReader() content = %q, want %q", buf.String(), string(content))
	}
}

func TestOSSDriver_Integration_UploadFromURL(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	testURL := "https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png"

	url, err := driver.UploadFromURL(testURL, "image", ".png")
	if err != nil {
		t.Fatalf("UploadFromURL() error = %v", err)
	}

	if url == "" {
		t.Error("UploadFromURL() returned empty URL")
	}

	if !strings.Contains(url, cfg.BucketName) {
		t.Errorf("UploadFromURL() URL = %q, should contain bucket name", url)
	}
}

func TestOSSDriver_Integration_UploadFromReader(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	content := []byte("test upload from reader content")
	reader := bytes.NewReader(content)

	url, err := driver.UploadFromReader(reader, "document", ".txt")
	if err != nil {
		t.Fatalf("UploadFromReader() error = %v", err)
	}

	if url == "" {
		t.Error("UploadFromReader() returned empty URL")
	}

	if !strings.Contains(url, cfg.BucketName) {
		t.Errorf("UploadFromReader() URL = %q, should contain bucket name", url)
	}
}

func TestOSSDriver_Integration_GetURL_WithoutBaseURL(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)
	cfg.BaseURL = ""

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	url := driver.GetURL("test/path/file.jpg")
	if url == "" {
		t.Error("GetURL() returned empty URL")
	}

	if !strings.Contains(url, cfg.BucketName) {
		t.Errorf("GetURL() URL = %q, should contain bucket name", url)
	}
}

func TestOSSDriver_Integration_UploadFromURL_InvalidURL(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	_, err = driver.UploadFromURL("http://invalid.nonexistent.domain/file.png", "image", ".png")
	if err == nil {
		t.Error("UploadFromURL() with invalid URL should fail")
	}
}

func TestOSSDriver_Integration_UploadFromURL_404(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	_, err = driver.UploadFromURL("https://httpstat.us/404", "image", ".png")
	if err == nil {
		t.Error("UploadFromURL() with 404 response should fail")
	}
}

func TestOSSDriver_Integration_UploadFromURL_DefaultExt(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	testURL := "https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png"
	url, err := driver.UploadFromURL(testURL, "image", "")
	if err != nil {
		t.Fatalf("UploadFromURL() error = %v", err)
	}

	if !strings.HasSuffix(url, ".png") {
		t.Errorf("UploadFromURL() URL = %q, should end with .png (default)", url)
	}
}

func TestOSSDriver_Integration_UploadFromReader_DefaultExt(t *testing.T) {
	cfg := skipIfNoOSSEnv(t)

	driver, err := NewOSSDriver(cfg)
	if err != nil {
		t.Fatalf("NewOSSDriver() error = %v", err)
	}

	content := []byte("test reader content")
	url, err := driver.UploadFromReader(bytes.NewReader(content), "document", "")
	if err != nil {
		t.Fatalf("UploadFromReader() error = %v", err)
	}

	if !strings.HasSuffix(url, ".png") {
		t.Errorf("UploadFromReader() URL = %q, should end with .png (default)", url)
	}
}
