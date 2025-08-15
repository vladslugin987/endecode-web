package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"photo-processing-server/internal/models"
	"photo-processing-server/internal/services"
)

type WooCommerceHandler struct {
	processor         *services.Processor
	logger           *services.Logger
	wpService        *services.WordPressService
	notificationService *services.NotificationService
	downloadLinks    map[string]*models.DownloadLink
}

// WooCommerce order processing request
type WooCommerceProcessRequest struct {
	OrderID       string `json:"order_id"`
	CustomerEmail string `json:"customer_email"`
	CustomerName  string `json:"customer_name"`
	ProductID     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	Quantity      int    `json:"quantity"`
	Settings      struct {
		SourceFolder          string                 `json:"source_folder"`
		NumCopies            int                    `json:"num_copies"`
		BaseText             string                 `json:"base_text"`
		AddSwap              bool                   `json:"add_swap"`
		SwapPairs            []map[string]string    `json:"swap_pairs"`
		AddWatermark         bool                   `json:"add_watermark"`
		WatermarkPositions   []int                  `json:"watermark_positions"`
		AddVisibleWatermark  bool                   `json:"add_visible_watermark"`
		VisibleWatermarkText string                 `json:"visible_watermark_text"`
		CreateZip            bool                   `json:"create_zip"`
		ZipName              string                 `json:"zip_name"`
		WatermarkText        string                 `json:"watermark_text"`
		ExpiryDays           int                    `json:"expiry_days"`
	} `json:"settings"`
}

// Timed download link structure
type TimedDownloadLink struct {
	ID           string    `json:"id"`
	Token        string    `json:"token"`
	OrderID      string    `json:"order_id"`
	FilePath     string    `json:"file_path"`
	CustomerEmail string   `json:"customer_email"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	DownloadCount int      `json:"download_count"`
	MaxDownloads int       `json:"max_downloads"`
	IsActive     bool      `json:"is_active"`
}

func NewWooCommerceHandler(processor *services.Processor, logger *services.Logger, wpService *services.WordPressService, notificationService *services.NotificationService) *WooCommerceHandler {
	return &WooCommerceHandler{
		processor:           processor,
		logger:             logger,
		wpService:          wpService,
		notificationService: notificationService,
		downloadLinks:      make(map[string]*models.DownloadLink),
	}
}

// SetupWooCommerceRoutes configures WooCommerce integration routes
func (h *WooCommerceHandler) SetupRoutes(router *gin.Engine) {
	woo := router.Group("/api/woocommerce")
	{
		woo.POST("/process-order", h.handleProcessOrder)
		woo.POST("/webhook", h.handleWebhook)
		woo.GET("/order-status/:order_id", h.handleOrderStatus)
	}

	// Timed download links
	downloads := router.Group("/api/downloads")
	{
		downloads.POST("/create", h.handleCreateTimedLink)
		downloads.GET("/link/:token", h.handleDownloadLink)
		downloads.GET("/status/:token", h.handleLinkStatus)
		downloads.DELETE("/revoke/:token", h.handleRevokeLink)
	}
}

// handleProcessOrder processes a WooCommerce order for photo processing
func (h *WooCommerceHandler) handleProcessOrder(c *gin.Context) {
	var req WooCommerceProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	h.logger.Log(fmt.Sprintf("Processing WooCommerce order %s for customer %s", req.OrderID, req.CustomerEmail))

	// Create a unique job ID
	jobID := fmt.Sprintf("woo_%s_%d", req.OrderID, time.Now().Unix())

	// Start background processing
	go h.processOrderAsync(jobID, req)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"job_id":  jobID,
		"message": "Photo processing job created successfully",
	})
}

// processOrderAsync handles the actual photo processing in the background
func (h *WooCommerceHandler) processOrderAsync(jobID string, req WooCommerceProcessRequest) {
	h.logger.Log(fmt.Sprintf("Starting background processing for job %s", jobID))

	// Send initial notification to customer
	h.notificationService.SendProcessingStatus(req.OrderID, req.CustomerEmail, "processing")

	// Create processing settings with extended options
	settings := services.BatchSettings{
		NumberOfCopies:       req.Settings.NumCopies,
		BaseText:            req.Settings.BaseText,
		AddSwapEncoding:     req.Settings.AddSwap,
		SwapPairs:           req.Settings.SwapPairs,
		AddVisibleWatermark: req.Settings.WatermarkText != "",
		WatermarkPositions:  req.Settings.WatermarkPositions,
		CreateZip:           req.Settings.CreateZip,
		WatermarkText:       req.Settings.WatermarkText,
	}

	// For demo purposes, simulate processing time
	time.Sleep(2 * time.Second)

	err := h.processor.PerformBatchCopy(req.Settings.SourceFolder, settings, func(progress float64) {
		h.logger.Log(fmt.Sprintf("Job %s progress: %.1f%%", jobID, progress*100))
	})

	if err != nil {
		h.logger.Error(fmt.Sprintf("Job %s failed: %v", jobID, err))
		// Send failure notification to customer
		h.notificationService.SendProcessingStatus(req.OrderID, req.CustomerEmail, "failed")
		return
	}

	// Send notification to admin that processing is complete and awaiting approval
	h.notificationService.SendAdminAlert(req.OrderID, jobID)
	
	// Send notification to customer that processing is complete but awaiting approval
	h.notificationService.SendProcessingStatus(req.OrderID, req.CustomerEmail, "pending_approval")

	// Send status update to WordPress
	h.notificationService.SendOrderStatusWebhook(req.OrderID, "pending_approval", "")

	// Create download link but don't activate it yet (awaiting admin approval)
	expiryDays := req.Settings.ExpiryDays
	if expiryDays <= 0 {
		expiryDays = 7 // Default to 7 days
	}

	downloadLink := &models.DownloadLink{
		Token:         generateUUID(),
		OrderID:       req.OrderID,
		CustomerEmail: req.CustomerEmail,
		FilePath:      req.Settings.SourceFolder + "-Copies", // Processed folder
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().AddDate(0, 0, expiryDays),
		MaxDownloads:  3,
		DownloadCount: 0,
		IsActive:      false, // Not active until admin approval
	}

	// Store download link (in production, use database)
	h.downloadLinks[downloadLink.Token] = downloadLink

	h.logger.Log(fmt.Sprintf("Job %s completed. Awaiting admin approval. Download link: %s (expires: %s)",
		jobID, downloadLink.Token, downloadLink.ExpiresAt.Format("2006-01-02 15:04:05")))
}

// handleWebhook processes WooCommerce webhooks
func (h *WooCommerceHandler) handleWebhook(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read webhook body"})
		return
	}

	h.logger.Log(fmt.Sprintf("Received WooCommerce webhook: %s", string(body)))

	// Parse webhook
	var webhook map[string]interface{}
	if err := json.Unmarshal(body, &webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Process webhook based on type
	if status, ok := webhook["status"].(string); ok && status == "completed" {
		if orderID, ok := webhook["id"].(float64); ok {
			h.logger.Log(fmt.Sprintf("Order %v completed, triggering photo processing", orderID))
			// In real implementation, extract order details and start processing
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// handleOrderStatus returns the status of a WooCommerce order processing
func (h *WooCommerceHandler) handleOrderStatus(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID required"})
		return
	}

	// Find download links for this order
	var links []models.DownloadLink
	for _, link := range h.downloadLinks {
		if link.OrderID == orderID {
			links = append(links, *link)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"status":   "completed",
		"links":    links,
	})
}

// Timed download link handlers

// handleCreateTimedLink creates a new timed download link
func (h *WooCommerceHandler) handleCreateTimedLink(c *gin.Context) {
	var req struct {
		OrderID       string `json:"order_id" binding:"required"`
		FilePath      string `json:"file_path" binding:"required"`
		CustomerEmail string `json:"customer_email" binding:"required"`
		ExpiryHours   int    `json:"expiry_hours"`
		MaxDownloads  int    `json:"max_downloads"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Set defaults
	if req.ExpiryHours <= 0 {
		req.ExpiryHours = 24 * 7 // 7 days default
	}
	if req.MaxDownloads <= 0 {
		req.MaxDownloads = 3
	}

	link := &models.DownloadLink{
		Token:         generateUUID(),
		OrderID:       req.OrderID,
		CustomerEmail: req.CustomerEmail,
		FilePath:      req.FilePath,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(time.Duration(req.ExpiryHours) * time.Hour),
		MaxDownloads:  req.MaxDownloads,
		DownloadCount: 0,
	}

	h.downloadLinks[link.Token] = link

	h.logger.Log(fmt.Sprintf("Created timed download link %s for order %s (expires: %s)",
		link.Token, req.OrderID, link.ExpiresAt.Format("2006-01-02 15:04:05")))

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"token":        link.Token,
		"download_url": fmt.Sprintf("/api/downloads/link/%s", link.Token),
		"expires_at":   link.ExpiresAt,
		"max_downloads": link.MaxDownloads,
	})
}

