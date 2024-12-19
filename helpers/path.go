package helpers

import (
	"os"
	"path/filepath"
)

// JoinWithProjectPath combines the input path with the project directory
func ProjectPath() (string, error) {
	// Get the current working directory (project directory)
	projectDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Construct the full path by joining the project directory and the relative path
	return projectDir, nil
}

// JoinWithProjectPath combines the input path with the project directory
func JoinWithProjectPath(relativePath ...string) string {
	// Get the current working directory (project directory)
	projectDir, err := ProjectPath()
	if err != nil {
		return ""
	}

	// Construct the full path by joining the project directory and the relative paths
	fullPath := filepath.Join(projectDir, filepath.Join(relativePath...))

	// Construct the full path by joining the project directory and the relative path
	return fullPath
}
