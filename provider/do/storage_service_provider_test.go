package do

import (
	"context"
	"io"
	"reflect"
	"testing"

	storage "github.com/KOMKZ/go-yogan-domain-storage"
	"github.com/KOMKZ/go-yogan-domain-storage/model"
	"github.com/KOMKZ/go-yogan-domain-storage/repository"
	"github.com/KOMKZ/go-yogan-framework/logger"
	samberdo "github.com/samber/do/v2"
)

func TestProvideStorageServiceDefined(t *testing.T) {
	fn := ProvideStorageService
	if reflect.ValueOf(fn).IsNil() {
		t.Fatal("ProvideStorageService should be defined")
	}
}

type mockDriver struct{}

func (m *mockDriver) Name() string                                   { return "mock" }
func (m *mockDriver) Upload(req *storage.UploadRequest) (*storage.UploadResult, error) { return nil, nil }
func (m *mockDriver) Delete(filePath string) error                   { return nil }
func (m *mockDriver) GetURL(filePath string) string                  { return "" }
func (m *mockDriver) Exists(filePath string) bool                    { return false }
func (m *mockDriver) GetReader(filePath string) (io.ReadCloser, error) { return nil, nil }

type mockRepo struct{}

func (m *mockRepo) Create(ctx context.Context, file *model.FileMetadata) error     { return nil }
func (m *mockRepo) FindByID(ctx context.Context, id uint) (*model.FileMetadata, error) { return nil, nil }
func (m *mockRepo) FindByStorageID(ctx context.Context, storageID string) (*model.FileMetadata, error) { return nil, nil }
func (m *mockRepo) FindByPath(ctx context.Context, filePath string) (*model.FileMetadata, error) { return nil, nil }
func (m *mockRepo) DeleteByID(ctx context.Context, id uint) error                    { return nil }
func (m *mockRepo) DeleteByStorageID(ctx context.Context, storageID string) error  { return nil }
func (m *mockRepo) Paginate(ctx context.Context, page, pageSize int, category, businessType string) ([]model.FileMetadata, int64, error) { return nil, 0, nil }

func TestProvideStorageService_Execute(t *testing.T) {
	injector := samberdo.New()

	samberdo.ProvideValue[storage.Driver](injector, &mockDriver{})
	samberdo.ProvideValue[repository.StorageRepository](injector, &mockRepo{})
	samberdo.ProvideValue(injector, storage.DefaultConfig())
	samberdo.ProvideNamedValue(injector, "storage", logger.GetLogger("storage"))

	svc, err := ProvideStorageService(injector)
	if err != nil {
		t.Fatalf("ProvideStorageService() error = %v", err)
	}
	if svc == nil {
		t.Error("ProvideStorageService() returned nil")
	}
}
