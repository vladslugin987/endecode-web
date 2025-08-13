package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"photo-processing-server/internal/config"
	"photo-processing-server/internal/services"
)

func main() {
	// Command line flags
	testMode := flag.Bool("test", false, "Run in test mode")
	testFile := flag.String("test-file", "", "Test file path for watermark operations")
	testDir := flag.String("test-dir", "", "Test directory for batch processing")
	flag.Parse()

	cfg := config.Load()
	
	// Initialize logging
	logger := services.GetGlobalLogger()
	logger.Info("Starting Photo Processing Server...")
	logger.Info(fmt.Sprintf("Configuration loaded: Port=%s, Environment=%s", cfg.Port, cfg.Environment))
	
	// Test mode for validating ported logic
	if *testMode {
		runTests(*testFile, *testDir)
		return
	}
	
	// TODO: Initialize database, redis, HTTP server
	log.Println("Server starting on port", cfg.Port)
	
	// Placeholder for server startup - will be implemented in future phases
	logger.Info("Server startup placeholder - full server implementation coming in Phase 2")
	logger.Info("Use --test flag to test ported functionality")
	
	// For now, just show usage
	showUsage()
}

func runTests(testFile, testDir string) {
	logger := services.GetGlobalLogger()
	logger.Info("=== Running Tests for Ported Logic ===")
	
	// Test 1: Caesar Cipher
	logger.Info("\n1. Testing Caesar Cipher...")
	services.TestCaesarCipher()
	
	// Test 2: Image Processing (OpenCV)
	logger.Info("\n2. Testing Image Processing...")
	services.TestImageProcessing()
	
	// Test 3: Watermark Operations (if test file provided)
	if testFile != "" {
		logger.Info("\n3. Testing Watermark Operations...")
		if _, err := os.Stat(testFile); err == nil {
			services.TestWatermarkOperations(testFile)
		} else {
			logger.Error(fmt.Sprintf("Test file not found: %s", testFile))
		}
	}
	
	// Test 4: Batch Processing (if test directory provided)
	if testDir != "" {
		logger.Info("\n4. Testing Batch Processing...")
		if info, err := os.Stat(testDir); err == nil && info.IsDir() {
			testBatchProcessing(testDir)
		} else {
			logger.Error(fmt.Sprintf("Test directory not found or not a directory: %s", testDir))
		}
	}
	
	logger.Info("\n=== Test Completed ===")
}

func testBatchProcessing(testDir string) {
	logger := services.GetGlobalLogger()
	
	// Create a test processing job (similar to original Kotlin parameters)
	job := services.CreateProcessingJob(
		"TEST001",           // orderID
		testDir,             // sourceFolder
		3,                   // numCopies
		"Test Project 001",  // baseText
		true,                // addSwap
		true,                // addWatermark
		true,                // createZip
		"CONFIDENTIAL",      // watermarkText
		intPtr(3),           // photoNumber
	)
	
	logger.Info(fmt.Sprintf("Running batch processing test on: %s", testDir))
	logger.Info("Parameters: 3 copies, swap=true, watermark=true, zip=true")
	
	// Progress callback
	progressCallback := func(progress float32) {
		logger.Info(fmt.Sprintf("Progress: %.1f%%", progress*100))
	}
	
	// Run the processing
	err := services.ProcessJob(job, progressCallback)
	if err != nil {
		logger.Error(fmt.Sprintf("Batch processing test failed: %v", err))
	} else {
		logger.Info("✓ Batch processing test completed successfully")
	}
}

func showUsage() {
	logger := services.GetGlobalLogger()
	logger.Info("\n=== Photo Processing Server ===")
	logger.Info("")
	logger.Info("Usage:")
	logger.Info("  ./photo-processor                    - Start server (future)")
	logger.Info("  ./photo-processor --test             - Run all tests")
	logger.Info("  ./photo-processor --test --test-file /path/to/image.jpg")
	logger.Info("  ./photo-processor --test --test-dir /path/to/photos")
	logger.Info("")
	logger.Info("Examples:")
	logger.Info("  # Test Caesar cipher and image processing")
	logger.Info("  ./photo-processor --test")
	logger.Info("")
	logger.Info("  # Test full batch processing on a photo directory")
	logger.Info("  ./photo-processor --test --test-dir /var/photos/session-001")
	logger.Info("")
	logger.Info("Current Status: Core logic ported from Kotlin ✓")
	logger.Info("Next Phase: WooCommerce integration, Admin Panel, etc.")
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}