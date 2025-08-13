package web

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "photo-processing-server/internal/services"
    "photo-processing-server/internal/config"
    "sync"
    "archive/zip"
    "strings"
    "net/url"
    "strconv"
    "regexp"
    "sort"
)

// Request/Response types matching TypeScript interfaces
type ProcessingRequest struct {
	SelectedPath string `json:"selectedPath"`
	NameToInject string `json:"nameToInject,omitempty"`
}

type BatchCopyRequest struct {
	SelectedPath string                 `json:"selectedPath"`
	Settings     services.BatchSettings `json:"settings"`
}

type AddTextRequest struct {
	SelectedPath string `json:"selectedPath"`
	Settings     struct {
		Text        string `json:"text"`
		PhotoNumber int    `json:"photoNumber"`
	} `json:"settings"`
}

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	JobID   string `json:"jobId,omitempty"`
	Error   string `json:"error,omitempty"`
	Token   string `json:"token,omitempty"`
}

type ProcessingJob struct {
	ID       string
	Status   string
	Progress float64
	Message  string
	Error    string
	Result   interface{}
	StartTime time.Time
}

// In-memory job store (in production, use Redis or database)
var jobs = make(map[string]*ProcessingJob)
// Temporary in-memory download tokens store: token -> filesystem path
var downloadTokens = make(map[string]string)

// Active operation locks: op|path -> jobID to prevent duplicate concurrent jobs
var activeOps = make(map[string]string)
var activeMutex sync.Mutex

func opKey(op, path string) string {
    return op + "|" + path
}

// WebHandler contains all web-related handlers
type WebHandler struct {
	processor *services.Processor
	logger    *services.Logger
}

// NewWebHandler creates a new web handler
func NewWebHandler(processor *services.Processor, logger *services.Logger) *WebHandler {
	return &WebHandler{
		processor: processor,
		logger:    logger,
	}
}

// SetupRoutes configures all web routes
func (h *WebHandler) SetupRoutes(router *gin.Engine) {
    // Serve static files (React build)
    router.Static("/assets", "./web/frontend/dist/assets")
    
    // Serve index.html with no-store to avoid caching stale bundles
    router.GET("/", func(c *gin.Context) {
        c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
        c.File("./web/frontend/dist/index.html")
    })
    router.GET("/favicon.ico", func(c *gin.Context) {
        c.File("./web/frontend/dist/favicon.ico")
    })
	
    // Fallback for React Router (SPA) with no-store to avoid cached HTML
    router.NoRoute(func(c *gin.Context) {
        // Do not override API/WS/asset routes
        path := c.Request.URL.Path
        if len(path) >= 4 && path[:4] == "/api" {
            c.Status(http.StatusNotFound)
            return
        }
        if len(path) >= 3 && path[:3] == "/ws" {
            c.Status(http.StatusNotFound)
            return
        }
        if len(path) >= 7 && path[:7] == "/assets" {
            c.Status(http.StatusNotFound)
            return
        }
        c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
        c.File("./web/frontend/dist/index.html")
    })

	// API routes
	api := router.Group("/api")
	{
		api.POST("/encrypt", h.handleEncrypt)
		api.POST("/decrypt", h.handleDecrypt)
		api.POST("/batch-copy", h.handleBatchCopy)
		api.POST("/add-text", h.handleAddText)
		api.POST("/remove-watermarks", h.handleRemoveWatermarks)
		api.POST("/upload", h.handleUpload)
		api.GET("/processing/:id", h.handleProcessingStatus)
		api.GET("/download/:token", h.handleDownload)
		// Admin
		api.GET("/admin/jobs", h.handleListJobs)
		api.GET("/admin/jobs/:id", h.handleJobDetails)
		api.POST("/admin/jobs/:id/approve", h.handleApproveJob)
		api.GET("/admin/jobs/:id/images", h.handleJobImages)
		api.GET("/admin/jobs/:id/preview", h.handleJobPreview)
		api.GET("/admin/jobs/:id/stats", h.handleJobStats)
		api.GET("/admin/jobs/:id/logs", h.handleJobLogs)
	}

	// WebSocket route will be added in websocket.go
}

