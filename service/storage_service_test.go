package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	storage "github.com/KOMKZ/go-yogan-domain-storage"
	storageerrors "github.com/KOMKZ/go-yogan-domain-storage/errors"
	"github.com/KOMKZ/go-yogan-domain-storage/model"
	"github.com/KOMKZ/go-yogan-domain-storage/repository"
	"github.com/KOMKZ/go-yogan-framework/logger"
)

// --- mock driver ---

type mockDriver struct {
	name         string
	uploadResult *storage.UploadResult
	uploadErr    error
	deleteErr    error
	existsVal    bool
	getURLVal    string
	readerData   string
	readerErr    error
	deleteCalls  []string
}

func newMockDriver() *mockDriver {
	return &mockDriver{name: "mock"}
}

func (d *mockDriver) Name() string { return d.name }

func (d *mockDriver) Upload(req *storage.UploadRequest) (*storage.UploadResult, error) {
	if d.uploadErr != nil {
		return nil, d.uploadErr
	}
	if d.uploadResult != nil {
		return d.uploadResult, nil
	}
	return &storage.UploadResult{
		StorageID:      "mock:test@file.jpg",
		FilePath:       "/tmp/file.jpg",
		StoredFilename: "file.jpg",
	}, nil
}

func (d *mockDriver) Delete(filePath string) error {
	d.deleteCalls = append(d.deleteCalls, filePath)
	return d.deleteErr
}

func (d *mockDriver) GetURL(filePath string) string { return d.getURLVal }
func (d *mockDriver) Exists(filePath string) bool   { return d.existsVal }

func (d *mockDriver) GetReader(filePath string) (io.ReadCloser, error) {
	if d.readerErr != nil {
		return nil, d.readerErr
	}
	return io.NopCloser(strings.NewReader(d.readerData)), nil
}

// --- mock repo ---

type mockRepo struct {
	mu             sync.RWMutex
	files          map[uint]*model.FileMetadata
	byStorageID    map[string]*model.FileMetadata
	nextID         uint
	createErr      error
	findByIDErr    error
	findBySIDErr   error
	deleteErr      error
	deleteBySIDErr error
	paginateErr    error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		files:       make(map[uint]*model.FileMetadata),
		byStorageID: make(map[string]*model.FileMetadata),
		nextID:      1,
	}
}

func (r *mockRepo) Create(ctx context.Context, file *model.FileMetadata) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	file.ID = r.nextID
	r.nextID++
	cp := *file
	r.files[file.ID] = &cp
	r.byStorageID[file.StorageID] = &cp
	return nil
}

