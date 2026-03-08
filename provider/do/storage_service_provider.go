package do

import (
	"github.com/KOMKZ/go-yogan-domain-storage/repository"
	"github.com/KOMKZ/go-yogan-domain-storage/service"
	"github.com/KOMKZ/go-yogan-framework/logger"
	samberdo "github.com/samber/do/v2"
	"gorm.io/gorm"

	storage "github.com/KOMKZ/go-yogan-domain-storage"
)

// ---- Repository Providers ----

func ProvideStorageRepository(i samberdo.Injector) (repository.StorageRepository, error) {
	db, err := samberdo.Invoke[*gorm.DB](i)
	if err != nil {
		return nil, err
	}
	return repository.NewStorageMySQLRepository(db), nil
}

// ---- Service Providers ----

// ProvideStorageService 从 DI 容器创建 StorageService
func ProvideStorageService(i samberdo.Injector) (*service.StorageService, error) {
	driver := samberdo.MustInvoke[storage.Driver](i)
	repo := samberdo.MustInvoke[repository.StorageRepository](i)
	config := samberdo.MustInvoke[*storage.Config](i)
	log := samberdo.MustInvokeNamed[*logger.CtxZapLogger](i, "storage")

	return service.NewStorageService(driver, repo, config, log), nil
}