// Helper function to create and start a processing job, passing jobID into operation
func (h *WebHandler) startJob(operation func(jobID string) error) string {
    jobID := uuid.New().String()

    job := &ProcessingJob{
        ID:        jobID,
        Status:    "processing",
        Progress:  0.0,
        StartTime: time.Now(),
    }

    SaveJob(job)

    // Run operation in background
    go func(localJobID string) {
        defer func() {
            if r := recover(); r != nil {
                if _, ok := GetJob(localJobID); ok {
                    SetJobStatus(localJobID, "error", fmt.Sprintf("Panic: %v", r))
                }
            }
        }()

        err := operation(localJobID)
        if err != nil {
            SetJobStatus(localJobID, "error", err.Error())
            BroadcastError(localJobID, err.Error())
        } else {
            SetJobStatus(localJobID, "completed", "")
            BroadcastComplete(localJobID, nil)
        }
    }(jobID)

    return jobID
}

// Encrypt handler
func (h *WebHandler) handleEncrypt(c *gin.Context) {
    // optional API token auth
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    var req ProcessingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ApiResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    // Deduplicate concurrent operations per path
    key := opKey("encrypt", req.SelectedPath)
    activeMutex.Lock()
    if existing, ok := activeOps[key]; ok {
        activeMutex.Unlock()
        c.JSON(http.StatusOK, ApiResponse{Success: true, JobID: existing, Message: "Encryption already in progress"})
        return
    }
    activeOps[key] = "pending"
    activeMutex.Unlock()

    // Log only once when we actually start
    h.logger.Log("=== Starting Encryption Process ===")
    h.logger.Log(fmt.Sprintf("Selected Path: %s", req.SelectedPath))
    h.logger.Log(fmt.Sprintf("Name to Inject: %s", req.NameToInject))

    jobID := h.startJob(func(id string) error {
        activeMutex.Lock()
        activeOps[key] = id
        activeMutex.Unlock()
        err := h.processor.EncryptFiles(req.SelectedPath, req.NameToInject, func(progress float64) {
            UpdateJobProgress(id, progress)
            BroadcastProgress(id, progress)
        })
        if err != nil {
            h.logger.Error(fmt.Sprintf("Encrypt error: %v", err))
        }
        activeMutex.Lock()
        delete(activeOps, key)
        activeMutex.Unlock()
        return err
    })

    c.JSON(http.StatusOK, ApiResponse{
        Success: true,
        JobID:   jobID,
        Message: "Encryption started",
    })

    // Log job start with job ID
    h.logger.Processing(fmt.Sprintf("JOB %s: Encryption started for %s", jobID, req.SelectedPath))
}

// Decrypt handler
func (h *WebHandler) handleDecrypt(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    var req ProcessingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ApiResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    key := opKey("decrypt", req.SelectedPath)
    activeMutex.Lock()
    if existing, ok := activeOps[key]; ok {
        activeMutex.Unlock()
        c.JSON(http.StatusOK, ApiResponse{Success: true, JobID: existing, Message: "Decryption already in progress"})
        return
    }
    activeOps[key] = "pending"
    activeMutex.Unlock()

    // Log only once when we actually start
    h.logger.Log("=== Starting Decryption Process ===")
    h.logger.Log(fmt.Sprintf("Selected Path: %s", req.SelectedPath))

    jobID := h.startJob(func(id string) error {
        activeMutex.Lock()
        activeOps[key] = id
        activeMutex.Unlock()
        err := h.processor.DecryptFiles(req.SelectedPath, func(progress float64) {
            UpdateJobProgress(id, progress)
            BroadcastProgress(id, progress)
        })
        if err != nil {
            h.logger.Error(fmt.Sprintf("Decrypt error: %v", err))
        }
        activeMutex.Lock()
        delete(activeOps, key)
        activeMutex.Unlock()
        return err
    })

    c.JSON(http.StatusOK, ApiResponse{
        Success: true,
        JobID:   jobID,
        Message: "Decryption started",
    })

    h.logger.Processing(fmt.Sprintf("JOB %s: Decryption started for %s", jobID, req.SelectedPath))
}

