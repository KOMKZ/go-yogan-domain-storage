package storage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalDriver_Name(t *testing.T) {
	d := NewLocalDriver("/tmp", "/prefix/")
	if d.Name() != "local" {
		t.Errorf("Name() = %q, want local", d.Name())
	}
}

func TestLocalDriver_UploadAndRead(t *testing.T) {
	dir := t.TempDir()
	d := NewLocalDriver(dir, "/files/")

	content := "hello storage"
	result, err := d.Upload(&UploadRequest{
		Reader:       strings.NewReader(content),
		OriginalName: "test.txt",
		FileSize:     int64(len(content)),
		ContentType:  "text/plain",
		Storage:      "docs",
	})
	if err != nil {
		t.Fatalf("Upload() err = %v", err)
	}

	if result.StorageID == "" {
		t.Error("StorageID should not be empty")
	}
	if result.StoredFilename == "" {
		t.Error("StoredFilename should not be empty")
	}
	if !strings.HasSuffix(result.StoredFilename, ".txt") {
		t.Errorf("StoredFilename = %q, should end with .txt", result.StoredFilename)
	}

	if !d.Exists(result.FilePath) {
		t.Error("file should exist after upload")
	}

	reader, err := d.GetReader(result.FilePath)
	if err != nil {
		t.Fatalf("GetReader() err = %v", err)
	}
	defer reader.Close()

	data, _ := io.ReadAll(reader)
	if string(data) != content {
		t.Errorf("read content = %q, want %q", string(data), content)
	}
}

func TestLocalDriver_Upload_EmptyStorage(t *testing.T) {
	d := NewLocalDriver(t.TempDir(), "/files/")
	_, err := d.Upload(&UploadRequest{
		Reader:       strings.NewReader("data"),
		OriginalName: "test.txt",
		Storage:      "",
	})
	if err == nil {
		t.Error("Upload with empty storage should fail")
	}
}

func TestLocalDriver_Delete(t *testing.T) {
	dir := t.TempDir()
	d := NewLocalDriver(dir, "/files/")

	result, _ := d.Upload(&UploadRequest{
		Reader:       strings.NewReader("data"),
		OriginalName: "del.txt",
		Storage:      "tmp",
	})

	if err := d.Delete(result.FilePath); err != nil {
		t.Fatalf("Delete() err = %v", err)
	}

	if d.Exists(result.FilePath) {
		t.Error("file should not exist after delete")
	}
}

func TestLocalDriver_Delete_NonExistent(t *testing.T) {
	d := NewLocalDriver(t.TempDir(), "/files/")
	if err := d.Delete("/nonexistent/file.txt"); err != nil {
		t.Errorf("Delete non-existent should not error, got %v", err)
	}
}

func TestLocalDriver_GetURL(t *testing.T) {
	d := NewLocalDriver("/tmp", "/api/files/")
	url := d.GetURL("docs/test.txt")
	if url != "/api/files/docs/test.txt" {
		t.Errorf("GetURL() = %q", url)
	}
}

func TestLocalDriver_Exists_False(t *testing.T) {
	d := NewLocalDriver(t.TempDir(), "/files/")
	if d.Exists("/nonexistent/file") {
		t.Error("should not exist")
	}
}

func TestLocalDriver_GetReader_NonExistent(t *testing.T) {
	d := NewLocalDriver(t.TempDir(), "/files/")
	_, err := d.GetReader("/nonexistent/file")
	if err == nil {
		t.Error("GetReader on non-existent file should fail")
	}
}

func TestLocalDriver_Upload_BinaryContent(t *testing.T) {
	dir := t.TempDir()
	d := NewLocalDriver(dir, "/files/")

	binData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	result, err := d.Upload(&UploadRequest{
		Reader:       bytes.NewReader(binData),
		OriginalName: "img.png",
		FileSize:     int64(len(binData)),
		ContentType:  "image/png",
		Storage:      "images",
	})
	if err != nil {
		t.Fatalf("Upload() err = %v", err)
	}

	stored, _ := os.ReadFile(result.FilePath)
	if !bytes.Equal(stored, binData) {
		t.Error("stored content does not match")
	}
}

func TestLocalDriver_StorageID_Format(t *testing.T) {
	dir := t.TempDir()
	d := NewLocalDriver(dir, "/files/")

	result, _ := d.Upload(&UploadRequest{
		Reader:       strings.NewReader("x"),
		OriginalName: "test.jpg",
		Storage:      "avatar",
	})

	sid, err := ParseStorageID(result.StorageID)
	if err != nil {
		t.Fatalf("ParseStorageID() err = %v", err)
	}
	if sid.Driver != "local" {
		t.Errorf("Driver = %q, want local", sid.Driver)
	}
	if sid.Storage != "avatar" {
		t.Errorf("Storage = %q, want avatar", sid.Storage)
	}
	if !strings.HasSuffix(sid.Filename, ".jpg") {
		t.Errorf("Filename = %q, should end with .jpg", sid.Filename)
	}
}

func TestLocalDriver_Upload_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	d := NewLocalDriver(dir, "/files/")

	_, err := d.Upload(&UploadRequest{
		Reader:       strings.NewReader("data"),
		OriginalName: "test.txt",
		Storage:      "nested",
	})
	if err != nil {
		t.Fatalf("Upload() err = %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "nested"))
	if err != nil {
		t.Fatalf("directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}
}

func TestLocalDriver_Upload_ReadError(t *testing.T) {
	d := NewLocalDriver(t.TempDir(), "/files/")
	errReader := &failReader{err: io.ErrUnexpectedEOF}

	_, err := d.Upload(&UploadRequest{
		Reader:       errReader,
		OriginalName: "fail.txt",
		FileSize:     100,
		Storage:      "test",
	})
	if err == nil {
		t.Error("Upload with failing reader should fail")
	}
}

type failReader struct {
	err error
}

func (r *failReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func TestLocalDriver_Upload_BadBasePath(t *testing.T) {
	d := NewLocalDriver("/dev/null/impossible", "/files/")
	_, err := d.Upload(&UploadRequest{
		Reader:       strings.NewReader("data"),
		OriginalName: "test.txt",
		Storage:      "test",
	})
	if err == nil {
		t.Error("Upload to impossible path should fail")
	}
}

// compile-time check
var _ Driver = (*LocalDriver)(nil)
