package models

import (
	"time"
)

// ProcessingJob represents a batch processing job
type ProcessingJob struct {
	ID           string    `json:"id" db:"id"`
	OrderID      string    `json:"order_id" db:"order_id"`
	SourcePath   string    `json:"source_path" db:"source_path"`
	NumCopies    int       `json:"num_copies" db:"num_copies"`
	BaseText     string    `json:"base_text" db:"base_text"`
	AddSwap      bool      `json:"add_swap" db:"add_swap"`
	AddWatermark bool      `json:"add_watermark" db:"add_watermark"`
	CreateZip    bool      `json:"create_zip" db:"create_zip"`
	WatermarkText string   `json:"watermark_text" db:"watermark_text"`
	PhotoNumber  *int      `json:"photo_number" db:"photo_number"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ProcessingStatus represents the status of a processing job
type ProcessingStatus struct {
	OrderID          string           `json:"order_id"`
	Status           string           `json:"status"`
	ProcessingLogs   []string         `json:"processing_logs"`
	SwapInfo         []SwapOperation  `json:"swap_info"`
	WatermarkPreview string           `json:"watermark_preview"`
	FileStats        ProcessingStats  `json:"file_stats"`
	Progress         float32          `json:"progress"`
}

// SwapOperation represents a file swap operation
type SwapOperation struct {
	FileA   string `json:"file_a"`
	FileB   string `json:"file_b"`
	NumberA int    `json:"number_a"`
	NumberB int    `json:"number_b"`
}

// ProcessingStats represents statistics about processed files
type ProcessingStats struct {
	TotalFiles  int    `json:"total_files"`
	Images      int    `json:"images"`
	Videos      int    `json:"videos"`
	TotalSize   int64  `json:"total_size"`
	TotalSizeMB string `json:"total_size_mb"`
}

// DownloadLink represents a temporary download link
type DownloadLink struct {
	ID            int       `json:"id" db:"id"`
	Token         string    `json:"token" db:"token"`
	OrderID       string    `json:"order_id" db:"order_id"`
	CustomerEmail string    `json:"customer_email" db:"customer_email"`
	FilePath      string    `json:"file_path" db:"file_path"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
	DownloadedAt  *time.Time `json:"downloaded_at" db:"downloaded_at"`
	DownloadCount int       `json:"download_count" db:"download_count"`
	MaxDownloads  int       `json:"max_downloads" db:"max_downloads"`
}

// WooCommerceOrder represents an order from WooCommerce webhook
type WooCommerceOrder struct {
	OrderID       string      `json:"order_id"`
	CustomerEmail string      `json:"customer_email"`
	Items         []OrderItem `json:"items"`
}

// OrderItem represents an item in a WooCommerce order
type OrderItem struct {
	ProductID    string `json:"product_id"`
	PhotoshootID string `json:"photoshoot_id"`
	Quantity     int    `json:"quantity"`
}

// TextPosition represents position for visible watermarks
type TextPosition string

const (
	TopLeft     TextPosition = "top_left"
	TopRight    TextPosition = "top_right"
	Center      TextPosition = "center"
	BottomLeft  TextPosition = "bottom_left"
	BottomRight TextPosition = "bottom_right"
)

// ProcessingJobStatus represents the status of a processing job
type ProcessingJobStatus string

const (
	StatusPending    ProcessingJobStatus = "pending"
	StatusProcessing ProcessingJobStatus = "processing"
	StatusCompleted  ProcessingJobStatus = "completed"
	StatusFailed     ProcessingJobStatus = "failed"
	StatusApproved   ProcessingJobStatus = "approved"
	StatusRejected   ProcessingJobStatus = "rejected"
)