// Batch copy handler
func (h *WebHandler) handleBatchCopy(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    var req BatchCopyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ApiResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    key := opKey("batch", req.SelectedPath)
    activeMutex.Lock()
    if existing, ok := activeOps[key]; ok {
        activeMutex.Unlock()
        c.JSON(http.StatusOK, ApiResponse{Success: true, JobID: existing, Message: "Batch already in progress"})
        return
    }
    activeOps[key] = "pending"
    activeMutex.Unlock()

    // Log only once when we actually start
    h.logger.Log("=== Starting Batch Copy Process ===")
    h.logger.Log(fmt.Sprintf("Selected Path: %s", req.SelectedPath))
    h.logger.Log(fmt.Sprintf("Number of copies: %d", req.Settings.NumberOfCopies))
    h.logger.Log(fmt.Sprintf("Base text: %s", req.Settings.BaseText))

    jobID := h.startJob(func(id string) error {
        activeMutex.Lock()
        activeOps[key] = id
        activeMutex.Unlock()
        err := h.processor.PerformBatchCopy(req.SelectedPath, req.Settings, func(progress float64) {
            UpdateJobProgress(id, progress)
            BroadcastProgress(id, progress)
        })
        if err != nil {
            h.logger.Error(fmt.Sprintf("Batch copy error: %v", err))
        } else {
            // After successful batch, detect the resulting zip path and create a one-time token
            // Result zip expected as <base>/...-Copies/<order>/...zip or last processed folder zip
            // For simplicity, set result to the Copies folder path
            resultPath := filepath.Join(filepath.Dir(req.SelectedPath), filepath.Base(req.SelectedPath)+"-Copies")
            dlToken := uuid.New().String()
            SaveDownloadToken(dlToken, resultPath)
            // Try to determine a sample image with visible watermark for zoom preview
            sample := map[string]string{}
            if req.Settings.AddVisibleWatermark {
                entries, _ := os.ReadDir(resultPath)
                orderDirs := make([]string, 0)
                for _, e := range entries { if e.IsDir() { orderDirs = append(orderDirs, e.Name()) } }
                sort.Strings(orderDirs)
                // iterate through order folders until sample found
                for _, order := range orderDirs {
                    // Determine target photo number for this order
                    targetNum := 0
                    if req.Settings.UseOrderNumberAsPhotoNumber || req.Settings.PhotoNumber == nil {
                        if n, errAtoi := strconv.Atoi(order); errAtoi == nil { targetNum = n }
                    } else {
                        targetNum = *req.Settings.PhotoNumber
                    }
                    orderPath := filepath.Join(resultPath, order)
                    // Check zip first
                    var zipPath string
                    files, _ := os.ReadDir(orderPath)
                    for _, f := range files {
                        if !f.IsDir() && strings.ToLower(filepath.Ext(f.Name())) == ".zip" {
                            zipPath = filepath.Join(orderPath, f.Name())
                            break
                        }
                    }
                    if zipPath != "" {
                        zr, errOpen := zip.OpenReader(zipPath)
                        if errOpen == nil {
                            for _, f := range zr.File {
                                base := filepath.Base(f.Name)
                                re := regexp.MustCompile(`\d+`)
                                if m := re.FindString(base); m != "" {
                                    if n, _ := strconv.Atoi(m); n == targetNum {
                                        relZip, _ := filepath.Rel(resultPath, zipPath)
                                        sample["zip"] = relZip
                                        sample["entry"] = f.Name
                                        zr.Close()
                                        break
                                    }
                                }
                            }
                            zr.Close()
                        }
                    }
                    // If not found in zip, scan files
                    if len(sample) == 0 {
                        _ = filepath.Walk(orderPath, func(p string, info os.FileInfo, err error) error {
                            if err != nil || info.IsDir() { return nil }
                            ext := strings.ToLower(filepath.Ext(p))
                            if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
                                base := filepath.Base(p)
                                re := regexp.MustCompile(`\d+`)
                                if m := re.FindString(base); m != "" {
                                    if n, _ := strconv.Atoi(m); n == targetNum {
                                        rel, _ := filepath.Rel(resultPath, p)
                                        sample["path"] = rel
                                        return io.EOF
                                    }
                                }
                            }
                            return nil
                        })
                    }
                    if len(sample) > 0 { break }
                }
            }
            resultObj := map[string]interface{}{"downloadToken": dlToken, "path": resultPath}
            if len(sample) > 0 { resultObj["watermarkSample"] = sample }
            SetJobResult(id, resultObj)
            BroadcastLog(fmt.Sprintf("Result available. Token: %s", dlToken))
        }
        activeMutex.Lock()
        delete(activeOps, key)
        activeMutex.Unlock()
        return err
    })

    c.JSON(http.StatusOK, ApiResponse{
        Success: true,
        JobID:   jobID,
        Message: "Batch copy started",
    })

    h.logger.Processing(fmt.Sprintf("JOB %s: Batch copy started for %s", jobID, req.SelectedPath))
}

