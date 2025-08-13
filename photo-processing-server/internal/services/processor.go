package services

import (
    "archive/zip"
    "fmt"
    "hash/crc32"
    "os"
    "path/filepath"
    "photo-processing-server/internal/models"
    "regexp"
    "strconv"
    "strings"
    "io"
)

// PerformBatchCopyAndEncode main function for batch copying and encoding (exact port from Kotlin)
func PerformBatchCopyAndEncode(
	sourceFolder string,
	numCopies int,
	baseText string,
	addSwap bool,
	addWatermark bool,
	createZip bool,
	watermarkText string,
	photoNumber *int,
	progress func(float32),
	cleanName string, // Original folder name without UUID suffixes for ZIP naming
) error {
	logger := GetGlobalLogger()
	
	// 1) Create main folder for all copies, e.g. "Test1-Bundle-Copies"
	copiesFolder := filepath.Join(filepath.Dir(sourceFolder), filepath.Base(sourceFolder)+"-Copies")
	err := EnsureDirectoryExists(copiesFolder)
	if err != nil {
		return err
	}
	
	// Extract start number and base text without number
	startNumber := extractStartNumber(baseText)
	baseTextWithoutNumber := strings.TrimSpace(regexp.MustCompile(`\d+$`).ReplaceAllString(baseText, ""))
	
	// Calculate total operations for progress (simplified for clarity)
	totalOperations := float32(numCopies * (1 + 1)) // copying + encoding
	if addWatermark {
		totalOperations += float32(numCopies)
	}
	if addSwap {
		totalOperations += float32(numCopies)  
	}
	if createZip {
		totalOperations += float32(numCopies)
	}
	
	var completedOperations float32 = 0
	
	// Will store (processedFolder, orderNumber) for ZIP creation
	var foldersToProcess []struct {
		folder      string
		orderNumber string
	}
	
	// 2) First pass: create subfolders (001, 002, 003, ...) and copy sourceFolder there
	for i := 0; i < numCopies; i++ {
		orderNumber := fmt.Sprintf("%03d", startNumber+i)
		orderFolder := filepath.Join(copiesFolder, orderNumber)
		err := EnsureDirectoryExists(orderFolder)
		if err != nil {
			return err
		}
		
		// destinationFolder is ".../Test1-Bundle-Copies/001/Test1-Bundle"
		destinationFolder := filepath.Join(orderFolder, filepath.Base(sourceFolder))
		
		// Copy original
		err = CopyDirectory(sourceFolder, destinationFolder)
		if err != nil {
			return err
		}
		completedOperations++
		if progress != nil {
			progress(completedOperations / totalOperations)
		}
		
		// Process files (invisible watermark, rename, etc.)
		err = processFiles(destinationFolder, baseTextWithoutNumber, orderNumber)
		if err != nil {
			return err
		}
		completedOperations++
		if progress != nil {
			progress(completedOperations / totalOperations)
		}
		
		// Visible watermark if needed
		if addWatermark {
			actualPhotoNumber := startNumber + i
			if photoNumber != nil {
				actualPhotoNumber = *photoNumber
			}
			
			actualWatermarkText := orderNumber
			if watermarkText != "" {
				actualWatermarkText = watermarkText
			}
			
			err = addVisibleWatermarkToPhoto(destinationFolder, actualWatermarkText, actualPhotoNumber)
			if err != nil {
				return err
			}
			completedOperations++
			if progress != nil {
				progress(completedOperations / totalOperations)
			}
		}
		
		// Swap if needed
		if addSwap {
			err = performSwap(destinationFolder, orderNumber)
			if err != nil {
				return err
			}
			completedOperations++
			if progress != nil {
				progress(completedOperations / totalOperations)
			}
		}
		
		// Save destination for ZIP stage
		foldersToProcess = append(foldersToProcess, struct {
			folder      string
			orderNumber string
		}{destinationFolder, orderNumber})
		
		logger.Log(fmt.Sprintf("Processed folder: %s", orderNumber))
	}
	
	// 3) Second pass: create ZIP archives and remove processed folders
	if createZip {
		for _, folderInfo := range foldersToProcess {
			// Creates ".../Test1-Bundle-Copies/001/Test1-Bundle.zip"
			// and removes the "Test1-Bundle" subfolder afterwards
			err = createNoCompressionZip(folderInfo.folder, cleanName)
			if err != nil {
				return err
			}
			
			// Now delete the folder (so only the zip remains in "001")
			err = os.RemoveAll(folderInfo.folder)
			if err != nil {
				return err
			}
			
			completedOperations++
			if progress != nil {
				progress(completedOperations / totalOperations)
			}
		}
	}
	
	logger.Log("Batch processing completed successfully")
	return nil
}

