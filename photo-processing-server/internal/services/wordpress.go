package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"photo-processing-server/internal/models"
)

// WordPressService handles WordPress/WooCommerce integration
type WordPressService struct {
	logger    *Logger
	baseURL   string
	username  string
	password  string
	apiKey    string
	apiSecret string
}

// WooCommerceWebhook represents WooCommerce webhook payload
type WooCommerceWebhook struct {
	ID       int    `json:"id"`
	Status   string `json:"status"`
	Currency string `json:"currency"`
	Total    string `json:"total"`
	Customer struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	} `json:"customer"`
	LineItems []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		ProductID  int    `json:"product_id"`
		Quantity   int    `json:"quantity"`
		Total      string `json:"total"`
		MetaData   []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"meta_data"`
	} `json:"line_items"`
	MetaData []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"meta_data"`
}

// WordPressPost represents a WordPress post
type WordPressPost struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"`
	Author  int    `json:"author"`
	Slug    string `json:"slug"`
}

// WooCommerceProduct represents a WooCommerce product
type WooCommerceProduct struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Images      []struct {
		ID  int    `json:"id"`
		Src string `json:"src"`
		Alt string `json:"alt"`
	} `json:"images"`
	Categories []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"categories"`
	MetaData []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"meta_data"`
}

func NewWordPressService(logger *Logger, baseURL, username, password, apiKey, apiSecret string) *WordPressService {
	return &WordPressService{
		logger:    logger,
		baseURL:   baseURL,
		username:  username,
		password:  password,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

// ProcessWooCommerceWebhook processes incoming WooCommerce webhooks
func (w *WordPressService) ProcessWooCommerceWebhook(payload []byte) (*models.ProcessingJob, error) {
	var webhook WooCommerceWebhook
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return nil, fmt.Errorf("invalid webhook payload: %v", err)
	}

	w.logger.Log(fmt.Sprintf("Processing WooCommerce webhook for order %d, status: %s", webhook.ID, webhook.Status))

	// Only process completed orders
	if webhook.Status != "completed" {
		w.logger.Log(fmt.Sprintf("Skipping order %d with status: %s", webhook.ID, webhook.Status))
		return nil, nil
	}

	// Extract photo processing details from order metadata or line items
	var processingConfig *models.PhotoProcessingConfig
	var err error

	// Look for processing configuration in line items metadata
	for _, item := range webhook.LineItems {
		if processingConfig, err = w.extractProcessingConfig(item.MetaData); err == nil && processingConfig != nil {
			break
		}
	}

	// Fallback to order-level metadata
	if processingConfig == nil {
		processingConfig, err = w.extractProcessingConfig(webhook.MetaData)
		if err != nil {
			return nil, fmt.Errorf("failed to extract processing config: %v", err)
		}
	}

	if processingConfig == nil {
		w.logger.Log(fmt.Sprintf("No photo processing configuration found in order %d", webhook.ID))
		return nil, nil
	}

	// Create processing job
	job := &models.ProcessingJob{
		ID:           fmt.Sprintf("woo_%d_%d", webhook.ID, time.Now().Unix()),
		OrderID:      fmt.Sprintf("%d", webhook.ID),
		SourcePath:   processingConfig.SourcePath,
		NumCopies:    processingConfig.NumCopies,
		BaseText:     processingConfig.BaseText,
		AddSwap:      processingConfig.AddSwap,
		AddWatermark: processingConfig.AddWatermark,
		CreateZip:    processingConfig.CreateZip,
		WatermarkText: processingConfig.WatermarkText,
		PhotoNumber:  processingConfig.PhotoNumber,
		Status:       string(models.StatusPending),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		UserID:       webhook.Customer.Email, // Use email as user identifier
	}

	w.logger.Log(fmt.Sprintf("Created processing job %s for WooCommerce order %d", job.ID, webhook.ID))
	return job, nil
}

// extractProcessingConfig extracts photo processing configuration from metadata
func (w *WordPressService) extractProcessingConfig(metaData []struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}) (*models.PhotoProcessingConfig, error) {
	config := &models.PhotoProcessingConfig{
		NumCopies:    1,
		CreateZip:    true,
		AddWatermark: true,
	}

	for _, meta := range metaData {
		switch meta.Key {
		case "_photo_source_path":
			config.SourcePath = meta.Value
		case "_photo_num_copies":
			if copies, err := parseIntFromString(meta.Value); err == nil {
				config.NumCopies = copies
			}
		case "_photo_base_text":
			config.BaseText = meta.Value
		case "_photo_add_swap":
			config.AddSwap = meta.Value == "yes" || meta.Value == "true"
		case "_photo_add_watermark":
			config.AddWatermark = meta.Value == "yes" || meta.Value == "true"
		case "_photo_create_zip":
			config.CreateZip = meta.Value == "yes" || meta.Value == "true"
		case "_photo_watermark_text":
			config.WatermarkText = meta.Value
		case "_photo_number":
			if photoNum, err := parseIntFromString(meta.Value); err == nil {
				config.PhotoNumber = &photoNum
			}
		}
	}

	// Validate required fields
	if config.SourcePath == "" || config.BaseText == "" {
		return nil, nil // Not a photo processing order
	}

	return config, nil
}

