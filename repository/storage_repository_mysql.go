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

func (r *StorageMySQLRepository) Paginate(
	ctx context.Context,
	page,
	pageSize int,
	category,
	businessType,
	keyword,
	sortBy,
	sortOrder string,
) ([]model.FileMetadata, int64, int64, error) {
	var files []model.FileMetadata
	var total int64
	var totalSize int64

	buildQuery := func() *gorm.DB {
		query := r.db.WithContext(ctx).Model(&model.FileMetadata{})
		if category != "" {
			query = query.Where("category = ?", category)
		}
		if businessType != "" {
			query = query.Where("business_type = ?", businessType)
		}
		if keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where(
				"original_filename LIKE ? OR stored_filename LIKE ? OR storage_id LIKE ?",
				like,
				like,
				like,
			)
		}
		return query
	}

	orderBy := "id DESC"
	if sortOrder == "asc" {
		switch sortBy {
		case "created_at":
			orderBy = "created_at ASC"
		case "file_size":
			orderBy = "file_size ASC"
		case "id":
			orderBy = "id ASC"
		}
	} else {
		switch sortBy {
		case "created_at":
			orderBy = "created_at DESC"
		case "file_size":
			orderBy = "file_size DESC"
		case "id":
			orderBy = "id DESC"
		}
	}

	query := buildQuery()
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	query = buildQuery()
	if err := query.Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize).Error; err != nil {
		return nil, 0, 0, err
	}

	offset := (page - 1) * pageSize
	query = buildQuery()
	if err := query.Offset(offset).Limit(pageSize).Order(orderBy).Find(&files).Error; err != nil {
		return nil, 0, 0, err
	}

	return files, total, totalSize, nil
}
