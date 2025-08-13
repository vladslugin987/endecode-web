package services

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	MAX_WATERMARK_LENGTH = 100
)

// Binary watermark markers (exact port from Kotlin)
var (
	WATERMARK_START = []byte("<<==")
	WATERMARK_END   = []byte("==>>")
)

// WatermarkInfo represents information about a found watermark (exact port from Kotlin)
type WatermarkInfo struct {
	StartPosition int
	EndPosition   int
	Content       []byte
}

// RemoveWatermarks removes invisible watermarks from all files in directory (exact port from Kotlin)
func RemoveWatermarks(directory string, progress func(float32)) error {
	logger := GetGlobalLogger()
	
	// Get all supported media files
	var files []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}
		
		if !info.IsDir() && (IsImageFile(path) || IsVideoFile(path)) {
			files = append(files, path)
		}
		return nil
	})
	
	if err != nil {
		logger.Error(fmt.Sprintf("Error during watermark removal process: %v", err))
		return err
	}
	
	processedFiles := 0.0
	totalFiles := float32(len(files))
	
	for _, file := range files {
		removed, err := removeWatermarkFromFile(file)
		if err != nil {
			logger.Error(fmt.Sprintf("Error removing watermark from %s: %v", filepath.Base(file), err))
		} else if removed {
			logger.Log(fmt.Sprintf("Watermark removed from %s", filepath.Base(file)))
		} else {
			logger.Log(fmt.Sprintf("No watermark found in %s", filepath.Base(file)))
		}
		
		processedFiles++
		if progress != nil {
			progress(float32(processedFiles) / totalFiles)
		}
	}
	
	logger.Log("Watermark removal completed")
	return nil
}

// ExtractWatermarkText extracts encoded text from watermark (exact port from Kotlin)
func ExtractWatermarkText(filePath string) (string, error) {
	tailData, _, err := readWatermarkData(filePath)
	if err != nil {
		return "", err
	}
	
	if tailData == nil {
		return "", nil
	}
	
	watermarkInfo := findWatermark(tailData, true)
	if watermarkInfo == nil {
		return "", nil
	}
	
	if watermarkInfo.Content != nil {
		result := string(watermarkInfo.Content)
		logger := GetGlobalLogger()
		logger.Log(fmt.Sprintf("Found watermark in %s: %s", filepath.Base(filePath), result))
		return result, nil
	}
	
	return "", nil
}

// HasWatermark checks if file has a watermark (exact port from Kotlin)
func HasWatermark(filePath string) (bool, error) {
	tailData, _, err := readWatermarkData(filePath)
	if err != nil {
		return false, err
	}
	
	if tailData == nil {
		return false, nil
	}
	
	watermarkInfo := findWatermark(tailData, false)
	return watermarkInfo != nil, nil
}

// AddBinaryWatermark adds encoded binary watermark to file (exact port from Kotlin)
func AddBinaryWatermark(filePath string, encodedText string) error {
	logger := GetGlobalLogger()
	
	// Check if watermark already exists
	hasWM, err := HasWatermark(filePath)
	if err != nil {
		return err
	}
	
    if hasWM {
		logger.Log(fmt.Sprintf("%s: Already has watermark", filepath.Base(filePath)))
		return nil
	}
	
	// Open file for appending
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("Error adding watermark to %s: %v", filepath.Base(filePath), err))
		return err
	}
	defer file.Close()
	
	// Create watermark: WATERMARK_START + encodedText + WATERMARK_END
    watermark := append(WATERMARK_START, []byte(encodedText)...)
    watermark = append(watermark, WATERMARK_END...)
	
	_, err = file.Write(watermark)
	if err != nil {
		logger.Error(fmt.Sprintf("Error adding watermark to %s: %v", filepath.Base(filePath), err))
		return err
	}
	
    logger.Log(fmt.Sprintf("%s: Watermark added successfully", filepath.Base(filePath)))
	return nil
}

// removeWatermarkFromFile removes watermark from specific file (exact port from Kotlin)
func removeWatermarkFromFile(filePath string) (bool, error) {
	tailData, fileSize, err := readWatermarkData(filePath)
	if err != nil {
		return false, err
	}
	
	if tailData == nil {
		return false, nil
	}
	
	watermarkInfo := findWatermark(tailData, false)
	if watermarkInfo == nil {
		return false, nil
	}
	
	// Calculate watermark position in file
	watermarkPosition := fileSize - int64(len(tailData)-watermarkInfo.StartPosition)
	
	// Truncate file at watermark position
	err = os.Truncate(filePath, watermarkPosition)
	if err != nil {
		logger := GetGlobalLogger()
		logger.Error(fmt.Sprintf("Error removing watermark from %s: %v", filepath.Base(filePath), err))
		return false, err
	}
	
	return true, nil
}

