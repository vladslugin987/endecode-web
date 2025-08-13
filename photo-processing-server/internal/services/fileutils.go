package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Supported file extensions (exact port from Kotlin)
var supportedExtensions = map[string]bool{
	"txt":  true,
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"mp4":  true,
	"avi":  true,
	"mov":  true,
	"mkv":  true,
}

// Video file extensions (exact port from Kotlin)
var videoExtensions = map[string]bool{
	"mp4": true,
	"avi": true,
	"mov": true,
	"mkv": true,
}

// Image file extensions
var imageExtensions = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
}

// CopyDirectory copies source directory to destination (exact port from Kotlin)
func CopyDirectory(source, destination string) error {
	logger := GetGlobalLogger()
	
	// Create destination directory
	err := os.MkdirAll(destination, 0755)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating directory %s: %v", destination, err)
		logger.Error(errMsg)
		return err
	}
	
	// Walk through source directory
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Calculate relative path
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		
		destPath := filepath.Join(destination, relPath)
		
		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// Copy file
			return copyFile(path, destPath, info.Mode())
		}
	})
	
	if err != nil {
		errMsg := fmt.Sprintf("Error copying directory from %s to %s: %v", source, destination, err)
		logger.Error(errMsg)
		return err
	}
	
	logger.Log(fmt.Sprintf("Directory copied: %s", filepath.Base(destination)))
	return nil
}

// copyFile copies a single file
func copyFile(src, dst string, mode os.FileMode) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	// Set file mode
	return os.Chmod(dst, mode)
}

// GetSupportedFiles returns all supported files in directory (exact port from Kotlin)
func GetSupportedFiles(directory string) ([]string, error) {
	logger := GetGlobalLogger()
	var supportedFiles []string
	
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error(fmt.Sprintf("Error accessing path %s: %v", path, err))
			return nil // Continue walking despite errors
		}
		
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			// Remove the dot from extension
			if len(ext) > 1 {
				ext = ext[1:]
			}
			
			if supportedExtensions[ext] {
				supportedFiles = append(supportedFiles, path)
			}
		}
		
		return nil
	})
	
	if err != nil {
		errMsg := fmt.Sprintf("Error getting files from directory %s: %v", directory, err)
		logger.Error(errMsg)
		return nil, err
	}
	
	return supportedFiles, nil
}

// CountFiles counts supported files in directory (exact port from Kotlin)
func CountFiles(directory string) (int, error) {
	logger := GetGlobalLogger()
	count := 0
	
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error(fmt.Sprintf("Error counting files in %s: %v", path, err))
			return nil // Continue walking despite errors
		}
		
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			// Remove the dot from extension
			if len(ext) > 1 {
				ext = ext[1:]
			}
			
			if supportedExtensions[ext] {
				count++
			}
		}
		
		return nil
	})
	
	if err != nil {
		logger.Error(fmt.Sprintf("Error counting files in directory %s: %v", directory, err))
		return 0, err
	}
	
	return count, nil
}

// IsImageFile checks if file is an image (exact port from Kotlin)
func IsImageFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	// Remove the dot from extension
	if len(ext) > 1 {
		ext = ext[1:]
	}
	return imageExtensions[ext]
}

// IsVideoFile checks if file is a video (exact port from Kotlin)
func IsVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	// Remove the dot from extension
	if len(ext) > 1 {
		ext = ext[1:]
	}
	return videoExtensions[ext]
}

// IsTextFile checks if file is a text file
func IsTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	// Remove the dot from extension
	if len(ext) > 1 {
		ext = ext[1:]
	}
	return ext == "txt"
}

// IsSupportedFile checks if file has supported extension
func IsSupportedFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	// Remove the dot from extension
	if len(ext) > 1 {
		ext = ext[1:]
	}
	return supportedExtensions[ext]
}

// GetFileSize returns file size in bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetDirectorySize calculates total size of all supported files in directory
func GetDirectorySize(directory string) (int64, error) {
	var totalSize int64
	
	files, err := GetSupportedFiles(directory)
	if err != nil {
		return 0, err
	}
	
	for _, file := range files {
		size, err := GetFileSize(file)
		if err != nil {
			continue // Skip files we can't read
		}
		totalSize += size
	}
	
	return totalSize, nil
}

// FormatFileSize formats bytes into human-readable string
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// EnsureDirectoryExists creates directory if it doesn't exist
func EnsureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0755)
}

// CleanupTempFiles removes temporary files
func CleanupTempFiles(tempDir string) error {
	logger := GetGlobalLogger()
	
	err := os.RemoveAll(tempDir)
	if err != nil {
		logger.Error(fmt.Sprintf("Error cleaning up temp files: %v", err))
		return err
	}
	
	logger.Log("Temp files cleaned up")
	return nil
}