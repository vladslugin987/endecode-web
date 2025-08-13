package services

import (
    "fmt"
    "image"
    "image/color"
    "path/filepath"

    "gocv.io/x/gocv"
)

// TextPosition represents position for visible watermarks (exact port from Kotlin enum)
type TextPosition int

const (
	TopLeft TextPosition = iota
	TopRight
	Center
	BottomLeft
	BottomRight
)

// OpenCV parameters (exact port from Kotlin constants)
const (
	FONT_FACE    = gocv.FontHersheySimplex
	FONT_SCALE   = 0.4
	THICKNESS    = 1
	ALPHA        = 0.5
	PADDING      = 5
)

var (
    WHITE_COLOR = color.RGBA{R: 255, G: 255, B: 255, A: 255}
    isOpenCVInitialized = false
)

// initializeOpenCV initializes OpenCV (port from Kotlin init block)
func initializeOpenCV() error {
	if isOpenCVInitialized {
		return nil
	}
	
	logger := GetGlobalLogger()
	
	// Check OpenCV version to ensure it's working
	version := gocv.OpenCVVersion()
	if version == "" {
		err := fmt.Errorf("failed to initialize OpenCV")
		logger.Error(fmt.Sprintf("Failed to initialize OpenCV: %v", err))
		return err
	}
	
	logger.Log(fmt.Sprintf("OpenCV initialized successfully (version: %s)", version))
	isOpenCVInitialized = true
	return nil
}

// AddTextToImage adds semi-transparent visible watermark to image (exact port from Kotlin)
func AddTextToImage(imagePath string, text string, position TextPosition) error {
	logger := GetGlobalLogger()
	
	// Initialize OpenCV if needed
	err := initializeOpenCV()
	if err != nil {
		return err
	}
	
	// Load image
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		msg := fmt.Sprintf("Failed to load image: %s", filepath.Base(imagePath))
		logger.Log(msg)
		return fmt.Errorf(msg)
	}
	defer img.Close()
	
	// Get text size for positioning
	textSize := gocv.GetTextSize(text, FONT_FACE, FONT_SCALE, THICKNESS)
	
	// Calculate text position based on enum value (exact port from Kotlin when expression)
	var textPoint image.Point
	imgSize := img.Size()
	
	switch position {
	case BottomRight:
		textPoint = image.Point{
			X: imgSize[1] - textSize.X - PADDING,
			Y: imgSize[0] - PADDING,
		}
	case BottomLeft:
		textPoint = image.Point{
			X: PADDING,
			Y: imgSize[0] - PADDING,
		}
	case TopRight:
		textPoint = image.Point{
			X: imgSize[1] - textSize.X - PADDING,
			Y: textSize.Y + PADDING,
		}
	case TopLeft:
		textPoint = image.Point{
			X: PADDING,
			Y: textSize.Y + PADDING,
		}
	case Center:
		textPoint = image.Point{
			X: (imgSize[1] - textSize.X) / 2,
			Y: (imgSize[0] + textSize.Y) / 2,
		}
	}
	
	// Create overlay Mat for alpha blending (draw only the text)
	overlay := gocv.NewMatWithSize(imgSize[0], imgSize[1], gocv.MatTypeCV8UC3)
	defer overlay.Close()
	// Fill overlay with zeros (no darkening outside text)
	overlay.SetTo(gocv.NewScalar(0, 0, 0, 0))
	
	// Draw text on overlay in white
	gocv.PutText(&overlay, text, textPoint, FONT_FACE, FONT_SCALE, WHITE_COLOR, THICKNESS)
	
	// Blend: keep base image weight 1.0, add only overlay with ALPHA
	gocv.AddWeighted(img, 1.0, overlay, ALPHA, 0.0, &img)
	
	// Save image
	success := gocv.IMWrite(imagePath, img)
	if success {
		logger.Log(fmt.Sprintf("Added text to %s", filepath.Base(imagePath)))
	} else {
		errMsg := fmt.Sprintf("Failed to save image %s", filepath.Base(imagePath))
		logger.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	
	return nil
}

// AddTextToImageAtPosition is a convenience function with default position
func AddTextToImageAtPosition(imagePath string, text string) error {
	return AddTextToImage(imagePath, text, BottomRight)
}

// ProcessImageWithVisibleWatermark processes image file with visible watermark
func ProcessImageWithVisibleWatermark(filePath string, watermarkText string, photoNumber int) error {
	if !IsImageFile(filePath) {
		return fmt.Errorf("file %s is not a supported image file", filePath)
	}
	
	// Extract file number from filename
	fileNumber := extractFileNumber(filepath.Base(filePath))
	if fileNumber == nil || *fileNumber != photoNumber {
		return nil // Skip this file, number doesn't match
	}
	
	return AddTextToImage(filePath, watermarkText, BottomRight)
}

// extractFileNumber extracts number from filename (exact port from Kotlin regex)
func extractFileNumber(filename string) *int {
	// This matches the Kotlin regex: """.*?(\d+).*""".toRegex()
	// We'll implement a simple number extraction
	var num int
	var found bool
	
	// Find the first sequence of digits in the filename
	for i := 0; i < len(filename); i++ {
		if filename[i] >= '0' && filename[i] <= '9' {
			num = 0
			j := i
			for j < len(filename) && filename[j] >= '0' && filename[j] <= '9' {
				num = num*10 + int(filename[j]-'0')
				j++
			}
			found = true
			break
		}
	}
	
	if found {
		return &num
	}
	return nil
}

// TestImageProcessing tests image watermarking functionality
func TestImageProcessing() {
	logger := GetGlobalLogger()
	logger.Log("Testing image processing...")
	
	// Test OpenCV initialization
	err := initializeOpenCV()
	if err != nil {
		logger.Error(fmt.Sprintf("OpenCV initialization failed: %v", err))
		return
	}
	
	logger.Log("✓ OpenCV initialization test passed")
	
	// Test file number extraction
	testCases := []struct {
		filename string
		expected *int
	}{
		{"Photo-001.jpg", intPtr(1)},
		{"Photo-0011.jpg", intPtr(11)},
		{"image-101.png", intPtr(101)},
		{"test.jpg", nil},
	}
	
	for _, tc := range testCases {
		result := extractFileNumber(tc.filename)
		if (tc.expected == nil && result == nil) ||
		   (tc.expected != nil && result != nil && *tc.expected == *result) {
			logger.Log(fmt.Sprintf("✓ File number extraction: %s → %v", tc.filename, result))
		} else {
			logger.Error(fmt.Sprintf("✗ File number extraction failed: %s → %v (expected %v)", 
				tc.filename, result, tc.expected))
		}
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// GetSupportedImageExtensions returns supported image extensions
func GetSupportedImageExtensions() []string {
	return []string{"jpg", "jpeg", "png"}
}

// ValidateImageFile checks if image file can be processed
func ValidateImageFile(imagePath string) error {
	if !IsImageFile(imagePath) {
		return fmt.Errorf("file %s is not a supported image format", imagePath)
	}
	
	// Try to load image to validate
	err := initializeOpenCV()
	if err != nil {
		return err
	}
	
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	defer img.Close()
	
	if img.Empty() {
		return fmt.Errorf("cannot read image file %s", imagePath)
	}
	
	return nil
}