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

// CryptoPaymentService handles cryptocurrency payments
type CryptoPaymentService struct {
	logger  *Logger
	subsService *SubscriptionService
	apiURL  string // CoinGate API URL
	apiKey  string // CoinGate API key
}

// CoinGate API structures
type CoinGateCreateOrderRequest struct {
	OrderID     string  `json:"order_id"`
	PriceAmount float64 `json:"price_amount"`
	PriceCurrency string `json:"price_currency"`
	ReceiveCurrency string `json:"receive_currency,omitempty"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	CallbackURL string  `json:"callback_url,omitempty"`
	SuccessURL  string  `json:"success_url,omitempty"`
	CancelURL   string  `json:"cancel_url,omitempty"`
}

type CoinGateOrderResponse struct {
	ID                int     `json:"id"`
	Status            string  `json:"status"`
	PriceAmount       string  `json:"price_amount"`
	PriceCurrency     string  `json:"price_currency"`
	ReceiveCurrency   string  `json:"receive_currency"`
	ReceiveAmount     string  `json:"receive_amount"`
	PaymentURL        string  `json:"payment_url"`
	PaymentAddress    string  `json:"payment_address"`
	OrderID           string  `json:"order_id"`
	Token             string  `json:"token"`
	CreatedAt         string  `json:"created_at"`
	ExpiresAt         string  `json:"expires_at"`
}

type CoinGateWebhookPayload struct {
	ID                int     `json:"id"`
	Status            string  `json:"status"`
	PriceAmount       string  `json:"price_amount"`
	PriceCurrency     string  `json:"price_currency"`
	ReceiveCurrency   string  `json:"receive_currency"`
	ReceiveAmount     string  `json:"receive_amount"`
	PaymentAddress    string  `json:"payment_address"`
	OrderID           string  `json:"order_id"`
	Token             string  `json:"token"`
	TransactionHash   string  `json:"transaction_hash,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

func NewCryptoPaymentService(logger *Logger, subsService *SubscriptionService) *CryptoPaymentService {
	return &CryptoPaymentService{
		logger:      logger,
		subsService: subsService,
		apiURL:      "https://api.coingate.com/v2", // Production API
		apiKey:      "", // Will be set from environment or config
	}
}

// SetAPIKey sets the CoinGate API key
func (c *CryptoPaymentService) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// CreateCryptoPayment creates a new cryptocurrency payment via CoinGate
func (c *CryptoPaymentService) CreateCryptoPayment(userID, planType string, receiveCurrency string) (*models.Payment, string, error) {
	// Get plan details
	plan := c.subsService.GetPlanByType(planType)
	if plan == nil {
		return nil, "", fmt.Errorf("invalid plan type: %s", planType)
	}

	// Create subscription record first
	subscription, err := c.subsService.GetUserSubscription(userID)
	if err != nil {
		return nil, "", err
	}

	// Create payment record
	payment, err := c.subsService.CreatePayment(
		userID,
		subscription.ID,
		plan.PriceUSD,
		receiveCurrency,
		models.PaymentMethodCrypto,
	)
	if err != nil {
		return nil, "", err
	}

	// If no API key, return mock payment for testing
	if c.apiKey == "" {
		c.logger.Log("Warning: No CoinGate API key set, creating mock crypto payment")
		return c.createMockCryptoPayment(payment, plan, receiveCurrency)
	}

	// Create CoinGate order
	orderRequest := CoinGateCreateOrderRequest{
		OrderID:         payment.ID,
		PriceAmount:     plan.PriceUSD,
		PriceCurrency:   "USD",
		ReceiveCurrency: receiveCurrency,
		Title:           fmt.Sprintf("ENDECode %s Subscription", plan.Name),
		Description:     fmt.Sprintf("1 month %s plan subscription", plan.Name),
		CallbackURL:     "http://localhost:8090/api/payments/crypto/webhook", // This should be your public URL
		SuccessURL:      "http://localhost:8090/dashboard?payment=success",
		CancelURL:       "http://localhost:8090/dashboard?payment=cancel",
	}

	orderResponse, err := c.createCoinGateOrder(orderRequest)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create CoinGate order: %v", err)
	}

	// Update payment with CoinGate details
	payment.ExternalPaymentID = fmt.Sprintf("%d", orderResponse.ID)
	payment.CryptoAddress = orderResponse.PaymentAddress
	payment.CryptoAmount = orderResponse.ReceiveAmount
	payment.Currency = orderResponse.ReceiveCurrency

	// Save updated payment
	if err := c.subsService.SavePayment(payment); err != nil {
		c.logger.Error(fmt.Sprintf("Failed to save payment: %v", err))
	}

	c.logger.Log(fmt.Sprintf("Created crypto payment: %s for user %s, amount: %s %s", 
		payment.ID, userID, orderResponse.ReceiveAmount, orderResponse.ReceiveCurrency))

	return payment, orderResponse.PaymentURL, nil
}