// Add text handler
func (h *WebHandler) handleAddText(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    var req AddTextRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ApiResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    key := opKey("addtext", req.SelectedPath)
    activeMutex.Lock()
    if existing, ok := activeOps[key]; ok {
        activeMutex.Unlock()
        c.JSON(http.StatusOK, ApiResponse{Success: true, JobID: existing, Message: "Add text already in progress"})
        return
    }
    activeOps[key] = "pending"
    activeMutex.Unlock()

    // Log only once when we actually start
    h.logger.Log("=== Adding Text to Photo ===")
    h.logger.Log(fmt.Sprintf("Selected Path: %s", req.SelectedPath))
    h.logger.Log(fmt.Sprintf("Text: %s", req.Settings.Text))
    h.logger.Log(fmt.Sprintf("Photo Number: %d", req.Settings.PhotoNumber))

    jobID := h.startJob(func(id string) error {
        activeMutex.Lock()
        activeOps[key] = id
        activeMutex.Unlock()
        err := h.processor.AddTextToPhoto(req.SelectedPath, req.Settings.Text, req.Settings.PhotoNumber)
        activeMutex.Lock()
        delete(activeOps, key)
        activeMutex.Unlock()
        return err
    })

    c.JSON(http.StatusOK, ApiResponse{
        Success: true,
        JobID:   jobID,
        Message: "Adding text to photo",
    })

    h.logger.Processing(fmt.Sprintf("JOB %s: Add text started for %s", jobID, req.SelectedPath))
}