func (r *mockRepo) FindByID(ctx context.Context, id uint) (*model.FileMetadata, error) {
	if r.findByIDErr != nil {
		return nil, r.findByIDErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.files[id]
	if !ok {
		return nil, nil
	}
	cp := *f
	return &cp, nil
}

func (r *mockRepo) FindByStorageID(ctx context.Context, storageID string) (*model.FileMetadata, error) {
	if r.findBySIDErr != nil {
		return nil, r.findBySIDErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.byStorageID[storageID]
	if !ok {
		return nil, nil
	}
	cp := *f
	return &cp, nil
}

func (r *mockRepo) FindByPath(ctx context.Context, filePath string) (*model.FileMetadata, error) {
	return nil, nil
}

func (r *mockRepo) DeleteByID(ctx context.Context, id uint) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if f, ok := r.files[id]; ok {
		delete(r.byStorageID, f.StorageID)
		delete(r.files, id)
	}
	return nil
}

func (r *mockRepo) DeleteByStorageID(ctx context.Context, storageID string) error {
	if r.deleteBySIDErr != nil {
		return r.deleteBySIDErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if f, ok := r.byStorageID[storageID]; ok {
		delete(r.files, f.ID)
		delete(r.byStorageID, storageID)
	}
	return nil
}

func (r *mockRepo) Paginate(
	ctx context.Context,
	page,
	pageSize int,
	category,
	businessType,
	keyword,
	sortBy,
	sortOrder string,
) ([]model.FileMetadata, int64, int64, error) {
	if r.paginateErr != nil {
		return nil, 0, 0, r.paginateErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []model.FileMetadata
	var totalSize int64
	for _, f := range r.files {
		if category != "" && f.Category != category {
			continue
		}
		if businessType != "" && f.BusinessType != businessType {
			continue
		}
		if keyword != "" {
			if !strings.Contains(f.OriginalFilename, keyword) && !strings.Contains(f.StoredFilename, keyword) && !strings.Contains(f.StorageID, keyword) {
				continue
			}
		}
		list = append(list, *f)
		totalSize += f.FileSize
	}
	if sortBy == "file_size" {
		for i := 0; i < len(list); i++ {
			for j := i + 1; j < len(list); j++ {
				if sortOrder == "asc" {
					if list[i].FileSize > list[j].FileSize {
						list[i], list[j] = list[j], list[i]
					}
				} else {
					if list[i].FileSize < list[j].FileSize {
						list[i], list[j] = list[j], list[i]
					}
				}
			}
		}
	}
	return list, int64(len(list)), totalSize, nil
}

var _ repository.StorageRepository = (*mockRepo)(nil)

// --- mock URL uploader driver ---

type mockURLDriver struct {
	mockDriver
	uploadFromURLResult    string
	uploadFromURLErr       error
	uploadFromReaderResult string
	uploadFromReaderErr    error
}

func (d *mockURLDriver) UploadFromURL(sourceURL string, s string, ext string) (string, error) {
	if d.uploadFromURLErr != nil {
		return "", d.uploadFromURLErr
	}
	return d.uploadFromURLResult, nil
}

func (d *mockURLDriver) UploadFromReader(reader io.Reader, s string, ext string) (string, error) {
	if d.uploadFromReaderErr != nil {
		return "", d.uploadFromReaderErr
	}
	return d.uploadFromReaderResult, nil
}

var (
	_ storage.Driver         = (*mockURLDriver)(nil)
	_ storage.URLUploader    = (*mockURLDriver)(nil)
	_ storage.ReaderUploader = (*mockURLDriver)(nil)
)

// --- helpers ---

func testConfig() *storage.Config {
	return &storage.Config{
		Driver: "local",
		Local:  storage.LocalConfig{BasePath: "./storage", URLPrefix: "/api/files/storage/"},
		StorageTypes: map[string]storage.StorageType{
			"image": {Visibility: "pub", Path: "images", MaxSize: 5, AllowedTypes: []string{".jpg", ".png"}},
			"doc":   {Visibility: "pri", Path: "docs", MaxSize: 10, AllowedTypes: []string{".pdf"}},
		},
	}
}

func testLogger() *logger.CtxZapLogger {
	return logger.GetLogger("storage_test")
}

func newTestService(driver storage.Driver, repo repository.StorageRepository) *StorageService {
	return NewStorageService(driver, repo, testConfig(), testLogger())
}

// --- Upload tests ---

func TestUpload_Success(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	svc := newTestService(d, r)

	info, err := svc.Upload(context.Background(), &UploadInput{
		Reader:       strings.NewReader("image data"),
		OriginalName: "photo.jpg",
		FileSize:     1024,
		ContentType:  "image/jpeg",
		Storage:      "image",
		Category:     "test",
	})
	if err != nil {
		t.Fatalf("Upload() err = %v", err)
	}
	if info.StorageID == "" {
		t.Error("StorageID should not be empty")
	}
	if info.URL == "" {
		t.Error("URL should not be empty")
	}
}

func TestUpload_UnsupportedStorageType(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, err := svc.Upload(context.Background(), &UploadInput{
		Reader:       strings.NewReader("data"),
		OriginalName: "file.txt",
		Storage:      "nonexistent",
	})
	if err == nil {
		t.Error("should fail with unsupported storage type")
	}
}

func TestUpload_FileSizeExceedsLimit(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, err := svc.Upload(context.Background(), &UploadInput{
		Reader:       strings.NewReader("data"),
		OriginalName: "big.jpg",
		FileSize:     10 * 1024 * 1024,
		Storage:      "image",
	})
	if err == nil {
		t.Error("should fail when file size exceeds limit")
	}
}

func TestUpload_UnsupportedFileType(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, err := svc.Upload(context.Background(), &UploadInput{
		Reader:       strings.NewReader("data"),
		OriginalName: "script.exe",
		FileSize:     100,
		Storage:      "image",
	})
	if err == nil {
		t.Error("should fail with unsupported file type")
	}
}

func TestUpload_DriverFailure(t *testing.T) {
	d := newMockDriver()
	d.uploadErr = errors.New("disk full")
	svc := newTestService(d, newMockRepo())

	_, err := svc.Upload(context.Background(), &UploadInput{
		Reader:       strings.NewReader("data"),
		OriginalName: "file.jpg",
		FileSize:     100,
		Storage:      "image",
	})
	if err == nil {
		t.Error("should fail when driver fails")
	}
}

func TestUpload_RepoCreateFailure_CleansUpFile(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	r.createErr = errors.New("db error")
	svc := newTestService(d, r)

	_, err := svc.Upload(context.Background(), &UploadInput{
		Reader:       strings.NewReader("data"),
		OriginalName: "file.jpg",
		FileSize:     100,
		Storage:      "image",
	})
	if err == nil {
		t.Error("should fail when repo create fails")
	}
	if len(d.deleteCalls) == 0 {
		t.Error("should attempt to delete uploaded file on repo failure")
	}
}

// --- Delete tests ---

func TestDelete_Success(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	if err := svc.Delete(context.Background(), info.ID); err != nil {
		t.Fatalf("Delete() err = %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	err := svc.Delete(context.Background(), 999)
	if err == nil {
		t.Error("Delete non-existent should fail")
	}
}

func TestDelete_RepoFindError(t *testing.T) {
	r := newMockRepo()
	r.findByIDErr = errors.New("db error")
	svc := newTestService(newMockDriver(), r)

	err := svc.Delete(context.Background(), 1)
	if err == nil {
		t.Error("should fail when repo find fails")
	}
}

func TestDelete_RepoDeleteError(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	r.deleteErr = errors.New("delete failed")
	err := svc.Delete(context.Background(), info.ID)
	if err == nil {
		t.Error("should fail when repo delete fails")
	}
}

// --- DeleteByStorageID tests ---

func TestDeleteByStorageID_Success(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	if err := svc.DeleteByStorageID(context.Background(), info.StorageID); err != nil {
		t.Fatalf("DeleteByStorageID() err = %v", err)
	}
}

func TestDeleteByStorageID_NotFound(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	err := svc.DeleteByStorageID(context.Background(), "mock:test@nonexistent.jpg")
	if err == nil {
		t.Error("should fail for non-existent storage ID")
	}
}

func TestDeleteByStorageID_RepoError(t *testing.T) {
	r := newMockRepo()
	r.findBySIDErr = errors.New("db error")
	svc := newTestService(newMockDriver(), r)

	err := svc.DeleteByStorageID(context.Background(), "mock:test@file.jpg")
	if err == nil {
		t.Error("should fail when repo find fails")
	}
}

// --- GetFileInfo tests ---

func TestGetFileInfo_Success(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, ContentType: "image/jpeg", Storage: "image",
	})

	file, err := svc.GetFileInfo(context.Background(), info.ID)
	if err != nil {
		t.Fatalf("GetFileInfo() err = %v", err)
	}
	if file.OriginalFilename != "f.jpg" {
		t.Errorf("OriginalFilename = %q, want f.jpg", file.OriginalFilename)
	}
}

func TestGetFileInfo_NotFound(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, err := svc.GetFileInfo(context.Background(), 999)
	if err == nil {
		t.Error("should fail for non-existent ID")
	}
}

// --- GetFileInfoByStorageID tests ---

func TestGetFileInfoByStorageID_Success(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	file, err := svc.GetFileInfoByStorageID(context.Background(), info.StorageID)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if file == nil {
		t.Fatal("file should not be nil")
	}
}

func TestGetFileInfoByStorageID_NotFound(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, err := svc.GetFileInfoByStorageID(context.Background(), "x:y@z")
	if err == nil {
		t.Error("should fail for non-existent")
	}
}

// --- GetReader tests ---

func TestGetReader_Success(t *testing.T) {
	d := newMockDriver()
	d.readerData = "file content"
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	reader, meta, err := svc.GetReader(context.Background(), info.StorageID)
	if err != nil {
		t.Fatalf("GetReader() err = %v", err)
	}
	defer reader.Close()
	if meta == nil {
		t.Fatal("metadata should not be nil")
	}

	data, _ := io.ReadAll(reader)
	if string(data) != "file content" {
		t.Errorf("content = %q, want 'file content'", string(data))
	}
}

func TestGetReader_NotFound(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, _, err := svc.GetReader(context.Background(), "x:y@z")
	if err == nil {
		t.Error("should fail for non-existent")
	}
}

func TestGetReader_DriverReadError(t *testing.T) {
	d := newMockDriver()
	d.readerErr = errors.New("read failed")
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	_, _, err := svc.GetReader(context.Background(), info.StorageID)
	if err == nil {
		t.Error("should fail when driver read fails")
	}
}

// --- GetReaderByPath tests ---

func TestGetReaderByPath_Success(t *testing.T) {
	d := newMockDriver()
	d.readerData = "content"
	r := newMockRepo()
	svc := newTestService(d, r)

	reader, ct, err := svc.GetReaderByPath(context.Background(), "local:image@test.jpg")
	if err != nil {
		t.Fatalf("GetReaderByPath() err = %v", err)
	}
	defer reader.Close()

	if ct != "image/jpeg" {
		t.Errorf("contentType = %q, want image/jpeg", ct)
	}
}

func TestGetReaderByPath_InvalidStorageID(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, _, err := svc.GetReaderByPath(context.Background(), "invalid")
	if err == nil {
		t.Error("should fail with invalid storage ID")
	}
}

func TestGetReaderByPath_UnknownStorageType(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, _, err := svc.GetReaderByPath(context.Background(), "local:unknown@file.jpg")
	if err == nil {
		t.Error("should fail with unknown storage type")
	}
}

// --- GetURLByStorageID tests ---

func TestGetURLByStorageID(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	url := svc.GetURLByStorageID("local:avatar@test.jpg")
	if url != "/api/files/storage/local:avatar@test.jpg" {
		t.Errorf("URL = %q", url)
	}
}

// --- UploadFromURL tests ---

func TestUploadFromURL_Success(t *testing.T) {
	d := &mockURLDriver{uploadFromURLResult: "https://oss.example.com/file.png"}
	d.name = "mock-oss"
	r := newMockRepo()
	svc := NewStorageService(d, r, testConfig(), testLogger())

	url, err := svc.UploadFromURL(context.Background(), "https://tmp.example.com/img.png", ".png")
	if err != nil {
		t.Fatalf("UploadFromURL() err = %v", err)
	}
	if url != "https://oss.example.com/file.png" {
		t.Errorf("url = %q", url)
	}
}

func TestUploadFromURL_DriverNotSupported(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, err := svc.UploadFromURL(context.Background(), "https://example.com/img.png", ".png")
	if err == nil {
		t.Error("should fail when driver doesn't support URL upload")
	}
}

func TestUploadFromURL_DriverError(t *testing.T) {
	d := &mockURLDriver{uploadFromURLErr: errors.New("upload failed")}
	d.name = "mock-oss"
	svc := NewStorageService(d, newMockRepo(), testConfig(), testLogger())

	_, err := svc.UploadFromURL(context.Background(), "https://example.com/img.png", ".png")
	if err == nil {
		t.Error("should fail when driver upload fails")
	}
}

// --- UploadFromReader tests ---

func TestUploadFromReader_Success(t *testing.T) {
	d := &mockURLDriver{uploadFromReaderResult: "https://oss.example.com/merged.png"}
	d.name = "mock-oss"
	svc := NewStorageService(d, newMockRepo(), testConfig(), testLogger())

	url, err := svc.UploadFromReader(context.Background(), bytes.NewReader([]byte("data")), ".png")
	if err != nil {
		t.Fatalf("UploadFromReader() err = %v", err)
	}
	if url != "https://oss.example.com/merged.png" {
		t.Errorf("url = %q", url)
	}
}

func TestUploadFromReader_DriverNotSupported(t *testing.T) {
	svc := newTestService(newMockDriver(), newMockRepo())
	_, err := svc.UploadFromReader(context.Background(), bytes.NewReader([]byte("data")), ".png")
	if err == nil {
		t.Error("should fail when driver doesn't support reader upload")
	}
}

func TestUploadFromReader_DriverError(t *testing.T) {
	d := &mockURLDriver{uploadFromReaderErr: errors.New("upload failed")}
	d.name = "mock-oss"
	svc := NewStorageService(d, newMockRepo(), testConfig(), testLogger())

	_, err := svc.UploadFromReader(context.Background(), bytes.NewReader([]byte("data")), ".png")
	if err == nil {
		t.Error("should fail when driver upload fails")
	}
}

// --- DeleteByStorageID additional tests ---

func TestDeleteByStorageID_RepoDeleteError(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	r.deleteBySIDErr = errors.New("delete failed")
	err := svc.DeleteByStorageID(context.Background(), info.StorageID)
	if err == nil {
		t.Error("should fail when repo delete by storage id fails")
	}
}

// --- Delete with driver delete error (non-fatal) ---

func TestDelete_DriverDeleteError_NonFatal(t *testing.T) {
	d := newMockDriver()
	d.deleteErr = errors.New("fs error")
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	d.deleteErr = errors.New("fs error")
	err := svc.Delete(context.Background(), info.ID)
	if err != nil {
		t.Errorf("Delete should succeed even if driver delete fails, got: %v", err)
	}
}

func TestDeleteByStorageID_DriverDeleteError_NonFatal(t *testing.T) {
	d := newMockDriver()
	r := newMockRepo()
	svc := newTestService(d, r)

	info, _ := svc.Upload(context.Background(), &UploadInput{
		Reader: strings.NewReader("data"), OriginalName: "f.jpg",
		FileSize: 10, Storage: "image",
	})

	d.deleteErr = errors.New("fs error")
	err := svc.DeleteByStorageID(context.Background(), info.StorageID)
	if err != nil {
		t.Errorf("DeleteByStorageID should succeed even if driver delete fails, got: %v", err)
	}
}

// --- GetFileInfo repo error ---

func TestGetFileInfo_RepoError(t *testing.T) {
	r := newMockRepo()
	r.findByIDErr = errors.New("db error")
	svc := newTestService(newMockDriver(), r)

	_, err := svc.GetFileInfo(context.Background(), 1)
	if err == nil {
		t.Error("should fail when repo find fails")
	}
}

// --- GetFileInfoByStorageID repo error ---

func TestGetFileInfoByStorageID_RepoError(t *testing.T) {
	r := newMockRepo()
	r.findBySIDErr = errors.New("db error")
	svc := newTestService(newMockDriver(), r)

	_, err := svc.GetFileInfoByStorageID(context.Background(), "x:y@z")
	if err == nil {
		t.Error("should fail when repo find by storage id fails")
	}
}

// --- GetReader repo error ---

func TestGetReader_RepoError(t *testing.T) {
	r := newMockRepo()
	r.findBySIDErr = errors.New("db error")
	svc := newTestService(newMockDriver(), r)

	_, _, err := svc.GetReader(context.Background(), "x:y@z")
	if err == nil {
		t.Error("should fail when repo find fails")
	}
}

// --- GetReaderByPath driver read error ---

func TestGetReaderByPath_DriverReadError(t *testing.T) {
	d := newMockDriver()
	d.readerErr = errors.New("read error")
	svc := newTestService(d, newMockRepo())

	_, _, err := svc.GetReaderByPath(context.Background(), "local:image@test.jpg")
	if err == nil {
		t.Error("should fail when driver read fails")
	}
}

func TestGetReaderByPath_NonLocalDriver(t *testing.T) {
	d := newMockDriver()
	d.readerData = "oss content"
	r := newMockRepo()

	cfg := &storage.Config{
		Driver: "oss",
		StorageTypes: map[string]storage.StorageType{
			"image": {Visibility: "pub", Path: "images", MaxSize: 5, AllowedTypes: []string{".jpg", ".png"}},
		},
	}
	svc := NewStorageService(d, r, cfg, testLogger())

	reader, ct, err := svc.GetReaderByPath(context.Background(), "oss:image@test.png")
	if err != nil {
		t.Fatalf("GetReaderByPath() err = %v", err)
	}
	defer reader.Close()

	if ct != "image/png" {
		t.Errorf("contentType = %q, want image/png", ct)
	}

	data, _ := io.ReadAll(reader)
	if string(data) != "oss content" {
		t.Errorf("content = %q, want 'oss content'", string(data))
	}
}

// --- helper tests ---

func TestContains(t *testing.T) {
	if !contains([]string{".jpg", ".png"}, ".jpg") {
		t.Error("should contain .jpg")
	}
	if contains([]string{".jpg", ".png"}, ".gif") {
		t.Error("should not contain .gif")
	}
	if contains(nil, ".jpg") {
		t.Error("nil slice should not contain anything")
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".jpg", "image/jpeg"},
		{".JPG", "image/jpeg"},
		{".png", "image/png"},
		{".pdf", "application/pdf"},
		{".unknown", "application/octet-stream"},
		{".mp4", "video/mp4"},
	}
	for _, tt := range tests {
		got := getContentType(tt.ext)
		if got != tt.want {
			t.Errorf("getContentType(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

// suppress unused import
var _ = storageerrors.ErrNotFound
