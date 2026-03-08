package model

import (
	"testing"
	"time"
)

func TestFileMetadata_TableName(t *testing.T) {
	f := FileMetadata{}
	if f.TableName() != "file_metadata" {
		t.Errorf("TableName() = %q, want file_metadata", f.TableName())
	}
}

func TestFileMetadata_IsPublic(t *testing.T) {
	f := &FileMetadata{Visibility: "pub"}
	if !f.IsPublic() {
		t.Error("IsPublic() should be true for pub")
	}
	if f.IsPrivate() {
		t.Error("IsPrivate() should be false for pub")
	}
}

func TestFileMetadata_IsPrivate(t *testing.T) {
	f := &FileMetadata{Visibility: "pri"}
	if !f.IsPrivate() {
		t.Error("IsPrivate() should be true for pri")
	}
	if f.IsPublic() {
		t.Error("IsPublic() should be false for pri")
	}
}

func TestFileMetadata_Fields(t *testing.T) {
	now := time.Now()
	f := &FileMetadata{
		ID:               1,
		StorageID:        "local:avatar@abc.jpg",
		OriginalFilename: "photo.jpg",
		StoredFilename:   "abc.jpg",
		FilePath:         "/storage/avatar/abc.jpg",
		URL:              "/api/files/storage/local:avatar@abc.jpg",
		FileSize:         1024,
		ContentType:      "image/jpeg",
		Driver:           "local",
		Visibility:       "pub",
		Category:         "avatar",
		BusinessID:       "user-1",
		BusinessType:     "profile",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if f.ID != 1 {
		t.Errorf("ID = %d, want 1", f.ID)
	}
	if f.StorageID != "local:avatar@abc.jpg" {
		t.Errorf("StorageID = %q", f.StorageID)
	}
	if f.FileSize != 1024 {
		t.Errorf("FileSize = %d, want 1024", f.FileSize)
	}
	if f.ContentType != "image/jpeg" {
		t.Errorf("ContentType = %q", f.ContentType)
	}
}