// readWatermarkData reads last MAX_WATERMARK_LENGTH bytes from file (exact port from Kotlin)
func readWatermarkData(filePath string) ([]byte, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger := GetGlobalLogger()
		logger.Error(fmt.Sprintf("Error reading watermark data from %s: %v", filepath.Base(filePath), err))
		return nil, 0, err
	}
	defer file.Close()
	
	// Get file size
	info, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	
	fileSize := info.Size()
	if fileSize < MAX_WATERMARK_LENGTH {
		return nil, fileSize, nil
	}
	
	// Read last MAX_WATERMARK_LENGTH bytes
	readLength := int64(MAX_WATERMARK_LENGTH)
	if readLength > fileSize {
		readLength = fileSize
	}
	
	_, err = file.Seek(fileSize-readLength, 0)
	if err != nil {
		return nil, fileSize, err
	}
	
	tailData := make([]byte, readLength)
	_, err = file.Read(tailData)
	if err != nil {
		return nil, fileSize, err
	}
	
	return tailData, fileSize, nil
}

// findWatermark looks for watermark signatures in byte array (exact port from Kotlin)
func findWatermark(data []byte, includeContent bool) *WatermarkInfo {
	// Find start position (search from end to beginning)
	startPosition := findBytesReverse(data, WATERMARK_START)
	if startPosition == -1 {
		return nil
	}
	
	// Find end position (search from start position forward)
	endPosition := findBytes(data, WATERMARK_END, startPosition)
	if endPosition == -1 || endPosition <= startPosition {
		return nil
	}
	
	var content []byte
	if includeContent {
		contentStart := startPosition + len(WATERMARK_START)
		content = data[contentStart:endPosition]
	}
	
	return &WatermarkInfo{
		StartPosition: startPosition,
		EndPosition:   endPosition,
		Content:       content,
	}
}

// findBytes searches for byte pattern in data array (exact port from Kotlin)
func findBytes(data []byte, pattern []byte, startFrom int) int {
	dataLen := len(data)
	patternLen := len(pattern)
	
	for i := startFrom; i <= dataLen-patternLen; i++ {
		if matchesAt(data, i, pattern) {
			return i
		}
	}
	return -1
}

// findBytesReverse searches for byte pattern from end to start (exact port from Kotlin)
func findBytesReverse(data []byte, pattern []byte) int {
	dataLen := len(data)
	patternLen := len(pattern)
	
	for i := dataLen - patternLen; i >= 0; i-- {
		if matchesAt(data, i, pattern) {
			return i
		}
	}
	return -1
}

// matchesAt checks if byte array matches pattern at given position (exact port from Kotlin)
func matchesAt(data []byte, pos int, pattern []byte) bool {
	if pos+len(pattern) > len(data) {
		return false
	}
	
	for i := 0; i < len(pattern); i++ {
		if data[pos+i] != pattern[i] {
			return false
		}
	}
	return true
}

// TestWatermarkOperations tests watermark functionality
func TestWatermarkOperations(testFilePath string) {
	logger := GetGlobalLogger()
	logger.Log("Testing watermark operations...")
	
	// Test adding watermark
    encodedText := EncodeText("Test Watermark 001")
    err := AddBinaryWatermark(testFilePath, encodedText)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to add watermark: %v", err))
		return
	}
	
	// Test checking watermark presence
	hasWM, err := HasWatermark(testFilePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to check watermark: %v", err))
		return
	}
	
	if hasWM {
		logger.Log("✓ Watermark presence test passed")
	} else {
		logger.Error("✗ Watermark presence test failed")
	}
	
	// Test extracting watermark
	extractedText, err := ExtractWatermarkText(testFilePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to extract watermark: %v", err))
		return
	}
	
	if extractedText == encodedText {
		logger.Log("✓ Watermark extraction test passed")
	} else {
		logger.Error(fmt.Sprintf("✗ Watermark extraction failed: got '%s', expected '%s'", extractedText, encodedText))
	}
}