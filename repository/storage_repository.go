package repository

import (
	"context"

	"github.com/KOMKZ/go-yogan-domain-storage/model"
)

// StorageRepository 文件元数据仓储接口
// 约定：记录不存在时返回 (nil, nil)，仅基础设施异常返回 error
type StorageRepository interface {
	Create(ctx context.Context, file *model.FileMetadata) error
	FindByID(ctx context.Context, id uint) (*model.FileMetadata, error)
	FindByStorageID(ctx context.Context, storageID string) (*model.FileMetadata, error)
	FindByPath(ctx context.Context, filePath string) (*model.FileMetadata, error)
	DeleteByID(ctx context.Context, id uint) error
	DeleteByStorageID(ctx context.Context, storageID string) error
	Paginate(
		ctx context.Context,
		page,
		pageSize int,
		category,
		businessType,
		keyword,
		sortBy,
		sortOrder string,
	) ([]model.FileMetadata, int64, int64, error)
}
