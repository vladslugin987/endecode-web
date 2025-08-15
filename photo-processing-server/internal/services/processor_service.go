package services

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

// BatchSettings maps 1:1 to the frontend BatchCopySettings
// and to the Kotlin dialog parameters.
type BatchSettings struct {
    NumberOfCopies               int                    `json:"numberOfCopies"`
    BaseText                     string                 `json:"baseText"`
    AddSwapEncoding              bool                   `json:"addSwapEncoding"`
    SwapPairs                    []map[string]string    `json:"swapPairs,omitempty"`
    AddVisibleWatermark          bool                   `json:"addVisibleWatermark"`
    WatermarkPositions           []int                  `json:"watermarkPositions,omitempty"`
    CreateZip                    bool                   `json:"createZip"`
    WatermarkText                string                 `json:"watermarkText,omitempty"`
    PhotoNumber                  *int                   `json:"photoNumber,omitempty"`
    UseOrderNumberAsPhotoNumber  bool                   `json:"useOrderNumberAsPhotoNumber,omitempty"`
}

// Processor provides high-level operations used by HTTP handlers.
type Processor struct {
    logger *Logger
}

// NewProcessor creates a new Processor instance.
func NewProcessor(logger *Logger) *Processor {
    if logger == nil {
        logger = GetGlobalLogger()
    }
    return &Processor{logger: logger}
}

// EncryptFiles applies watermarks/encoding across supported files in the directory.
// Progress callback receives values in [0.0, 1.0].
func (p *Processor) EncryptFiles(selectedPath string, nameToInject string, progress func(float64)) error {
    if selectedPath == "" {
        return fmt.Errorf("selectedPath is empty")
    }
    if info, err := os.Stat(selectedPath); err != nil || !info.IsDir() {
        return fmt.Errorf("selectedPath is not a directory or does not exist: %s", selectedPath)
    }

    p.logger.Processing("[ENCRYPT] Scanning files...")
    files, err := GetSupportedFiles(selectedPath)
    if err != nil {
        return err
    }
    if len(files) == 0 {
        return fmt.Errorf("no supported files found in: %s", selectedPath)
    }
    p.logger.Log(fmt.Sprintf("[ENCRYPT] Found %d supported files", len(files)))
    total := float32(len(files))
    var processed float32 = 0

    // Pre-compute watermark strings
    textWatermark := AddWatermark(nameToInject) // includes markers and encoded text
    encodedOnly := EncodeText(nameToInject)     // for binary watermark API

    for _, file := range files {
        switch {
        case IsVideoFile(file):
            // Add invisible binary watermark to media
            if err := AddBinaryWatermark(file, encodedOnly); err != nil {
                return err
            }
        case IsTextFile(file):
            // Append text watermark to text files
            if _, err := ProcessFile(file, textWatermark); err != nil {
                return err
            }
        default:
            // Images: keep consistent with batch logic — add invisible binary watermark
            if IsImageFile(file) {
                if err := AddBinaryWatermark(file, encodedOnly); err != nil {
                    return err
                }
            }
        }

        processed++
        if progress != nil && total > 0 {
            progress(float64(processed / total))
        }
    }

    p.logger.Log("[ENCRYPT] Encryption completed successfully")
    return nil
}

// DecryptFiles scans files and logs extracted watermarks.
func (p *Processor) DecryptFiles(selectedPath string, progress func(float64)) error {
    if selectedPath == "" {
        return fmt.Errorf("selectedPath is empty")
    }
    if info, err := os.Stat(selectedPath); err != nil || !info.IsDir() {
        return fmt.Errorf("selectedPath is not a directory or does not exist: %s", selectedPath)
    }

    p.logger.Processing("[DECRYPT] Scanning files...")
    files, err := GetSupportedFiles(selectedPath)
    if err != nil {
        return err
    }
    if len(files) == 0 {
        return fmt.Errorf("no supported files found in: %s", selectedPath)
    }
    p.logger.Log(fmt.Sprintf("[DECRYPT] Found %d supported files", len(files)))
    total := float32(len(files))
    var processed float32 = 0

    foundCount := 0
    for _, file := range files {
        var decoded string

        // Try binary watermark extraction first (images/videos)
        if IsImageFile(file) || IsVideoFile(file) {
            if encoded, err := ExtractWatermarkText(file); err == nil && encoded != "" {
                decoded = DecodeText(encoded)
            }
        }

        // For text files, check appended textual watermark
        if decoded == "" && IsTextFile(file) {
            if content, err := ioutil.ReadFile(file); err == nil {
                decoded = ExtractAndDecodeWatermark(string(content))
            }
        }

        if decoded != "" {
            p.logger.Log(fmt.Sprintf("%s → %s", getFileName(file), decoded))
            foundCount++
        }

        processed++
        if progress != nil && total > 0 {
            progress(float64(processed / total))
        }
    }

    if foundCount == 0 {
        p.logger.Log("[DECRYPT] Completed: no watermarks found")
    } else {
        p.logger.Log(fmt.Sprintf("[DECRYPT] Completed: scanned %d files, found %d watermarks", len(files), foundCount))
    }
    return nil
}

// PerformBatchCopy runs the full batch copy and encoding flow.
func (p *Processor) PerformBatchCopy(selectedPath string, settings BatchSettings, progress func(float64)) error {
    if selectedPath == "" {
        return fmt.Errorf("selectedPath is empty")
    }
    if info, err := os.Stat(selectedPath); err != nil || !info.IsDir() {
        return fmt.Errorf("selectedPath is not a directory or does not exist: %s", selectedPath)
    }

    // Extract clean folder name (remove UUID suffix if present)
    folderName := filepath.Base(selectedPath)
    cleanName := folderName
    
    // If folder name contains UUID pattern (ends with _xxxxxxxx), remove it
    if parts := strings.Split(folderName, "_"); len(parts) > 1 {
        lastPart := parts[len(parts)-1]
        // Check if last part looks like an 8-character UUID fragment
        if len(lastPart) == 8 && isHexString(lastPart) {
            cleanName = strings.Join(parts[:len(parts)-1], "_")
        }
    }

    // If using order number as photo number, pass nil to use dynamic number inside the loop.
    var photoPtr *int
    if !settings.UseOrderNumberAsPhotoNumber {
        photoPtr = settings.PhotoNumber
    }

    return PerformBatchCopyAndEncode(
        selectedPath,
        settings.NumberOfCopies,
        settings.BaseText,
        settings.AddSwapEncoding,
        settings.AddVisibleWatermark,
        settings.CreateZip,
        settings.WatermarkText,
        photoPtr,
        func(pf float32) {
            if progress != nil {
                progress(float64(pf))
            }
        },
        cleanName, // Pass clean name for ZIP files
    )
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
    for _, r := range s {
        if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
            return false
        }
    }
    return true
}

// AddTextToPhoto adds a visible watermark text to a specific photo number in the folder.
func (p *Processor) AddTextToPhoto(selectedPath string, text string, photoNumber int) error {
    if selectedPath == "" {
        return fmt.Errorf("selectedPath is empty")
    }
    return addVisibleWatermarkToPhoto(selectedPath, text, photoNumber)
}

// RemoveWatermarks removes invisible watermarks from supported media files.
func (p *Processor) RemoveWatermarks(selectedPath string, progress func(float64)) error {
    if selectedPath == "" {
        return fmt.Errorf("selectedPath is empty")
    }
    return RemoveWatermarks(selectedPath, func(pf float32) {
        if progress != nil {
            progress(float64(pf))
        }
    })
}


