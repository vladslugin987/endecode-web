package services

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	SHIFT                = 7
	WATERMARK_PREFIX     = "<<=="
	WATERMARK_SUFFIX     = "==>>"
	OLD_WATERMARK_PREFIX = "*/"
)

// EncodeText applies Caesar cipher with shift=7 (exact port from Kotlin)
func EncodeText(text string) string {
	var result strings.Builder
	
	for _, char := range text {
		switch {
		case unicode.IsUpper(char):
			// Uppercase letters: (char - 'A' + SHIFT) % 26 + 'A'
			index := int(char - 'A')
			shifted := (index + SHIFT) % 26
			newChar := rune('A' + shifted)
			result.WriteRune(newChar)
			
		case unicode.IsLower(char):
			// Lowercase letters: (char - 'a' + SHIFT) % 26 + 'a'  
			index := int(char - 'a')
			shifted := (index + SHIFT) % 26
			newChar := rune('a' + shifted)
			result.WriteRune(newChar)
			
		case unicode.IsDigit(char):
			// Digits: (digit + SHIFT) % 10
			digit := int(char - '0')
			shifted := (digit + SHIFT) % 10
			result.WriteRune(rune('0' + shifted))
			
		default:
			// Other characters remain unchanged
			result.WriteRune(char)
		}
	}
	
	return result.String()
}

// DecodeText reverses Caesar cipher with shift=7 (exact port from Kotlin)
func DecodeText(text string) string {
	var result strings.Builder
	
	for _, char := range text {
		switch {
		case unicode.IsUpper(char):
			// Uppercase letters: (char - 'A' - SHIFT + 26) % 26 + 'A'
			index := int(char - 'A')
			shifted := (index - SHIFT + 26) % 26
			newChar := rune('A' + shifted)
			result.WriteRune(newChar)
			
		case unicode.IsLower(char):
			// Lowercase letters: (char - 'a' - SHIFT + 26) % 26 + 'a'
			index := int(char - 'a')
			shifted := (index - SHIFT + 26) % 26
			newChar := rune('a' + shifted)
			result.WriteRune(newChar)
			
		case unicode.IsDigit(char):
			// Digits: (digit - SHIFT + 10) % 10
			digit := int(char - '0')
			shifted := (digit - SHIFT + 10) % 10
			result.WriteRune(rune('0' + shifted))
			
		default:
			// Other characters remain unchanged
			result.WriteRune(char)
		}
	}
	
	return result.String()
}

// AddWatermark creates watermark in format <<==[encoded text]==>> (exact port from Kotlin)
func AddWatermark(text string) string {
	encoded := EncodeText(text)
	return fmt.Sprintf("%s%s%s", WATERMARK_PREFIX, encoded, WATERMARK_SUFFIX)
}

// ExtractWatermark extracts watermark from content, supporting both formats (exact port from Kotlin)
func ExtractWatermark(content string) string {
	// Check for new format first: <<==[text]==>>
	if strings.Contains(content, WATERMARK_PREFIX) {
		startIndex := strings.LastIndex(content, WATERMARK_PREFIX)
		endIndex := strings.LastIndex(content, WATERMARK_SUFFIX)
		
		if startIndex != -1 && endIndex != -1 && startIndex < endIndex {
			start := startIndex + len(WATERMARK_PREFIX)
			return content[start:endIndex]
		}
	}
	
	// Check for old format: */[text]
	if strings.Contains(content, OLD_WATERMARK_PREFIX) {
		startIndex := strings.LastIndex(content, OLD_WATERMARK_PREFIX)
		if startIndex != -1 {
			start := startIndex + len(OLD_WATERMARK_PREFIX)
			return strings.TrimSpace(content[start:])
		}
	}
	
	return ""
}

// ProcessFile adds watermark to file if not already present (exact port from Kotlin)
func ProcessFile(filePath string, watermark string) (bool, error) {
	logger := GetGlobalLogger()
	
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		errMsg := fmt.Sprintf("Error processing file %s: %v", filePath, err)
		logger.Error(errMsg)
		return false, err
	}
	
	contentStr := string(content)
	
	// Check if watermark already exists
	if strings.Contains(contentStr, watermark) {
		msg := fmt.Sprintf("%s: Encrypted text already present", filePath)
		logger.Log(msg)
		return false, nil
	}
	
	// Append watermark to file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		errMsg := fmt.Sprintf("Error opening file %s for writing: %v", filePath, err)
		logger.Error(errMsg)
		return false, err
	}
	defer file.Close()
	
	_, err = file.WriteString(watermark)
	if err != nil {
		errMsg := fmt.Sprintf("Error writing watermark to %s: %v", filePath, err)
		logger.Error(errMsg)
		return false, err
	}
	
	// Log success (matching Kotlin output exactly)
	logger.Success(getFileName(filePath))
	return true, nil
}

// Helper function to get filename from path
func getFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}

// TestCaesarCipher tests the Caesar cipher implementation
func TestCaesarCipher() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Test 123", "Alzk 890"},
		{"Project Alpha 001", "Wyvqlua Hswoh 008"},
		{"ABC", "HIJ"},
		{"xyz", "efg"},
		{"987", "654"},
	}
	
	logger := GetGlobalLogger()
	logger.Log("Testing Caesar cipher implementation...")
	
	for _, tc := range testCases {
		encoded := EncodeText(tc.input)
		decoded := DecodeText(encoded)
		
		if encoded == tc.expected {
			logger.Log(fmt.Sprintf("✓ '%s' → '%s' (correct)", tc.input, encoded))
		} else {
			logger.Error(fmt.Sprintf("✗ '%s' → '%s' (expected '%s')", tc.input, encoded, tc.expected))
		}
		
		if decoded == tc.input {
			logger.Log(fmt.Sprintf("✓ Decode test passed: '%s'", decoded))
		} else {
			logger.Error(fmt.Sprintf("✗ Decode failed: got '%s', expected '%s'", decoded, tc.input))
		}
	}
}

// CreateEncodedWatermark creates a complete encoded watermark for given text and order number
func CreateEncodedWatermark(baseText string, orderNumber string) string {
	fullText := fmt.Sprintf("%s %s", baseText, orderNumber)
	return AddWatermark(fullText)
}

// ExtractAndDecodeWatermark extracts and decodes watermark from content
func ExtractAndDecodeWatermark(content string) string {
	extracted := ExtractWatermark(content)
	if extracted == "" {
		return ""
	}
	return DecodeText(extracted)
}