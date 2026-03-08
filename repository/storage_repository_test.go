package repository

import (
	"testing"
)

func TestStorageRepositoryInterface(t *testing.T) {
	// Ensure the interface is defined and compile-check
	var _ StorageRepository = (StorageRepository)(nil)
}
