package internal

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// StorageBackend is an interface to abstract file storage implementations.
type StorageBackend interface {
	// Save stores the uploaded file to the destination path.
	Save(dstPath string, file multipart.File) error

	// Delete removes the file at the given path.
	Delete(path string) error

	// Exists checks if a file exists at the given path.
	Exists(path string) (bool, error)
}

type LocalStorage struct {
	BasePath string // e.g. "uploads"
}

var Backend StorageBackend

func InitLocalStorage(basePath string) {
	Backend = NewLocalStorage(basePath)
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{BasePath: basePath}
}

func (l *LocalStorage) Save(dstPath string, file multipart.File) error {
	fullPath := filepath.Join(l.BasePath, dstPath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	outFile, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	return err
}

func (l *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(l.BasePath, path)
	return os.Remove(fullPath)
}

func (l *LocalStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(l.BasePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
