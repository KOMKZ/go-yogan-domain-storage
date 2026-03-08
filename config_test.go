package storage

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Driver != "local" {
		t.Errorf("Driver = %q, want local", cfg.Driver)
	}
	if cfg.Local.BasePath != "./storage" {
		t.Errorf("BasePath = %q, want ./storage", cfg.Local.BasePath)
	}
	if len(cfg.StorageTypes) != 3 {
		t.Errorf("len(StorageTypes) = %d, want 3", len(cfg.StorageTypes))
	}
}

func TestGetStorageType_Exists(t *testing.T) {
	cfg := DefaultConfig()
	st := cfg.GetStorageType("avatar")
	if st == nil {
		t.Fatal("GetStorageType(avatar) returned nil")
	}
	if st.Visibility != "pub" {
		t.Errorf("Visibility = %q, want pub", st.Visibility)
	}
	if st.MaxSize != 2 {
		t.Errorf("MaxSize = %d, want 2", st.MaxSize)
	}
}

func TestGetStorageType_NotExists(t *testing.T) {
	cfg := DefaultConfig()
	st := cfg.GetStorageType("nonexistent")
	if st != nil {
		t.Error("GetStorageType(nonexistent) should return nil")
	}
}

func TestGetStorageType_Document(t *testing.T) {
	cfg := DefaultConfig()
	st := cfg.GetStorageType("document")
	if st == nil {
		t.Fatal("GetStorageType(document) returned nil")
	}
	if st.Visibility != "pri" {
		t.Errorf("Visibility = %q, want pri", st.Visibility)
	}
	if st.MaxSize != 10 {
		t.Errorf("MaxSize = %d, want 10", st.MaxSize)
	}
}

func TestGetStorageType_Image(t *testing.T) {
	cfg := DefaultConfig()
	st := cfg.GetStorageType("image")
	if st == nil {
		t.Fatal("GetStorageType(image) returned nil")
	}
	if len(st.AllowedTypes) == 0 {
		t.Error("image AllowedTypes should not be empty")
	}
}
