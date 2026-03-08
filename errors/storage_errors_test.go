package errors

import (
	"testing"
)

func TestErrDatabaseError(t *testing.T) {
	if ErrDatabaseError == nil {
		t.Fatal("ErrDatabaseError should not be nil")
	}
}

func TestErrNotFound(t *testing.T) {
	if ErrNotFound == nil {
		t.Fatal("ErrNotFound should not be nil")
	}
}

func TestErrBadRequest(t *testing.T) {
	if ErrBadRequest == nil {
		t.Fatal("ErrBadRequest should not be nil")
	}
}

func TestErrUploadFailed(t *testing.T) {
	if ErrUploadFailed == nil {
		t.Fatal("ErrUploadFailed should not be nil")
	}
}

func TestErrReadFailed(t *testing.T) {
	if ErrReadFailed == nil {
		t.Fatal("ErrReadFailed should not be nil")
	}
}

func TestModuleCode(t *testing.T) {
	if ModuleStorage != 21 {
		t.Errorf("ModuleStorage = %d, want 21", ModuleStorage)
	}
}