// processFiles processes files based on their type (exact port from Kotlin)
func processFiles(folder string, baseText string, orderNumber string) error {
	files, err := GetSupportedFiles(folder)
	if err != nil {
		return err
	}
	
	encodedText := fmt.Sprintf("%s %s", baseText, orderNumber)
	encodedWatermark := EncodeText(encodedText)
	watermark := AddWatermark(encodedText)
	
	for _, file := range files {
		if IsVideoFile(file) {
			// Only add invisible watermark to video files
            err := AddBinaryWatermark(file, encodedWatermark)
			if err != nil {
				return err
			}
			logger := GetGlobalLogger()
			logger.Log(fmt.Sprintf("Added watermark to video: %s", filepath.Base(file)))
		} else {
			// Process other files normally (text files get text watermarks)
			_, err := ProcessFile(file, watermark)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// addVisibleWatermarkToPhoto adds visible watermark to photo with specified number (exact port from Kotlin)
func addVisibleWatermarkToPhoto(folder string, watermarkText string, photoNumber int) error {
	logger := GetGlobalLogger()
	
	files, err := GetSupportedFiles(folder)
	if err != nil {
		return err
	}
	
	found := false
	for _, file := range files {
		if IsImageFile(file) {
			fileNumber := extractFileNumber(filepath.Base(file))
			if fileNumber != nil && *fileNumber == photoNumber {
				err := AddTextToImage(file, watermarkText, BottomRight)
				if err != nil {
					return err
				}
				found = true
				break
			}
		}
	}
	
	if !found {
		logger.Log(fmt.Sprintf("No photo with number %d found in %s", photoNumber, filepath.Base(folder)))
	}
	
	return nil
}

// performSwap performs swap operation for files in folder (exact port from Kotlin)
func performSwap(folder string, orderNumber string) error {
	logger := GetGlobalLogger()
	
	baseNumber, err := strconv.Atoi(orderNumber)
	if err != nil {
		return err
	}
	swapNumber := baseNumber + 10
	
	logger.Processing(fmt.Sprintf("Starting swap operation for number %d with %d ...", baseNumber, swapNumber))
	
	// Take all images in folder
	files, err := GetSupportedFiles(folder)
	if err != nil {
		return err
	}
	
	var allImages []string
	for _, file := range files {
		if IsImageFile(file) {
			allImages = append(allImages, file)
		}
	}
	
	// Find file with the baseNumber
	var fileA, fileB string
	for _, file := range allImages {
		fileNum := extractFileNumber(filepath.Base(file))
		if fileNum != nil {
			if *fileNum == baseNumber {
				fileA = file
			} else if *fileNum == swapNumber {
				fileB = file
			}
		}
	}
	
	// If there are no matching files - stop swapping
	if fileA == "" || fileB == "" {
		logger.Log(fmt.Sprintf("No matching pair found for swapping in folder %s (need %d and %d)", 
			filepath.Base(folder), baseNumber, swapNumber))
		return nil
	}
	
	// If found - rename
	err = swapFiles(fileA, fileB)
	if err != nil {
		return err
	}
	
	logger.Log(fmt.Sprintf("Finished swap operation for folder %s", filepath.Base(folder)))
	return nil
}

// swapFiles swaps two files: fileA -> temp, fileB -> fileA, temp -> fileB (exact port from Kotlin)
func swapFiles(fileA, fileB string) error {
	logger := GetGlobalLogger()
	
	logger.Log("Swapping files:")
	logger.Log(fmt.Sprintf("  - %s", filepath.Base(fileA)))
	logger.Log(fmt.Sprintf("  - %s", filepath.Base(fileB)))
	
	// Create temporary file
	tempFile := filepath.Join(filepath.Dir(fileA), fmt.Sprintf("temp_%d_%s", 
		os.Getpid(), filepath.Base(fileA)))
	
	// Step 1: fileA -> temp
	err := os.Rename(fileA, tempFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to swap files %s <--> %s: %v", 
			filepath.Base(fileA), filepath.Base(fileB), err))
		return err
	}
	
	// Step 2: fileB -> fileA
	err = os.Rename(fileB, fileA)
	if err != nil {
		// Try to recover
		os.Rename(tempFile, fileA)
		logger.Error(fmt.Sprintf("Failed to swap files %s <--> %s: %v", 
			filepath.Base(fileA), filepath.Base(fileB), err))
		return err
	}
	
	// Step 3: temp -> fileB
	err = os.Rename(tempFile, fileB)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to swap files %s <--> %s: %v", 
			filepath.Base(fileA), filepath.Base(fileB), err))
		return err
	}
	
	logger.Log(fmt.Sprintf("Successfully swapped %s <--> %s", 
		filepath.Base(fileA), filepath.Base(fileB)))
	return nil
}

