package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"photo-processing-server/internal/services"
)

type SubscriptionHandler struct {
	subsService   *services.SubscriptionService
	cryptoService *services.CryptoPaymentService
	logger        *services.Logger
}

func NewSubscriptionHandler(subsService *services.SubscriptionService, cryptoService *services.CryptoPaymentService, logger *services.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		subsService:   subsService,
		cryptoService: cryptoService,
		logger:        logger,
	}
}

// SetupSubscriptionRoutes configures subscription-related routes
func (h *SubscriptionHandler) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/subscription")
	{
		// User subscription routes
		api.GET("/my", h.handleGetMySubscription)
		api.GET("/plans", h.handleGetPlans)
		api.GET("/usage", h.handleGetMyUsage)
		
		// Payment routes
		api.POST("/payment/crypto", h.handleCreateCryptoPayment)
		api.GET("/payment/:id", h.handleGetPaymentStatus)
		api.POST("/payment/mock/:id/complete", h.handleMockPaymentComplete) // For testing
	}

	// Admin routes
	admin := router.Group("/api/admin/subscription")
	admin.Use(requireAdminAuth())
	{
		admin.GET("/all", h.handleGetAllSubscriptions)
		admin.GET("/stats", h.handleGetSubscriptionStats)
		admin.POST("/extend", h.handleExtendSubscription)
		admin.GET("/payments", h.handleGetAllPayments)
	}

	// Webhook routes (no auth required)
	router.POST("/api/payments/crypto/webhook", h.handleCryptoWebhook)
}

// handleGetMySubscription returns current user's subscription
func (h *SubscriptionHandler) handleGetMySubscription(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}

	subscription, err := h.subsService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
		return
	}

	// Get plan details
	plan := h.subsService.GetPlanByType(subscription.PlanType)
	
	response := gin.H{
		"subscription": subscription,
		"plan":         plan,
	}

	c.JSON(http.StatusOK, response)
}

// handleGetPlans returns all available subscription plans
func (h *SubscriptionHandler) handleGetPlans(c *gin.Context) {
	plans := h.subsService.GetAllPlans()
	
	// Add supported currencies for crypto payments
	currencies := h.cryptoService.GetSupportedCurrencies()
	
	c.JSON(http.StatusOK, gin.H{
		"plans":       plans,
		"currencies":  currencies,
	})
}

