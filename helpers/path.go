package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

// JoinWithProjectPath combines the input path with the project directory
func ProjectPath() *string {
	// Get the current working directory (project directory)
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting project path: %v\n", err)
		return nil
	}

	// Construct the full path by joining the project directory and the relative path
	return &projectDir
}

// JoinWithProjectPath combines the input path with the project directory
func JoinWithProjectPath(relativePath ...string) string {
	// Ensure that the relativePath is not empty
	if len(relativePath) == 0 {
		fmt.Println("Warning: The input path is empty.")
		return *ProjectPath() // return just the project path if no relative path is provided
	}

	// Construct the full path by joining the project directory and the relative paths
	fullPath := filepath.Join(*ProjectPath(), filepath.Join(relativePath...))

	// Construct the full path by joining the project directory and the relative path
	return fullPath
}
