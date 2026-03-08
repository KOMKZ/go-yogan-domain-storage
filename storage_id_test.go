package storage

import "testing"

func TestParseStorageID_Valid(t *testing.T) {
	sid, err := ParseStorageID("local:avatar@c36722b7a4b116aa.jpg")
	if err != nil {
		t.Fatalf("ParseStorageID() err = %v", err)
	}
	if sid.Driver != "local" {
		t.Errorf("Driver = %q, want local", sid.Driver)
	}
	if sid.Storage != "avatar" {
		t.Errorf("Storage = %q, want avatar", sid.Storage)
	}
	if sid.Filename != "c36722b7a4b116aa.jpg" {
		t.Errorf("Filename = %q, want c36722b7a4b116aa.jpg", sid.Filename)
	}
	if sid.RawID != "local:avatar@c36722b7a4b116aa.jpg" {
		t.Errorf("RawID = %q", sid.RawID)
	}
}

func TestParseStorageID_OSSFormat(t *testing.T) {
	sid, err := ParseStorageID("oss:images@abc123.png")
	if err != nil {
		t.Fatalf("ParseStorageID() err = %v", err)
	}
	if sid.Driver != "oss" {
		t.Errorf("Driver = %q, want oss", sid.Driver)
	}
	if sid.Storage != "images" {
		t.Errorf("Storage = %q, want images", sid.Storage)
	}
}

func TestParseStorageID_Empty(t *testing.T) {
	_, err := ParseStorageID("")
	if err == nil {
		t.Error("ParseStorageID('') should fail")
	}
}

func TestParseStorageID_NoColon(t *testing.T) {
	_, err := ParseStorageID("localavatar@file.jpg")
	if err == nil {
		t.Error("ParseStorageID without ':' should fail")
	}
}

func TestParseStorageID_NoAt(t *testing.T) {
	_, err := ParseStorageID("local:avatarfile.jpg")
	if err == nil {
		t.Error("ParseStorageID without '@' should fail")
	}
}

func TestParseStorageID_EmptyStorage(t *testing.T) {
	_, err := ParseStorageID("local:@file.jpg")
	if err == nil {
		t.Error("ParseStorageID with empty storage should fail")
	}
}

func TestParseStorageID_EmptyFilename(t *testing.T) {
	_, err := ParseStorageID("local:avatar@")
	if err == nil {
		t.Error("ParseStorageID with empty filename should fail")
	}
}

func TestBuildStorageID_Valid(t *testing.T) {
	id, err := BuildStorageID("local", "avatar", "file.jpg")
	if err != nil {
		t.Fatalf("BuildStorageID() err = %v", err)
	}
	if id != "local:avatar@file.jpg" {
		t.Errorf("id = %q, want local:avatar@file.jpg", id)
	}
}

func TestBuildStorageID_EmptyDriver(t *testing.T) {
	_, err := BuildStorageID("", "avatar", "file.jpg")
	if err == nil {
		t.Error("BuildStorageID with empty driver should fail")
	}
}

func TestBuildStorageID_EmptyStorage(t *testing.T) {
	_, err := BuildStorageID("local", "", "file.jpg")
	if err == nil {
		t.Error("BuildStorageID with empty storage should fail")
	}
}

func TestBuildStorageID_EmptyFilename(t *testing.T) {
	_, err := BuildStorageID("local", "avatar", "")
	if err == nil {
		t.Error("BuildStorageID with empty filename should fail")
	}
}

func TestStorageID_FullPath(t *testing.T) {
	sid := &StorageID{Storage: "avatar", Filename: "file.jpg"}
	if sid.FullPath() != "avatar/file.jpg" {
		t.Errorf("FullPath() = %q, want avatar/file.jpg", sid.FullPath())
	}
}

func TestStorageID_FullPath_EmptyStorage(t *testing.T) {
	sid := &StorageID{Filename: "file.jpg"}
	if sid.FullPath() != "file.jpg" {
		t.Errorf("FullPath() = %q, want file.jpg", sid.FullPath())
	}
}

func TestBuildAndParse_Roundtrip(t *testing.T) {
	id, err := BuildStorageID("oss", "image", "test.png")
	if err != nil {
		t.Fatal(err)
	}
	sid, err := ParseStorageID(id)
	if err != nil {
		t.Fatal(err)
	}
	if sid.Driver != "oss" || sid.Storage != "image" || sid.Filename != "test.png" {
		t.Errorf("roundtrip mismatch: %+v", sid)
	}
}