// handleDownloadLink serves files via timed download links
func (h *WooCommerceHandler) handleDownloadLink(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Download token required"})
		return
	}

	link, exists := h.downloadLinks[token]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Download link not found"})
		return
	}

	// Check if link is expired
	if time.Now().After(link.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "Download link has expired"})
		return
	}

	// Check download count
	if link.DownloadCount >= link.MaxDownloads {
		c.JSON(http.StatusGone, gin.H{"error": "Maximum downloads exceeded"})
		return
	}

	// Increment download count
	link.DownloadCount++
	h.downloadLinks[token] = link

	h.logger.Log(fmt.Sprintf("Download %d/%d for token %s (order: %s)",
		link.DownloadCount, link.MaxDownloads, token, link.OrderID))

	// In real implementation, serve the actual file
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"photos_%s.zip\"", link.OrderID))
	c.JSON(http.StatusOK, gin.H{
		"message":        "File download started",
		"remaining_downloads": link.MaxDownloads - link.DownloadCount,
		"expires_at":     link.ExpiresAt,
	})
}

// handleLinkStatus returns the status of a timed download link
func (h *WooCommerceHandler) handleLinkStatus(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Download token required"})
		return
	}

	link, exists := h.downloadLinks[token]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Download link not found"})
		return
	}

	isExpired := time.Now().After(link.ExpiresAt)
	isExhausted := link.DownloadCount >= link.MaxDownloads

	c.JSON(http.StatusOK, gin.H{
		"token":             token,
		"order_id":          link.OrderID,
		"created_at":        link.CreatedAt,
		"expires_at":        link.ExpiresAt,
		"download_count":    link.DownloadCount,
		"max_downloads":     link.MaxDownloads,
		"remaining_downloads": link.MaxDownloads - link.DownloadCount,
		"is_expired":        isExpired,
		"is_exhausted":      isExhausted,
		"is_active":         !isExpired && !isExhausted,
	})
}

// handleRevokeLink revokes a timed download link
func (h *WooCommerceHandler) handleRevokeLink(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Download token required"})
		return
	}

	if _, exists := h.downloadLinks[token]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Download link not found"})
		return
	}

	delete(h.downloadLinks, token)
	h.logger.Log(fmt.Sprintf("Revoked download link: %s", token))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Download link revoked",
	})
}
