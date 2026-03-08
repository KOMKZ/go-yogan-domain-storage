package repository

import (
	"context"
	"errors"

	"github.com/KOMKZ/go-yogan-domain-storage/model"
	"github.com/KOMKZ/go-yogan-framework/database"
	"gorm.io/gorm"
)

type StorageMySQLRepository struct {
	base *database.BaseRepository[model.FileMetadata]
	db   *gorm.DB
}

func NewStorageMySQLRepository(db *gorm.DB) *StorageMySQLRepository {
	return &StorageMySQLRepository{
		base: database.NewBaseRepository[model.FileMetadata](db),
		db:   db,
	}
}

func (r *StorageMySQLRepository) Create(ctx context.Context, file *model.FileMetadata) error {
	return r.base.Create(ctx, file)
}

func (r *StorageMySQLRepository) FindByID(ctx context.Context, id uint) (*model.FileMetadata, error) {
	result, err := r.base.FindByID(ctx, id)
	if errors.Is(err, database.ErrRecordNotFound) {
		return nil, nil
	}
	return result, err
}

func (r *StorageMySQLRepository) FindByStorageID(ctx context.Context, storageID string) (*model.FileMetadata, error) {
	var file model.FileMetadata
	err := r.db.WithContext(ctx).Where("storage_id = ?", storageID).First(&file).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *StorageMySQLRepository) FindByPath(ctx context.Context, filePath string) (*model.FileMetadata, error) {
	var file model.FileMetadata
	err := r.db.WithContext(ctx).Where("file_path = ?", filePath).First(&file).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *StorageMySQLRepository) DeleteByID(ctx context.Context, id uint) error {
	return r.base.Delete(ctx, id)
}

func (r *StorageMySQLRepository) DeleteByStorageID(ctx context.Context, storageID string) error {
	return r.db.WithContext(ctx).Where("storage_id = ?", storageID).Delete(&model.FileMetadata{}).Error
}

func (r *StorageMySQLRepository) Paginate(ctx context.Context, page, pageSize int, category, businessType string) ([]model.FileMetadata, int64, error) {
	var files []model.FileMetadata
	var total int64

	query := r.db.WithContext(ctx).Model(&model.FileMetadata{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if businessType != "" {
		query = query.Where("business_type = ?", businessType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}