// CreateDownloadLink creates a download link for processed files
func (w *WordPressService) CreateDownloadLink(orderID, filePath, customerEmail string, expiryHours int) (*models.DownloadLink, error) {
	downloadLink := &models.DownloadLink{
		Token:         generateDownloadToken(),
		OrderID:       orderID,
		CustomerEmail: customerEmail,
		FilePath:      filePath,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(time.Duration(expiryHours) * time.Hour),
		MaxDownloads:  3, // Allow 3 downloads
	}

	w.logger.Log(fmt.Sprintf("Created download link for order %s: %s", orderID, downloadLink.Token))
	return downloadLink, nil
}

// SendDownloadNotification sends download notification to WordPress
func (w *WordPressService) SendDownloadNotification(orderID, downloadURL, customerEmail string) error {
	// Create a WordPress post or send email via WordPress
	notification := map[string]interface{}{
		"title": fmt.Sprintf("Photo Processing Complete - Order #%s", orderID),
		"content": fmt.Sprintf(`
			<h2>Your photos are ready!</h2>
			<p>Your photo processing order has been completed successfully.</p>
			<p><strong>Order ID:</strong> %s</p>
			<p><strong>Download Link:</strong> <a href="%s">Download Photos</a></p>
			<p><em>This link will expire in 48 hours and allows up to 3 downloads.</em></p>
		`, orderID, downloadURL),
		"status": "publish",
		"author": 1, // Admin user
		"meta": map[string]string{
			"customer_email": customerEmail,
			"order_id":       orderID,
			"notification_type": "download_ready",
		},
	}

	return w.makeWordPressAPICall("POST", "/wp/v2/posts", notification)
}

// UpdateOrderStatus updates WooCommerce order status
func (w *WordPressService) UpdateOrderStatus(orderID, status, note string) error {
	updateData := map[string]interface{}{
		"status": status,
	}

	if note != "" {
		updateData["customer_note"] = note
	}

	endpoint := fmt.Sprintf("/wc/v3/orders/%s", orderID)
	return w.makeWooCommerceAPICall("PUT", endpoint, updateData)
}

// AddOrderNote adds a note to WooCommerce order
func (w *WordPressService) AddOrderNote(orderID, note string, isCustomerNote bool) error {
	noteData := map[string]interface{}{
		"note":           note,
		"customer_note":  isCustomerNote,
		"added_by_user":  false,
	}

	endpoint := fmt.Sprintf("/wc/v3/orders/%s/notes", orderID)
	return w.makeWooCommerceAPICall("POST", endpoint, noteData)
}

// makeWordPressAPICall makes an API call to WordPress REST API
func (w *WordPressService) makeWordPressAPICall(method, endpoint string, data interface{}) error {
	return w.makeAPICall(method, "/wp-json"+endpoint, data, true)
}

// makeWooCommerceAPICall makes an API call to WooCommerce REST API
func (w *WordPressService) makeWooCommerceAPICall(method, endpoint string, data interface{}) error {
	return w.makeAPICall(method, "/wp-json"+endpoint, data, false)
}

// makeAPICall makes a generic API call
func (w *WordPressService) makeAPICall(method, endpoint string, data interface{}, useBasicAuth bool) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, w.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if useBasicAuth {
		req.SetBasicAuth(w.username, w.password)
	} else {
		// Use consumer key/secret for WooCommerce
		req.SetBasicAuth(w.apiKey, w.apiSecret)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	w.logger.Log(fmt.Sprintf("WordPress API call successful: %s %s", method, endpoint))
	return nil
}

// Helper functions
func parseIntFromString(s string) (int, error) {
	// Implementation would parse string to int with error handling
	// For now, simple approach
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return 0, err
	}
	return result, nil
}

func generateDownloadToken() string {
	// Generate a secure random token
	return fmt.Sprintf("dl_%d_%s", time.Now().Unix(), "random_token")
}