// createCoinGateOrder creates an order with CoinGate API
func (c *CryptoPaymentService) createCoinGateOrder(request CoinGateCreateOrderRequest) (*CoinGateOrderResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.apiURL+"/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+c.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("CoinGate API error: %d - %s", resp.StatusCode, string(body))
	}

	var orderResponse CoinGateOrderResponse
	if err := json.Unmarshal(body, &orderResponse); err != nil {
		return nil, err
	}

	return &orderResponse, nil
}

// ProcessWebhook processes CoinGate webhook notifications
func (c *CryptoPaymentService) ProcessWebhook(payload []byte) error {
	var webhook CoinGateWebhookPayload
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return fmt.Errorf("invalid webhook payload: %v", err)
	}

	c.logger.Log(fmt.Sprintf("Processing crypto webhook for order: %s, status: %s", webhook.OrderID, webhook.Status))

	// Get payment by order ID
	payment, err := c.subsService.GetPayment(webhook.OrderID)
	if err != nil {
		return fmt.Errorf("payment not found: %s", webhook.OrderID)
	}

	// Update payment status based on webhook status
	switch webhook.Status {
	case "paid":
		payment.Status = models.PaymentStatusCompleted
		payment.TransactionHash = webhook.TransactionHash
		now := time.Now()
		payment.PaidAt = &now
		payment.UpdatedAt = now

		// Activate subscription
		if err := c.activateSubscription(payment); err != nil {
			c.logger.Error(fmt.Sprintf("Failed to activate subscription: %v", err))
			return err
		}

	case "expired", "canceled":
		payment.Status = models.PaymentStatusExpired
		payment.UpdatedAt = time.Now()

	case "invalid", "failed":
		payment.Status = models.PaymentStatusFailed
		payment.UpdatedAt = time.Now()
	}

	// Save updated payment
	if err := c.subsService.SavePayment(payment); err != nil {
		return fmt.Errorf("failed to save payment: %v", err)
	}

	c.logger.Log(fmt.Sprintf("Updated payment %s status to: %s", payment.ID, payment.Status))
	return nil
}

// activateSubscription activates user subscription after successful payment
func (c *CryptoPaymentService) activateSubscription(payment *models.Payment) error {
	subscription, err := c.subsService.GetUserSubscription(payment.UserID)
	if err != nil {
		return err
	}

	// Get plan details to determine duration
	plan := c.subsService.GetPlanByType(subscription.PlanType)
	if plan == nil {
		return fmt.Errorf("invalid plan type: %s", subscription.PlanType)
	}

	now := time.Now()
	
	// Extend subscription
	if subscription.EndDate.Before(now) {
		subscription.StartDate = now
		subscription.EndDate = now.AddDate(0, 0, plan.DurationDays)
	} else {
		subscription.EndDate = subscription.EndDate.AddDate(0, 0, plan.DurationDays)
	}

	subscription.Status = models.SubscriptionStatusActive
	subscription.LastPaymentID = payment.ID
	subscription.LastPaymentDate = payment.PaidAt
	subscription.UpdatedAt = now

	if err := c.subsService.SaveSubscription(subscription); err != nil {
		return err
	}

	c.logger.Log(fmt.Sprintf("Activated subscription for user %s: %s plan until %s", 
		payment.UserID, subscription.PlanType, subscription.EndDate.Format("2006-01-02")))

	return nil
}

