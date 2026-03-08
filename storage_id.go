package storage

import (
	"fmt"
	"strings"
)

// StorageID 存储ID解析结果
// 格式：驱动:存储地@文件名，例如 local:avatar@c36722b7a4b116aa.jpg
type StorageID struct {
	Driver   string
	Storage  string
	Filename string
	RawID    string
}

// FullPath 返回 "存储地/文件名"
func (s *StorageID) FullPath() string {
	if s.Storage == "" {
		return s.Filename
	}
	return s.Storage + "/" + s.Filename
}

// ParseStorageID 解析存储ID字符串
func ParseStorageID(storageID string) (*StorageID, error) {
	if storageID == "" {
		return nil, fmt.Errorf("storage id is empty")
	}

	result := &StorageID{RawID: storageID}

	colonIdx := strings.Index(storageID, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("storage id missing driver separator ':'")
	}
	result.Driver = storageID[:colonIdx]

	rest := storageID[colonIdx+1:]
	atIdx := strings.Index(rest, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("storage id missing separator '@'")
	}

	result.Storage = rest[:atIdx]
	if result.Storage == "" {
		return nil, fmt.Errorf("storage id has empty storage identifier")
	}

	result.Filename = rest[atIdx+1:]
	if result.Filename == "" {
		return nil, fmt.Errorf("storage id has empty filename")
	}

	return result, nil
}

// BuildStorageID 构建存储ID字符串
func BuildStorageID(driver, storage, filename string) (string, error) {
	if driver == "" {
		return "", fmt.Errorf("driver name is required")
	}
	if storage == "" {
		return "", fmt.Errorf("storage identifier is required")
	}
	if filename == "" {
		return "", fmt.Errorf("filename is required")
	}
	return fmt.Sprintf("%s:%s@%s", driver, storage, filename), nil
}
