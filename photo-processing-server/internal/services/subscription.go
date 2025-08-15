package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"photo-processing-server/internal/models"
	"github.com/google/uuid"
	"github.com/go-redis/redis/v8"
)

type SubscriptionService struct {
	logger      *Logger
	redisClient *redis.Client
}

func NewSubscriptionService(logger *Logger, redisClient *redis.Client) *SubscriptionService {
	return &SubscriptionService{
		logger:      logger,
		redisClient: redisClient,
	}
}

// GetUserSubscription returns the current subscription for a user
func (s *SubscriptionService) GetUserSubscription(userID string) (*models.Subscription, error) {
	if s.redisClient == nil {
		return s.getDefaultSubscription(userID), nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("subscription:%s", userID)
	
	data, err := s.redisClient.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		// Create default free subscription
		return s.createDefaultSubscription(userID)
	}

	subscription := &models.Subscription{}
	subscription.ID = data["id"]
	subscription.UserID = data["user_id"]
	subscription.PlanType = data["plan_type"]
	subscription.Status = data["status"]
	
	if startDate, err := time.Parse(time.RFC3339, data["start_date"]); err == nil {
		subscription.StartDate = startDate
	}
	if endDate, err := time.Parse(time.RFC3339, data["end_date"]); err == nil {
		subscription.EndDate = endDate
	}
	if createdAt, err := time.Parse(time.RFC3339, data["created_at"]); err == nil {
		subscription.CreatedAt = createdAt
	}
	if updatedAt, err := time.Parse(time.RFC3339, data["updated_at"]); err == nil {
		subscription.UpdatedAt = updatedAt
	}

	subscription.LastPaymentID = data["last_payment_id"]
	if lastPaymentDate := data["last_payment_date"]; lastPaymentDate != "" {
		if t, err := time.Parse(time.RFC3339, lastPaymentDate); err == nil {
			subscription.LastPaymentDate = &t
		}
	}
	if nextPaymentDate := data["next_payment_date"]; nextPaymentDate != "" {
		if t, err := time.Parse(time.RFC3339, nextPaymentDate); err == nil {
			subscription.NextPaymentDate = &t
		}
	}
	if autoRenewal := data["auto_renewal"]; autoRenewal == "true" {
		subscription.AutoRenewal = true
	}

	return subscription, nil
}

