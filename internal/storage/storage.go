package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dzeleniak/icu/internal/types"
)

// Storage handles persistence of catalog data
type Storage struct {
	dataDir string
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Storage{
		dataDir: dataDir,
	}, nil
}

// catalogPath returns the path to the catalog file
func (s *Storage) catalogPath() string {
	return filepath.Join(s.dataDir, "catalog.json")
}

// Save persists the catalog to disk
func (s *Storage) Save(catalog *types.Catalog) error {
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal catalog: %w", err)
	}

	if err := os.WriteFile(s.catalogPath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write catalog file: %w", err)
	}

	return nil
}

// Load reads the catalog from disk
func (s *Storage) Load() (*types.Catalog, error) {
	data, err := os.ReadFile(s.catalogPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No catalog exists yet
		}
		return nil, fmt.Errorf("failed to read catalog file: %w", err)
	}

	var catalog types.Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to unmarshal catalog: %w", err)
	}

	return &catalog, nil
}

// Exists checks if a catalog file exists
func (s *Storage) Exists() bool {
	_, err := os.Stat(s.catalogPath())
	return err == nil
}
