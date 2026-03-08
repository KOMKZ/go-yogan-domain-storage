package model

import "time"

// FileMetadata 文件元数据实体，对应表 file_metadata
type FileMetadata struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	StorageID        string    `gorm:"size:500;uniqueIndex;not null" json:"storage_id"`
	OriginalFilename string    `gorm:"size:255;not null" json:"original_filename"`
	StoredFilename   string    `gorm:"size:255;not null" json:"stored_filename"`
	FilePath         string    `gorm:"size:500;not null;index" json:"file_path"`
	URL              string    `gorm:"size:500" json:"url"`
	FileSize         int64     `gorm:"not null" json:"file_size"`
	ContentType      string    `gorm:"size:100;not null" json:"content_type"`
	Driver           string    `gorm:"size:50;not null" json:"driver"`
	Visibility       string    `gorm:"size:10;not null;default:'pub';index" json:"visibility"`
	Category         string    `gorm:"size:50;index" json:"category"`
	BusinessID       string    `gorm:"size:100" json:"business_id"`
	BusinessType     string    `gorm:"size:50;index" json:"business_type"`
	CreatedAt        time.Time `gorm:"not null;index" json:"created_at"`
	UpdatedAt        time.Time `gorm:"not null" json:"updated_at"`
}

func (FileMetadata) TableName() string { return "file_metadata" }

func (f *FileMetadata) IsPublic() bool  { return f.Visibility == "pub" }
func (f *FileMetadata) IsPrivate() bool { return f.Visibility == "pri" }
