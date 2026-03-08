package storage

import "io"

// Driver 文件存储驱动接口
type Driver interface {
	Upload(req *UploadRequest) (*UploadResult, error)
	Delete(filePath string) error
	GetURL(filePath string) string
	Exists(filePath string) bool
	GetReader(filePath string) (io.ReadCloser, error)
	Name() string
}

// URLUploader 支持从 URL 转存的驱动（可选能力）
type URLUploader interface {
	UploadFromURL(sourceURL string, storage string, ext string) (string, error)
}

// ReaderUploader 支持从 Reader 上传的驱动（可选能力）
type ReaderUploader interface {
	UploadFromReader(reader io.Reader, storage string, ext string) (string, error)
}

// UploadRequest 上传请求
type UploadRequest struct {
	Reader       io.Reader
	OriginalName string
	FileSize     int64
	ContentType  string
	Storage      string
	Category     string
	BusinessID   string
	BusinessType string
	Overwrite    bool
}

// UploadResult 上传结果
type UploadResult struct {
	StorageID      string
	FilePath       string
	URL            string
	StoredFilename string
}