// Remove watermarks handler
func (h *WebHandler) handleRemoveWatermarks(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    var req ProcessingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ApiResponse{
            Success: false,
            Error:   fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    key := opKey("remove", req.SelectedPath)
    activeMutex.Lock()
    if existing, ok := activeOps[key]; ok {
        activeMutex.Unlock()
        c.JSON(http.StatusOK, ApiResponse{Success: true, JobID: existing, Message: "Removal already in progress"})
        return
    }
    activeOps[key] = "pending"
    activeMutex.Unlock()

    // Log only once when we actually start
    h.logger.Log("=== Removing Watermarks ===")
    h.logger.Log(fmt.Sprintf("Selected Path: %s", req.SelectedPath))

    jobID := h.startJob(func(id string) error {
        activeMutex.Lock()
        activeOps[key] = id
        activeMutex.Unlock()
        err := h.processor.RemoveWatermarks(req.SelectedPath, func(progress float64) {
            UpdateJobProgress(id, progress)
            BroadcastProgress(id, progress)
        })
        if err != nil {
            h.logger.Error(fmt.Sprintf("Remove watermarks error: %v", err))
        }
        activeMutex.Lock()
        delete(activeOps, key)
        activeMutex.Unlock()
        return err
    })

    c.JSON(http.StatusOK, ApiResponse{
        Success: true,
        JobID:   jobID,
        Message: "Removing watermarks",
    })

    h.logger.Processing(fmt.Sprintf("JOB %s: Remove watermarks started for %s", jobID, req.SelectedPath))
}

// Upload handler
func (h *WebHandler) handleUpload(c *gin.Context) {
    // optional API token auth
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(http.StatusBadRequest, ApiResponse{
            Success: false,
            Error:   fmt.Sprintf("Failed to parse multipart form: %v", err),
        })
        return
    }

    files := form.File
    if len(files) == 0 {
        c.JSON(http.StatusBadRequest, ApiResponse{
            Success: false,
            Error:   "No files uploaded",
        })
        return
    }

    // First, check if folder name was explicitly provided via form field
    var originalFolderName string
    if folderNames := form.Value["folderName"]; len(folderNames) > 0 && folderNames[0] != "" {
        originalFolderName = folderNames[0]
        h.logger.Log(fmt.Sprintf("Debug: Using provided folder name: %s", originalFolderName))
    } else {
        // Extract original folder name from first file path (fallback)
        for _, fileHeaders := range files {
            if len(fileHeaders) > 0 {
                // Get the first directory in the path
                firstPath := fileHeaders[0].Filename
                h.logger.Log(fmt.Sprintf("Debug: First file path: %s", firstPath))
                
                if strings.Contains(firstPath, "/") {
                    parts := strings.Split(firstPath, "/")
                    if len(parts) > 0 && parts[0] != "" {
                        originalFolderName = parts[0]
                        break
                    }
                } else {
                    // If no slash, check if filename contains folder structure info
                    // Some browsers might format differently
                    baseName := filepath.Base(firstPath)
                    if baseName != firstPath {
                        originalFolderName = filepath.Dir(firstPath)
                    }
                }
            }
        }
        h.logger.Log(fmt.Sprintf("Debug: Extracted folder name from path: '%s'", originalFolderName))
    }
    
    // Fallback to timestamp if no folder structure found
    if originalFolderName == "" || originalFolderName == "." {
        originalFolderName = fmt.Sprintf("upload_%d", time.Now().Unix())
        h.logger.Log(fmt.Sprintf("Debug: Using fallback name: %s", originalFolderName))
    }

    // Create upload directory using only original folder name (add UUID only if collision)
    uploadDir := filepath.Join(cfg.UploadsPath, originalFolderName)
    
    // Check if directory already exists, if so add UUID suffix
    if _, err := os.Stat(uploadDir); err == nil {
        uploadDir = filepath.Join(cfg.UploadsPath, fmt.Sprintf("%s_%s", originalFolderName, uuid.New().String()[:8]))
    }
    
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        c.JSON(http.StatusInternalServerError, ApiResponse{
            Success: false,
            Error:   fmt.Sprintf("Failed to create upload directory: %v", err),
        })
        return
    }

    var savedFiles []string
    totalFiles := 0
    
    for _, fileHeaders := range files {
        for _, fileHeader := range fileHeaders {
            totalFiles++
            
            // Open uploaded file
            file, err := fileHeader.Open()
            if err != nil {
                h.logger.Log(fmt.Sprintf("Failed to open file %s: %v", fileHeader.Filename, err))
                continue
            }
            defer file.Close()

            // Preserve directory structure within the upload
            dst := filepath.Join(uploadDir, fileHeader.Filename)
            
            // Create directories if needed
            if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
                h.logger.Log(fmt.Sprintf("Failed to create directory for %s: %v", dst, err))
                continue
            }
            
            out, err := os.Create(dst)
            if err != nil {
                h.logger.Log(fmt.Sprintf("Failed to create file %s: %v", dst, err))
                continue
            }
            defer out.Close()

            // Copy file content
            if _, err = io.Copy(out, file); err != nil {
                h.logger.Log(fmt.Sprintf("Failed to save file %s: %v", dst, err))
                continue
            }

            savedFiles = append(savedFiles, dst)
        }
    }

    // Issue download token for the uploaded folder
    token := uuid.New().String()
    SaveDownloadToken(token, uploadDir)

    h.logger.Log(fmt.Sprintf("Uploaded %d/%d files to %s (original: %s)", len(savedFiles), totalFiles, uploadDir, originalFolderName))

    // Backward + forward compatible response
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": fmt.Sprintf("Uploaded %d files", len(savedFiles)),
        // legacy fields expected by some clients
        "path": uploadDir,
        "downloadToken": token,
        // new canonical fields used by ApiResponse
        "jobId": uploadDir,
        "token": token,
        // Add original folder name for display and processing
        "originalName": originalFolderName,
        "cleanName": originalFolderName, // Store clean name without any UUID suffixes
    })
}