// createNoCompressionZip creates ZIP archive without compression (exact port from Kotlin)
func createNoCompressionZip(folderToZip string, cleanName ...string) error {
	logger := GetGlobalLogger()
	
	// Use clean name if provided, otherwise use folder name
	var zipFileName string
	if len(cleanName) > 0 && cleanName[0] != "" {
		zipFileName = cleanName[0] + ".zip"
	} else {
		zipFileName = filepath.Base(folderToZip) + ".zip"
	}
	
	zipFile := filepath.Join(filepath.Dir(folderToZip), zipFileName)
	
	// Create zip file
	zipFileHandle, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer zipFileHandle.Close()
	
	zipWriter := zip.NewWriter(zipFileHandle)
	defer zipWriter.Close()
	
	// Walk through folder to zip
	err = filepath.Walk(folderToZip, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip system files (exact port from Kotlin filter)
		if strings.HasPrefix(info.Name(), "__MACOSX") ||
		   strings.HasPrefix(info.Name(), ".") ||
		   strings.HasSuffix(info.Name(), ".DS_Store") {
			return nil
		}
		
		// Calculate relative path
		relPath, err := filepath.Rel(folderToZip, path)
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			// Add directory entry
			return addDirectoryToZip(relPath, zipWriter)
		} else {
			// Add file entry
			return addFileToZip(path, relPath, zipWriter)
		}
	})
	
	if err != nil {
		return err
	}
	
	logger.Log(fmt.Sprintf("Created ZIP archive: %s", zipFile))
	return nil
}

// addFileToZip adds file to ZIP archive without compression (exact port from Kotlin)
func addFileToZip(filePath string, entryPath string, zipWriter *zip.Writer) error {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	// Calculate CRC32 for stored method
	crc := crc32.ChecksumIEEE(fileBytes)
	
	// Create header for STORED method (no compression)
	header := &zip.FileHeader{
		Name:               entryPath,
		Method:             zip.Store, // No compression
		UncompressedSize64: uint64(len(fileBytes)),
		CRC32:              crc,
	}
	
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	
	_, err = writer.Write(fileBytes)
	return err
}

// addDirectoryToZip adds directory entry to ZIP archive (exact port from Kotlin)
func addDirectoryToZip(entryPath string, zipWriter *zip.Writer) error {
	// Ensure directory path ends with /
	if !strings.HasSuffix(entryPath, "/") {
		entryPath += "/"
	}
	
	header := &zip.FileHeader{
		Name:   entryPath,
		Method: zip.Store,
	}
	
	_, err := zipWriter.CreateHeader(header)
	return err
}

// StreamNoCompressionZip writes a no-compression ZIP of folderToZip to the provided writer
func StreamNoCompressionZip(w io.Writer, folderToZip string) error {
    zipWriter := zip.NewWriter(w)
    defer zipWriter.Close()

    // Walk through folder to zip
    err := filepath.Walk(folderToZip, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip system files
        if strings.HasPrefix(info.Name(), "__MACOSX") ||
           strings.HasPrefix(info.Name(), ".") ||
           strings.HasSuffix(info.Name(), ".DS_Store") {
            return nil
        }

        // Calculate relative path
        relPath, err := filepath.Rel(folderToZip, path)
        if err != nil {
            return err
        }

        if info.IsDir() {
            return addDirectoryToZip(relPath, zipWriter)
        }

        // File entry with STORED method
        fileBytes, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        crc := crc32.ChecksumIEEE(fileBytes)
        header := &zip.FileHeader{
            Name:               relPath,
            Method:             zip.Store,
            UncompressedSize64: uint64(len(fileBytes)),
            CRC32:              crc,
        }
        writer, err := zipWriter.CreateHeader(header)
        if err != nil {
            return err
        }
        _, err = writer.Write(fileBytes)
        return err
    })

    return err
}

// extractStartNumber extracts number from the end of text (exact port from Kotlin)
func extractStartNumber(text string) int {
	re := regexp.MustCompile(`\d+$`)
	match := re.FindString(text)
	if match == "" {
		return 1
	}
	
	num, err := strconv.Atoi(match)
	if err != nil {
		return 1
	}
	return num
}

// CreateProcessingJob creates a processing job from parameters
func CreateProcessingJob(orderID, sourceFolder string, numCopies int, baseText string, 
	addSwap, addWatermark, createZip bool, watermarkText string, photoNumber *int) *models.ProcessingJob {
	
	return &models.ProcessingJob{
		OrderID:       orderID,
		SourcePath:    sourceFolder,
		NumCopies:     numCopies,
		BaseText:      baseText,
		AddSwap:       addSwap,
		AddWatermark:  addWatermark,
		CreateZip:     createZip,
		WatermarkText: watermarkText,
		PhotoNumber:   photoNumber,
		Status:        string(models.StatusPending),
	}
}

// ProcessJob processes a processing job
func ProcessJob(job *models.ProcessingJob, progress func(float32)) error {
	return PerformBatchCopyAndEncode(
		job.SourcePath,
		job.NumCopies,
		job.BaseText,
		job.AddSwap,
		job.AddWatermark,
		job.CreateZip,
		job.WatermarkText,
		job.PhotoNumber,
		progress,
		job.SourcePath, // Pass the original source folder name for clean naming
	)
}