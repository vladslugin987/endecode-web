package models

import (
	"time"
)

// Subscription represents a user subscription
type Subscription struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	PlanType  string    `json:"plan_type" db:"plan_type"` // "free", "basic", "pro", "enterprise"
	Status    string    `json:"status" db:"status"`       // "active", "expired", "cancelled", "pending"
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	
	// Payment tracking
	LastPaymentID     string     `json:"last_payment_id,omitempty" db:"last_payment_id"`
	LastPaymentDate   *time.Time `json:"last_payment_date,omitempty" db:"last_payment_date"`
	NextPaymentDate   *time.Time `json:"next_payment_date,omitempty" db:"next_payment_date"`
	AutoRenewal       bool       `json:"auto_renewal" db:"auto_renewal"`
}

// SubscriptionPlan represents available subscription plans
type SubscriptionPlan struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	PriceUSD        float64 `json:"price_usd"`
	PriceCrypto     string  `json:"price_crypto,omitempty"` // "0.001 BTC" or "50 USDT"
	DurationDays    int     `json:"duration_days"`
	MaxProcessingJobs int   `json:"max_processing_jobs"`   // -1 for unlimited
	MaxFileSize     int64   `json:"max_file_size"`         // in bytes
	Features        []string `json:"features"`
	Active          bool    `json:"active"`
}

// Payment represents a payment transaction
type Payment struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	SubscriptionID string   `json:"subscription_id" db:"subscription_id"`
	Amount        float64   `json:"amount" db:"amount"`
	Currency      string    `json:"currency" db:"currency"` // "USD", "BTC", "ETH", "USDT"
	PaymentMethod string    `json:"payment_method" db:"payment_method"` // "crypto", "card"
	Status        string    `json:"status" db:"status"` // "pending", "completed", "failed", "expired"
	
	// Crypto payment details
	CryptoAddress   string `json:"crypto_address,omitempty" db:"crypto_address"`
	CryptoAmount    string `json:"crypto_amount,omitempty" db:"crypto_amount"`
	TransactionHash string `json:"transaction_hash,omitempty" db:"transaction_hash"`
	
	// External payment gateway data
	ExternalPaymentID string `json:"external_payment_id,omitempty" db:"external_payment_id"`
	GatewayResponse   string `json:"gateway_response,omitempty" db:"gateway_response"`
	
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	PaidAt    *time.Time `json:"paid_at,omitempty" db:"paid_at"`
}

// UserUsage tracks user's monthly usage
type UserUsage struct {
	ID                string    `json:"id" db:"id"`
	UserID            string    `json:"user_id" db:"user_id"`
	Month             string    `json:"month" db:"month"` // "2025-01"
	ProcessingJobs    int       `json:"processing_jobs" db:"processing_jobs"`
	FilesProcessed    int       `json:"files_processed" db:"files_processed"`
	StorageUsedBytes  int64     `json:"storage_used_bytes" db:"storage_used_bytes"`
	LastResetDate     time.Time `json:"last_reset_date" db:"last_reset_date"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Predefined subscription plans
var DefaultSubscriptionPlans = []SubscriptionPlan{
	{
		ID:              "free",
		Name:            "Free Plan",
		Description:     "Basic photo processing with limitations",
		PriceUSD:        0,
		DurationDays:    30,
		MaxProcessingJobs: 5,
		MaxFileSize:     50 * 1024 * 1024, // 50MB
		Features: []string{
			"5 processing jobs per month",
			"Max 50MB file size",
			"Basic watermarking",
			"Email support",
		},
		Active: true,
	},
	{
		ID:              "basic",
		Name:            "Basic Plan",
		Description:     "Enhanced processing for small teams",
		PriceUSD:        29.99,
		PriceCrypto:     "0.0008 BTC",
		DurationDays:    30,
		MaxProcessingJobs: 50,
		MaxFileSize:     200 * 1024 * 1024, // 200MB
		Features: []string{
			"50 processing jobs per month",
			"Max 200MB file size",
			"Advanced watermarking",
			"Batch processing",
			"Priority email support",
		},
		Active: true,
	},
	{
		ID:              "pro",
		Name:            "Professional Plan",
		Description:     "Full-featured plan for professionals",
		PriceUSD:        89.99,
		PriceCrypto:     "0.0025 BTC",
		DurationDays:    30,
		MaxProcessingJobs: 200,
		MaxFileSize:     1024 * 1024 * 1024, // 1GB
		Features: []string{
			"200 processing jobs per month",
			"Max 1GB file size",
			"All watermarking features",
			"Advanced batch processing",
			"ZIP archive creation",
			"API access",
			"Priority support",
		},
		Active: true,
	},
	{
		ID:              "enterprise",
		Name:            "Enterprise Plan",
		Description:     "Unlimited processing for large organizations",
		PriceUSD:        299.99,
		PriceCrypto:     "0.0085 BTC",
		DurationDays:    30,
		MaxProcessingJobs: -1, // unlimited
		MaxFileSize:     5 * 1024 * 1024 * 1024, // 5GB
		Features: []string{
			"Unlimited processing jobs",
			"Max 5GB file size",
			"All features included",
			"Custom integrations",
			"WordPress/WooCommerce integration",
			"Dedicated support",
			"Custom branding",
		},
		Active: true,
	},
}

// Subscription status constants
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusExpired   = "expired"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusPending   = "pending"
)

// Payment status constants
const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusExpired   = "expired"
)

// Payment method constants
const (
	PaymentMethodCrypto = "crypto"
	PaymentMethodCard   = "card"
)

// Currency constants
const (
	CurrencyUSD  = "USD"
	CurrencyBTC  = "BTC"
	CurrencyETH  = "ETH"
	CurrencyUSDT = "USDT"
)