// Processing status handler
func (h *WebHandler) handleProcessingStatus(c *gin.Context) {
	jobID := c.Param("id")
	
	job, exists := GetJob(jobID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	response := gin.H{
		"status":   job.Status,
		"progress": job.Progress,
	}

	if job.Message != "" {
		response["message"] = job.Message
	}
	if job.Error != "" {
		response["error"] = job.Error
	}
	if job.Result != nil {
		response["result"] = job.Result
	}

	c.JSON(http.StatusOK, response)
}

// Download handler (placeholder - implement based on your needs)
func (h *WebHandler) handleDownload(c *gin.Context) {
    token := c.Param("token")
    path, ok := GetDownloadPath(token)
    if !ok {
        c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired token"})
        return
    }

    // If path is a directory, create a zip on the fly and stream it
    info, err := os.Stat(path)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Path not found"})
        return
    }

    if info.IsDir() {
        zipName := filepath.Base(path) + ".zip"
        c.Header("Content-Type", "application/zip")
        c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipName))
        c.Status(http.StatusOK)

        if err := services.StreamNoCompressionZip(c.Writer, path); err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create zip"})
            return
        }
        // one-time token: delete after successful stream
        DeleteDownloadToken(token)
        return
    }

    // Otherwise stream a single file
    c.File(path)
    DeleteDownloadToken(token)
}

// Admin: list jobs
func (h *WebHandler) handleListJobs(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    c.JSON(http.StatusOK, gin.H{"jobs": ListJobs()})
}

// Admin: job details
func (h *WebHandler) handleJobDetails(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    id := c.Param("id")
    if job, ok := GetJob(id); ok {
        c.JSON(http.StatusOK, job)
        return
    }
    c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
}

// Admin: approve job and issue download token
func (h *WebHandler) handleApproveJob(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    id := c.Param("id")
    job, ok := GetJob(id)
    if !ok {
        c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
        return
    }
    // Expect result to contain path
    var path string
    if m, ok := job.Result.(map[string]interface{}); ok {
        if p, ok := m["path"].(string); ok && p != "" {
            path = p
        }
    }
    if path == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Job has no result path to approve"})
        return
    }
    token := uuid.New().String()
    SaveDownloadToken(token, path)
    // update job result with approved token
    if m, ok := job.Result.(map[string]interface{}); ok {
        m["approvedToken"] = token
        SetJobResult(id, m)
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "token": token,
    })
}

// Helpers
func secureJoin(base string, parts ...string) (string, bool) {
    p := filepath.Join(append([]string{base}, parts...)...)
    ap, _ := filepath.Abs(p)
    ab, _ := filepath.Abs(base)
    if len(ap) >= len(ab) && ap[:len(ab)] == ab {
        return ap, true
    }
    return "", false
}

// Admin: list image previews grouped by archives/folders within job result
func (h *WebHandler) handleJobImages(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    id := c.Param("id")
    job, ok := GetJob(id)
    if !ok {
        c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
        return
    }
    var basePath string
    if m, ok := job.Result.(map[string]interface{}); ok {
        if p, ok := m["path"].(string); ok {
            basePath = p
        }
    }
    if basePath == "" {
        c.JSON(http.StatusOK, gin.H{"archives": []gin.H{}})
        return
    }
    
    type img struct{ Name string `json:"name"`; PreviewURL string `json:"previewURL"` }
    type archive struct {
        Name string `json:"name"`
        Path string `json:"path"`
        Images []img `json:"images"`
        Type string `json:"type"` // "zip" or "folder"
    }
    
    archives := make([]archive, 0)
    
    // First, collect all order directories (001, 002, 003, etc.)
    entries, err := os.ReadDir(basePath)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"archives": []archive{}})
        return
    }
    
    for _, entry := range entries {
        if !entry.IsDir() { continue }
        
        orderPath := filepath.Join(basePath, entry.Name())
        orderFiles, err := os.ReadDir(orderPath)
        if err != nil { continue }
        
        // Check for ZIP files in this order directory
        for _, orderFile := range orderFiles {
            if orderFile.IsDir() { continue }
            if strings.ToLower(filepath.Ext(orderFile.Name())) == ".zip" {
                zipPath := filepath.Join(orderPath, orderFile.Name())
                relZip, _ := filepath.Rel(basePath, zipPath)
                
                arch := archive{
                    Name: fmt.Sprintf("%s/%s", entry.Name(), orderFile.Name()),
                    Path: relZip,
                    Type: "zip",
                    Images: make([]img, 0),
                }
                
                // Extract images from ZIP
                zr, err := zip.OpenReader(zipPath)
                if err != nil { continue }
                
                for _, f := range zr.File {
                    ext := strings.ToLower(filepath.Ext(f.Name))
                    if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
                        arch.Images = append(arch.Images, img{
                            Name:       filepath.Base(f.Name),
                            PreviewURL: fmt.Sprintf("/api/admin/jobs/%s/preview?zip=%s&entry=%s", id, url.QueryEscape(relZip), url.QueryEscape(f.Name)),
                        })
                    }
                }
                zr.Close()
                
                if len(arch.Images) > 0 {
                    archives = append(archives, arch)
                }
            }
        }
        
        // Also check for folder-based results (if no ZIPs)
        if len(archives) == 0 || len(orderFiles) == 1 { // Only folder, no ZIPs
            folderImages := make([]img, 0)
            filepath.Walk(orderPath, func(path string, info os.FileInfo, err error) error {
                if err != nil || info.IsDir() { return nil }
                ext := strings.ToLower(filepath.Ext(path))
                if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
                    rel, _ := filepath.Rel(basePath, path)
                    folderImages = append(folderImages, img{
                        Name:       filepath.Base(path),
                        PreviewURL: fmt.Sprintf("/api/admin/jobs/%s/preview?path=%s", id, url.QueryEscape(rel)),
                    })
                }
                return nil
            })
            
            if len(folderImages) > 0 {
                archives = append(archives, archive{
                    Name: entry.Name(),
                    Path: entry.Name(),
                    Type: "folder",
                    Images: folderImages,
                })
            }
        }
    }
    
    c.JSON(http.StatusOK, gin.H{"archives": archives})
}

