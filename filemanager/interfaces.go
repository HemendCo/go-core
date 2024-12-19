package filemanager

import "os"

// FileManagerInterface defines the methods for the FileManager
type FileManagerInterface interface {
	ListFilesInDirectory(dir string) ([]os.FileInfo, error)
	ListFilesRecursively(dir string) ([]os.FileInfo, error)
	WriteFile(filePath string, content interface{}) error
	ReadFile(filePath string, output interface{}) error
	Has(path string) (bool, error)
	RemoveFileOrDirectory(path string) error
	Rename(oldPath, newPath string) error
	ReadUserInput() (string, error)
}
