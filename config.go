package storage

// Config 存储配置
type Config struct {
	Driver       string                 `mapstructure:"driver"`
	Local        LocalConfig            `mapstructure:"local"`
	OSS          OSSConfig              `mapstructure:"oss"`
	StorageTypes map[string]StorageType `mapstructure:"storage_types"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	BasePath  string `mapstructure:"base_path"`
	URLPrefix string `mapstructure:"url_prefix"`
}

// StorageType 存储类型配置
type StorageType struct {
	Visibility   string   `mapstructure:"visibility"`
	Path         string   `mapstructure:"path"`
	MaxSize      int64    `mapstructure:"max_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Driver: "local",
		Local: LocalConfig{
			BasePath:  "./storage",
			URLPrefix: "/api/admin/files/storage/",
		},
		StorageTypes: map[string]StorageType{
			"avatar": {
				Visibility:   "pub",
				Path:         "avatars",
				MaxSize:      2,
				AllowedTypes: []string{".jpg", ".jpeg", ".png", ".gif"},
			},
			"document": {
				Visibility:   "pri",
				Path:         "documents",
				MaxSize:      10,
				AllowedTypes: []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"},
			},
			"image": {
				Visibility:   "pub",
				Path:         "images",
				MaxSize:      5,
				AllowedTypes: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"},
			},
			"export": {
				Visibility:   "pri",
				Path:         "exports",
				MaxSize:      50,
				AllowedTypes: []string{".csv", ".xlsx"},
			},
		},
	}
}

// GetStorageType 获取存储类型配置，不存在返回 nil
func (c *Config) GetStorageType(storage string) *StorageType {
	if st, ok := c.StorageTypes[storage]; ok {
		return &st
	}
	return nil
}