// handleGetMyUsage returns current user's usage statistics
func (h *SubscriptionHandler) handleGetMyUsage(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}

	usage, err := h.subsService.GetUserUsage(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage"})
		return
	}

	subscription, err := h.subsService.GetUserSubscription(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
		return
	}

	plan := h.subsService.GetPlanByType(subscription.PlanType)
	
	response := gin.H{
		"usage": usage,
		"limits": gin.H{
			"processing_jobs": plan.MaxProcessingJobs,
			"max_file_size":   plan.MaxFileSize,
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleCreateCryptoPayment creates a new cryptocurrency payment
func (h *SubscriptionHandler) handleCreateCryptoPayment(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}

	var req struct {
		PlanType string `json:"plan_type" binding:"required"`
		Currency string `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate plan type
	plan := h.subsService.GetPlanByType(req.PlanType)
	if plan == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan type"})
		return
	}

	// Validate currency
	supportedCurrencies := h.cryptoService.GetSupportedCurrencies()
	validCurrency := false
	for _, curr := range supportedCurrencies {
		if curr["code"] == req.Currency {
			validCurrency = true
			break
		}
	}
	if !validCurrency {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported currency"})
		return
	}

	// Create payment
	payment, paymentURL, err := h.cryptoService.CreateCryptoPayment(userID, req.PlanType, req.Currency)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to create crypto payment: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_id":     payment.ID,
		"payment_url":    paymentURL,
		"crypto_address": payment.CryptoAddress,
		"crypto_amount":  payment.CryptoAmount,
		"currency":       payment.Currency,
		"expires_at":     payment.ExpiresAt,
		"status":         payment.Status,
	})
}

// handleGetPaymentStatus returns the status of a payment
func (h *SubscriptionHandler) handleGetPaymentStatus(c *gin.Context) {
	userID := getCurrentUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}

	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID required"})
		return
	}

	payment, err := h.cryptoService.GetPaymentStatus(paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Verify payment belongs to user
	if payment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_id":     payment.ID,
		"status":         payment.Status,
		"amount":         payment.Amount,
		"currency":       payment.Currency,
		"crypto_address": payment.CryptoAddress,
		"crypto_amount":  payment.CryptoAmount,
		"created_at":     payment.CreatedAt,
		"expires_at":     payment.ExpiresAt,
		"paid_at":        payment.PaidAt,
	})
}

// handleMockPaymentComplete completes a mock payment for testing
func (h *SubscriptionHandler) handleMockPaymentComplete(c *gin.Context) {
	if !strings.Contains(c.Request.Host, "localhost") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Mock payments only available in development"})
		return
	}

	userID := getCurrentUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}

	paymentID := c.Param("id")
	payment, err := h.subsService.GetPayment(paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	if payment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Simulate successful payment
	mockWebhook := fmt.Sprintf(`{
		"id": 12345,
		"status": "paid",
		"order_id": "%s",
		"transaction_hash": "mock_hash_123456789"
	}`, paymentID)

	if err := h.cryptoService.ProcessWebhook([]byte(mockWebhook)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mock payment completed successfully",
	})
}

// handleCryptoWebhook processes cryptocurrency payment webhooks
func (h *SubscriptionHandler) handleCryptoWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read webhook body"})
		return
	}

	if err := h.cryptoService.ProcessWebhook(body); err != nil {
		h.logger.Error(fmt.Sprintf("Webhook processing failed: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook processing failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Admin handlers

// handleGetAllSubscriptions returns all subscriptions (admin only)
func (h *SubscriptionHandler) handleGetAllSubscriptions(c *gin.Context) {
	subscriptions, err := h.subsService.GetAllSubscriptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
}

// handleGetSubscriptionStats returns subscription statistics (admin only)
func (h *SubscriptionHandler) handleGetSubscriptionStats(c *gin.Context) {
	stats, err := h.subsService.GetSubscriptionStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// handleExtendSubscription allows admin to extend user subscription
func (h *SubscriptionHandler) handleExtendSubscription(c *gin.Context) {
	adminUserID := getCurrentUserID(c)
	if adminUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin login required"})
		return
	}

	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		PlanType string `json:"plan_type" binding:"required"`
		Days     int    `json:"days" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate plan type
	if h.subsService.GetPlanByType(req.PlanType) == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan type"})
		return
	}

	// Validate days (1-365)
	if req.Days < 1 || req.Days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Days must be between 1 and 365"})
		return
	}

	if err := h.subsService.ExtendSubscription(req.UserID, req.PlanType, req.Days, adminUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extend subscription"})
		return
	}

	// Get updated subscription
	subscription, err := h.subsService.GetUserSubscription(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("Extended %s subscription by %d days", req.PlanType, req.Days),
		"subscription": subscription,
	})
}

// handleGetAllPayments returns all payments (admin only)
func (h *SubscriptionHandler) handleGetAllPayments(c *gin.Context) {
	// This would require a method to get all payments from Redis
	// For now, return empty array
	c.JSON(http.StatusOK, gin.H{
		"payments": []interface{}{},
		"message":  "Payment history not implemented yet",
	})
}

// Middleware to check subscription limits before processing
func (h *SubscriptionHandler) CheckSubscriptionLimits(actionType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getCurrentUserID(c)
		if userID == "" {
			c.Next() // Let auth middleware handle this
			return
		}

		allowed, message, err := h.subsService.CheckSubscriptionLimits(userID, actionType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check subscription limits"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Subscription limit exceeded",
				"message": message,
				"upgrade_required": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function to increment usage after successful operations
func (h *SubscriptionHandler) IncrementUsage(userID, usageType string, amount int) {
	if err := h.subsService.IncrementUsage(userID, usageType, amount); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to increment usage for user %s: %v", userID, err))
	}
}