// Admin: stream preview (original image bytes). Accepts either ?path=rel or ?zip=zipname&entry=path
func (h *WebHandler) handleJobPreview(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    id := c.Param("id")
    job, ok := GetJob(id)
    if !ok {
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Job not found"})
        return
    }
    var basePath string
    if m, ok := job.Result.(map[string]interface{}); ok {
        if p, ok := m["path"].(string); ok {
            basePath = p
        }
    }
    if basePath == "" {
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No result path"})
        return
    }
    if rel := c.Query("path"); rel != "" {
        full, ok := secureJoin(basePath, rel)
        if !ok { c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"}); return }
        http.ServeFile(c.Writer, c.Request, full)
        return
    }
    zipName := c.Query("zip")
    entry := c.Query("entry")
    if zipName != "" && entry != "" {
        zp, ok := secureJoin(basePath, zipName)
        if !ok { c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"}); return }
        zr, err := zip.OpenReader(zp)
        if err != nil { c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to open zip"}); return }
        defer zr.Close()
        for _, f := range zr.File {
            if f.Name == entry {
                rc, err := f.Open()
                if err != nil { break }
                defer rc.Close()
                // naive content-type by extension
                ext := strings.ToLower(filepath.Ext(f.Name))
                if ext == ".png" { c.Header("Content-Type", "image/png") } else { c.Header("Content-Type", "image/jpeg") }
                io.Copy(c.Writer, rc)
                return
            }
        }
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
        return
    }
    c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
}

// Admin: stats (counts and total size)
func (h *WebHandler) handleJobStats(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    id := c.Param("id")
    job, ok := GetJob(id)
    if !ok { c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Job not found"}); return }
    var basePath string
    if m, ok := job.Result.(map[string]interface{}); ok {
        if p, ok := m["path"].(string); ok { basePath = p }
    }
    if basePath == "" { c.JSON(http.StatusOK, gin.H{"stats": gin.H{}}); return }
    var images, videos, texts, zips int
    var total int64
    filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() { return nil }
        ext := strings.ToLower(filepath.Ext(path))
        if ext == ".jpg" || ext == ".jpeg" || ext == ".png" { images++ }
        if ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".mkv" { videos++ }
        if ext == ".txt" { texts++ }
        if ext == ".zip" { zips++ }
        total += info.Size()
        return nil
    })
    c.JSON(http.StatusOK, gin.H{"stats": gin.H{
        "images": images,
        "videos": videos,
        "texts": texts,
        "zips": zips,
        "totalBytes": total,
    }})
}

// Admin: logs since job start
func (h *WebHandler) handleJobLogs(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    id := c.Param("id")
    job, ok := GetJob(id)
    if !ok { c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Job not found"}); return }
    logs := services.GetGlobalLogger().GetLogs()
    var out []services.LogEntry
    for _, le := range logs {
        if le.Timestamp.After(job.StartTime) {
            out = append(out, le)
        }
    }
    c.JSON(http.StatusOK, gin.H{"logs": out})
}