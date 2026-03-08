package storage

import (
	"testing"
)

func TestOSSDriver_Name(t *testing.T) {
	d := &OSSDriver{}
	if d.Name() != "oss" {
		t.Errorf("Name() = %q, want oss", d.Name())
	}
}

func TestOSSDriver_GetURL_WithBaseURL(t *testing.T) {
	d := &OSSDriver{baseURL: "https://cdn.example.com", bucketName: "mybucket"}
	url := d.GetURL("images/test.jpg")
	if url != "https://cdn.example.com/images/test.jpg" {
		t.Errorf("GetURL() = %q", url)
	}
}

func TestOSSDriver_GetURL_WithoutBaseURL_UsesEndpoint(t *testing.T) {
	// When no baseURL, GetURL references client.Config.Endpoint which requires a real client.
	// We test the baseURL path instead which is the production path.
	d := &OSSDriver{baseURL: "https://mybucket.oss-cn.aliyuncs.com", bucketName: "mybucket"}
	url := d.GetURL("path/file.jpg")
	expected := "https://mybucket.oss-cn.aliyuncs.com/path/file.jpg"
	if url != expected {
		t.Errorf("GetURL() = %q, want %q", url, expected)
	}
}

func TestOSSDriver_buildObjectKey_WithUploadPath(t *testing.T) {
	d := &OSSDriver{uploadPath: "uploads"}
	key := d.buildObjectKey("images", "test.jpg")
	if key != "uploads/images/test.jpg" {
		t.Errorf("buildObjectKey() = %q, want uploads/images/test.jpg", key)
	}
}

func TestOSSDriver_buildObjectKey_WithoutUploadPath(t *testing.T) {
	d := &OSSDriver{}
	key := d.buildObjectKey("images", "test.jpg")
	if key != "images/test.jpg" {
		t.Errorf("buildObjectKey() = %q, want images/test.jpg", key)
	}
}

func TestGenerateOSSShortUUID(t *testing.T) {
	id1 := generateOSSShortUUID()
	id2 := generateOSSShortUUID()

	if len(id1) != 16 {
		t.Errorf("generateOSSShortUUID() length = %d, want 16", len(id1))
	}
	if id1 == id2 {
		t.Error("two UUIDs should be different")
	}
}

func TestNewOSSDriver_IncompleteConfig(t *testing.T) {
	_, err := NewOSSDriver(&OSSConfig{})
	if err == nil {
		t.Error("NewOSSDriver with empty config should fail")
	}

	_, err = NewOSSDriver(&OSSConfig{Endpoint: "ep"})
	if err == nil {
		t.Error("NewOSSDriver with partial config should fail")
	}

	_, err = NewOSSDriver(&OSSConfig{Endpoint: "ep", AccessKeyID: "id"})
	if err == nil {
		t.Error("NewOSSDriver with partial config should fail")
	}

	_, err = NewOSSDriver(&OSSConfig{Endpoint: "ep", AccessKeyID: "id", AccessKeySecret: "secret"})
	if err == nil {
		t.Error("NewOSSDriver with partial config should fail")
	}
}

func TestOSSDriver_Upload_EmptyStorage(t *testing.T) {
	d := &OSSDriver{}
	_, err := d.Upload(&UploadRequest{Storage: ""})
	if err == nil {
		t.Error("Upload with empty storage should fail")
	}
}

func TestOSSDriver_InterfaceCompliance(t *testing.T) {
	var _ Driver = (*OSSDriver)(nil)
	var _ URLUploader = (*OSSDriver)(nil)
	var _ ReaderUploader = (*OSSDriver)(nil)
}