// SaveSubscription saves a subscription to Redis
func (s *SubscriptionService) SaveSubscription(subscription *models.Subscription) error {
	if s.redisClient == nil {
		s.logger.Log("Warning: Redis not available, subscription not saved")
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("subscription:%s", subscription.UserID)
	
	data := map[string]interface{}{
		"id":         subscription.ID,
		"user_id":    subscription.UserID,
		"plan_type":  subscription.PlanType,
		"status":     subscription.Status,
		"start_date": subscription.StartDate.Format(time.RFC3339),
		"end_date":   subscription.EndDate.Format(time.RFC3339),
		"created_at": subscription.CreatedAt.Format(time.RFC3339),
		"updated_at": subscription.UpdatedAt.Format(time.RFC3339),
		"auto_renewal": fmt.Sprintf("%v", subscription.AutoRenewal),
	}

	if subscription.LastPaymentID != "" {
		data["last_payment_id"] = subscription.LastPaymentID
	}
	if subscription.LastPaymentDate != nil {
		data["last_payment_date"] = subscription.LastPaymentDate.Format(time.RFC3339)
	}
	if subscription.NextPaymentDate != nil {
		data["next_payment_date"] = subscription.NextPaymentDate.Format(time.RFC3339)
	}

	return s.redisClient.HSet(ctx, key, data).Err()
}

// ExtendSubscription extends a user's subscription by the specified number of days
func (s *SubscriptionService) ExtendSubscription(userID string, planType string, days int, adminUserID string) error {
	subscription, err := s.GetUserSubscription(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	
	// If subscription is expired or this is a plan change, start from now
	if subscription.Status == models.SubscriptionStatusExpired || subscription.PlanType != planType {
		subscription.StartDate = now
		subscription.EndDate = now.AddDate(0, 0, days)
	} else {
		// Extend current subscription
		if subscription.EndDate.Before(now) {
			subscription.EndDate = now.AddDate(0, 0, days)
		} else {
			subscription.EndDate = subscription.EndDate.AddDate(0, 0, days)
		}
	}

	subscription.PlanType = planType
	subscription.Status = models.SubscriptionStatusActive
	subscription.UpdatedAt = now

	err = s.SaveSubscription(subscription)
	if err != nil {
		return err
	}

	s.logger.Log(fmt.Sprintf("Admin %s extended user %s subscription: %s plan for %d days", 
		adminUserID, userID, planType, days))

	return nil
}

// CheckSubscriptionLimits checks if user can perform an action based on their subscription
func (s *SubscriptionService) CheckSubscriptionLimits(userID string, actionType string) (bool, string, error) {
	subscription, err := s.GetUserSubscription(userID)
	if err != nil {
		return false, "Failed to get subscription", err
	}

	// Check if subscription is active
	if subscription.Status != models.SubscriptionStatusActive {
		return false, "Subscription is not active", nil
	}

	// Check if subscription has expired
	if subscription.EndDate.Before(time.Now()) {
		// Update status to expired
		subscription.Status = models.SubscriptionStatusExpired
		subscription.UpdatedAt = time.Now()
		s.SaveSubscription(subscription)
		return false, "Subscription has expired", nil
	}

	// Get plan details
	plan := s.GetPlanByType(subscription.PlanType)
	if plan == nil {
		return false, "Invalid subscription plan", nil
	}

	// Check usage limits based on action type
	switch actionType {
	case "processing_job":
		if plan.MaxProcessingJobs == -1 {
			return true, "", nil // unlimited
		}
		
		usage, err := s.GetUserUsage(userID)
		if err != nil {
			return false, "Failed to get usage data", err
		}
		
		if usage.ProcessingJobs >= plan.MaxProcessingJobs {
			return false, fmt.Sprintf("Monthly limit reached: %d/%d processing jobs", 
				usage.ProcessingJobs, plan.MaxProcessingJobs), nil
		}
		
	case "file_upload":
		// File size limits are checked during upload
		return true, "", nil
	}

	return true, "", nil
}

// IncrementUsage increments user's usage counter
func (s *SubscriptionService) IncrementUsage(userID string, usageType string, amount int) error {
	usage, err := s.GetUserUsage(userID)
	if err != nil {
		return err
	}

	switch usageType {
	case "processing_jobs":
		usage.ProcessingJobs += amount
	case "files_processed":
		usage.FilesProcessed += amount
	}

	usage.UpdatedAt = time.Now()
	return s.SaveUserUsage(usage)
}

// GetUserUsage returns user's current month usage
func (s *SubscriptionService) GetUserUsage(userID string) (*models.UserUsage, error) {
	if s.redisClient == nil {
		return s.getDefaultUsage(userID), nil
	}

	ctx := context.Background()
	currentMonth := time.Now().Format("2006-01")
	key := fmt.Sprintf("usage:%s:%s", userID, currentMonth)
	
	data, err := s.redisClient.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		usage, err := s.createDefaultUsage(userID)
		if err != nil {
			return nil, err
		}
		return usage, nil
	}

	usage := &models.UserUsage{
		ID:     data["id"],
		UserID: data["user_id"],
		Month:  data["month"],
	}

	if processingJobs, err := strconv.Atoi(data["processing_jobs"]); err == nil {
		usage.ProcessingJobs = processingJobs
	}
	if filesProcessed, err := strconv.Atoi(data["files_processed"]); err == nil {
		usage.FilesProcessed = filesProcessed
	}
	if storageUsed, err := strconv.ParseInt(data["storage_used_bytes"], 10, 64); err == nil {
		usage.StorageUsedBytes = storageUsed
	}
	if lastResetDate, err := time.Parse(time.RFC3339, data["last_reset_date"]); err == nil {
		usage.LastResetDate = lastResetDate
	}
	if createdAt, err := time.Parse(time.RFC3339, data["created_at"]); err == nil {
		usage.CreatedAt = createdAt
	}
	if updatedAt, err := time.Parse(time.RFC3339, data["updated_at"]); err == nil {
		usage.UpdatedAt = updatedAt
	}

	return usage, nil
}

// SaveUserUsage saves user usage to Redis
func (s *SubscriptionService) SaveUserUsage(usage *models.UserUsage) error {
	if s.redisClient == nil {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("usage:%s:%s", usage.UserID, usage.Month)
	
	data := map[string]interface{}{
		"id":                 usage.ID,
		"user_id":           usage.UserID,
		"month":             usage.Month,
		"processing_jobs":   strconv.Itoa(usage.ProcessingJobs),
		"files_processed":   strconv.Itoa(usage.FilesProcessed),
		"storage_used_bytes": strconv.FormatInt(usage.StorageUsedBytes, 10),
		"last_reset_date":   usage.LastResetDate.Format(time.RFC3339),
		"created_at":        usage.CreatedAt.Format(time.RFC3339),
		"updated_at":        usage.UpdatedAt.Format(time.RFC3339),
	}

	return s.redisClient.HSet(ctx, key, data).Err()
}

// GetPlanByType returns a subscription plan by type
func (s *SubscriptionService) GetPlanByType(planType string) *models.SubscriptionPlan {
	for _, plan := range models.DefaultSubscriptionPlans {
		if plan.ID == planType {
			return &plan
		}
	}
	return nil
}

// GetAllPlans returns all available subscription plans
func (s *SubscriptionService) GetAllPlans() []models.SubscriptionPlan {
	return models.DefaultSubscriptionPlans
}

// CreatePayment creates a new payment record
func (s *SubscriptionService) CreatePayment(userID, subscriptionID string, amount float64, currency, paymentMethod string) (*models.Payment, error) {
	payment := &models.Payment{
		ID:             uuid.New().String(),
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Amount:         amount,
		Currency:       currency,
		PaymentMethod:  paymentMethod,
		Status:         models.PaymentStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Set expiration for crypto payments (30 minutes)
	if paymentMethod == models.PaymentMethodCrypto {
		expiry := time.Now().Add(30 * time.Minute)
		payment.ExpiresAt = &expiry
	}

	if err := s.SavePayment(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

// SavePayment saves a payment to Redis
func (s *SubscriptionService) SavePayment(payment *models.Payment) error {
	if s.redisClient == nil {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("payment:%s", payment.ID)
	
	paymentData, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, paymentData, 24*time.Hour).Err()
}

// GetPayment retrieves a payment by ID
func (s *SubscriptionService) GetPayment(paymentID string) (*models.Payment, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("payment storage not available")
	}

	ctx := context.Background()
	key := fmt.Sprintf("payment:%s", paymentID)
	
	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var payment models.Payment
	if err := json.Unmarshal([]byte(data), &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

// Helper functions for default values

func (s *SubscriptionService) getDefaultSubscription(userID string) *models.Subscription {
	now := time.Now()
	return &models.Subscription{
		ID:        uuid.New().String(),
		UserID:    userID,
		PlanType:  "free",
		Status:    models.SubscriptionStatusActive,
		StartDate: now,
		EndDate:   now.AddDate(0, 1, 0), // 1 month
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s *SubscriptionService) createDefaultSubscription(userID string) (*models.Subscription, error) {
	subscription := s.getDefaultSubscription(userID)
	if err := s.SaveSubscription(subscription); err != nil {
		return nil, err
	}
	return subscription, nil
}

func (s *SubscriptionService) getDefaultUsage(userID string) *models.UserUsage {
	now := time.Now()
	return &models.UserUsage{
		ID:               uuid.New().String(),
		UserID:           userID,
		Month:            now.Format("2006-01"),
		ProcessingJobs:   0,
		FilesProcessed:   0,
		StorageUsedBytes: 0,
		LastResetDate:    now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (s *SubscriptionService) createDefaultUsage(userID string) (*models.UserUsage, error) {
	usage := s.getDefaultUsage(userID)
	if err := s.SaveUserUsage(usage); err != nil {
		return nil, err
	}
	return usage, nil
}

// Admin functions

// GetAllSubscriptions returns all subscriptions (admin only)
func (s *SubscriptionService) GetAllSubscriptions() ([]models.Subscription, error) {
	if s.redisClient == nil {
		return []models.Subscription{}, nil
	}

	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, "subscription:*").Result()
	if err != nil {
		return nil, err
	}

	var subscriptions []models.Subscription
	for _, key := range keys {
		userID := strings.TrimPrefix(key, "subscription:")
		if sub, err := s.GetUserSubscription(userID); err == nil {
			subscriptions = append(subscriptions, *sub)
		}
	}

	return subscriptions, nil
}

// GetSubscriptionStats returns subscription statistics (admin only)
func (s *SubscriptionService) GetSubscriptionStats() (map[string]interface{}, error) {
	subscriptions, err := s.GetAllSubscriptions()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	planCounts := make(map[string]int)
	statusCounts := make(map[string]int)
	totalRevenue := 0.0

	for _, sub := range subscriptions {
		planCounts[sub.PlanType]++
		statusCounts[sub.Status]++
		
		// Calculate approximate revenue (this is simplified)
		if plan := s.GetPlanByType(sub.PlanType); plan != nil {
			totalRevenue += plan.PriceUSD
		}
	}

	stats["total_subscriptions"] = len(subscriptions)
	stats["plan_distribution"] = planCounts
	stats["status_distribution"] = statusCounts
	stats["estimated_monthly_revenue"] = totalRevenue

	return stats, nil
}