// createMockCryptoPayment creates a mock payment for testing without real API
func (c *CryptoPaymentService) createMockCryptoPayment(payment *models.Payment, plan *models.SubscriptionPlan, currency string) (*models.Payment, string, error) {
	// Generate mock crypto address and amount
	var mockAddress, mockAmount string
	
	switch currency {
	case "BTC":
		mockAddress = "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh"
		mockAmount = "0.001"
	case "ETH":
		mockAddress = "0x742d35Cc6634C0532925a3b8D4C09644e3f4C1a7"
		mockAmount = "0.02"
	case "USDT":
		mockAddress = "TYDzsYUEpvnYmQk4zGP9sWWcTEd2MiAtW6"
		mockAmount = fmt.Sprintf("%.2f", plan.PriceUSD)
	default:
		mockAddress = "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh"
		mockAmount = "0.001"
	}

	payment.CryptoAddress = mockAddress
	payment.CryptoAmount = mockAmount
	payment.Currency = currency
	payment.ExternalPaymentID = "mock_" + payment.ID

	if err := c.subsService.SavePayment(payment); err != nil {
		return nil, "", err
	}

	// Create mock payment URL
	mockPaymentURL := fmt.Sprintf("http://localhost:8090/payment/mock/%s", payment.ID)

	c.logger.Log(fmt.Sprintf("Created mock crypto payment: %s for user %s, amount: %s %s", 
		payment.ID, payment.UserID, mockAmount, currency))

	return payment, mockPaymentURL, nil
}

// GetPaymentStatus returns the current status of a payment
func (c *CryptoPaymentService) GetPaymentStatus(paymentID string) (*models.Payment, error) {
	payment, err := c.subsService.GetPayment(paymentID)
	if err != nil {
		return nil, err
	}

	// If payment is pending and has CoinGate ID, check status
	if payment.Status == models.PaymentStatusPending && payment.ExternalPaymentID != "" && c.apiKey != "" {
		if err := c.updatePaymentStatusFromCoinGate(payment); err != nil {
			c.logger.Error(fmt.Sprintf("Failed to update payment status from CoinGate: %v", err))
		}
	}

	return payment, nil
}

// updatePaymentStatusFromCoinGate fetches latest status from CoinGate API
func (c *CryptoPaymentService) updatePaymentStatusFromCoinGate(payment *models.Payment) error {
	if payment.ExternalPaymentID == "" || c.apiKey == "" {
		return nil
	}

	req, err := http.NewRequest("GET", c.apiURL+"/orders/"+payment.ExternalPaymentID, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+c.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CoinGate API error: %d", resp.StatusCode)
	}

	var orderResponse CoinGateOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResponse); err != nil {
		return err
	}

	// Update payment status
	oldStatus := payment.Status
	switch orderResponse.Status {
	case "paid":
		payment.Status = models.PaymentStatusCompleted
		now := time.Now()
		payment.PaidAt = &now
		c.activateSubscription(payment)
	case "expired", "canceled":
		payment.Status = models.PaymentStatusExpired
	case "invalid", "failed":
		payment.Status = models.PaymentStatusFailed
	}

	if payment.Status != oldStatus {
		payment.UpdatedAt = time.Now()
		c.subsService.SavePayment(payment)
		c.logger.Log(fmt.Sprintf("Updated payment %s status from %s to %s", payment.ID, oldStatus, payment.Status))
	}

	return nil
}

// GetSupportedCurrencies returns list of supported cryptocurrencies
func (c *CryptoPaymentService) GetSupportedCurrencies() []map[string]string {
	return []map[string]string{
		{"code": "BTC", "name": "Bitcoin", "symbol": "₿"},
		{"code": "ETH", "name": "Ethereum", "symbol": "Ξ"},
		{"code": "USDT", "name": "Tether", "symbol": "₮"},
		{"code": "LTC", "name": "Litecoin", "symbol": "Ł"},
		{"code": "BCH", "name": "Bitcoin Cash", "symbol": "BCH"},
	}
}
