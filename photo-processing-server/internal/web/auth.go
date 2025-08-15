package web

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Simple user model stored in Redis or memory

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Password  string    `json:"-"`
	Role      string    `json:"role"` // "user", "admin"
	CreatedAt time.Time `json:"createdAt"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`
}

var (
	usersByEmail = map[string]*User{}
	sessions     = map[string]string{} // sessionToken -> userID
)

func userKey(email string) string   { return "user:" + strings.ToLower(email) }
func sessionKey(token string) string { return "sess:" + token }

func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil { return "", err }
	return string(b), nil
}

func checkPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func getUserByEmail(email string) (*User, bool) {
	// Redis preferred
	if redisClient != nil {
		ctx := context.Background()
		if s, err := redisClient.HGet(ctx, userKey(email), "id").Result(); err == nil && s != "" {
			emailLower := strings.ToLower(email)
			u := &User{ID: s, Email: emailLower}
			if ph, err := redisClient.HGet(ctx, userKey(emailLower), "pw").Result(); err == nil { u.Password = ph }
			if nickname, err := redisClient.HGet(ctx, userKey(emailLower), "nickname").Result(); err == nil { u.Nickname = nickname }
			if role, err := redisClient.HGet(ctx, userKey(emailLower), "role").Result(); err == nil { u.Role = role }
			if ct, err := redisClient.HGet(ctx, userKey(emailLower), "createdAt").Result(); err == nil { if t, e := time.Parse(time.RFC3339, ct); e == nil { u.CreatedAt = t } }
			if ll, err := redisClient.HGet(ctx, userKey(emailLower), "lastLogin").Result(); err == nil && ll != "" { if t, e := time.Parse(time.RFC3339, ll); e == nil { u.LastLogin = &t } }
			return u, true
		}
	}
	u, ok := usersByEmail[strings.ToLower(email)]
	return u, ok
}

func saveUser(u *User) error {
	if redisClient != nil {
		ctx := context.Background()
		emailLower := strings.ToLower(u.Email)
		data := map[string]interface{}{
			"id":        u.ID,
			"email":     emailLower,
			"nickname":  u.Nickname,
			"pw":        u.Password,
			"role":      u.Role,
			"createdAt": u.CreatedAt.Format(time.RFC3339),
		}
		if u.LastLogin != nil {
			data["lastLogin"] = u.LastLogin.Format(time.RFC3339)
		}
		if err := redisClient.HSet(ctx, userKey(emailLower), data).Err(); err != nil {
			return err
		}
	}
	usersByEmail[strings.ToLower(u.Email)] = u
	return nil
}

func createSession(userID string) (string, error) {
	token := newUUID()
	if redisClient != nil {
		ctx := context.Background()
		if err := redisClient.Set(ctx, sessionKey(token), userID, 7*24*time.Hour).Err(); err != nil {
			return "", err
		}
	}
	sessions[token] = userID
	return token, nil
}

func getUserIDBySession(token string) (string, bool) {
	if token == "" { return "", false }
	if redisClient != nil {
		ctx := context.Background()
		if s, err := redisClient.Get(ctx, sessionKey(token)).Result(); err == nil && s != "" {
			return s, true
		}
	}
	uid, ok := sessions[token]
	return uid, ok
}

func deleteSession(token string) {
	delete(sessions, token)
	if redisClient != nil {
		ctx := context.Background()
		redisClient.Del(ctx, sessionKey(token))
	}
}

// Extract nickname from email (everything before @)
func getNicknameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return email
}

// Determine user role based on email
func getUserRole(email string) string {
	if strings.ToLower(email) == "vladslugin987@gmail.com" {
		return "admin"
	}
	return "user"
}

func newUUID() string { return uuidNew() }

// separated to allow reuse without importing google/uuid here
func uuidNew() string { return generateUUID() }

// Helper functions for session cookie management
func setSessionCookie(c *gin.Context, token string) {
	c.SetCookie("session_token", token, 7*24*3600, "/", "", false, true)
}

func clearSessionCookie(c *gin.Context) {
	c.SetCookie("session_token", "", -1, "/", "", false, true)
}

// Provide a small UUID helper via google/uuid from handlers.go file

// Auth middleware to attach user to context if session cookie exists
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session_token")
		if err == nil && cookie != "" {
			if userID, ok := getUserIDBySession(cookie); ok {
				c.Set("userId", userID)
			}
		}
		c.Next()
	}
}

func SetupAuthRoutes(router *gin.Engine) {
	a := router.Group("/api/auth")
	{
		a.POST("/register", handleRegister)
		a.POST("/login", handleLogin)
		a.POST("/logout", handleLogout)
		a.GET("/me", handleMe)
	}
	
	// Admin routes
	admin := router.Group("/api/admin")
	admin.Use(requireAdminAuth())
	{
		admin.GET("/users", handleListUsers)
		admin.GET("/users/stats", handleUserStats)
	}
}

// Middleware to require admin authentication
func requireAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, exists := c.Get("userId")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
			return
		}
		
		userID := uid.(string)
		// Find user and check role
		for _, u := range usersByEmail {
			if u.ID == userID && u.Role == "admin" {
				c.Next()
				return
			}
		}
		
		// Check Redis if not found in memory
		if redisClient != nil {
			ctx := context.Background()
			keys := redisClient.Keys(ctx, "user:*").Val()
			for _, key := range keys {
				if id, err := redisClient.HGet(ctx, key, "id").Result(); err == nil && id == userID {
					if role, err := redisClient.HGet(ctx, key, "role").Result(); err == nil && role == "admin" {
						c.Next()
						return
					}
				}
			}
		}
		
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
	}
}

// List all users (admin only)
func handleListUsers(c *gin.Context) {
	var users []gin.H
	
	// Get from memory first
	for _, u := range usersByEmail {
		users = append(users, gin.H{
			"id": u.ID,
			"email": u.Email,
			"nickname": u.Nickname,
			"role": u.Role,
			"createdAt": u.CreatedAt,
			"lastLogin": u.LastLogin,
		})
	}
	
	// Get from Redis if available
	if redisClient != nil {
		ctx := context.Background()
		keys := redisClient.Keys(ctx, "user:*").Val()
		existingEmails := make(map[string]bool)
		
		// Mark existing emails from memory
		for _, u := range usersByEmail {
			existingEmails[u.Email] = true
		}
		
		for _, key := range keys {
			email, _ := redisClient.HGet(ctx, key, "email").Result()
			if !existingEmails[email] { // Avoid duplicates
				id, _ := redisClient.HGet(ctx, key, "id").Result()
				nickname, _ := redisClient.HGet(ctx, key, "nickname").Result()
				role, _ := redisClient.HGet(ctx, key, "role").Result()
				createdAtStr, _ := redisClient.HGet(ctx, key, "createdAt").Result()
				lastLoginStr, _ := redisClient.HGet(ctx, key, "lastLogin").Result()
				
				var createdAt time.Time
				var lastLogin *time.Time
				if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
					createdAt = t
				}
				if lastLoginStr != "" {
					if t, err := time.Parse(time.RFC3339, lastLoginStr); err == nil {
						lastLogin = &t
					}
				}
				
				users = append(users, gin.H{
					"id": id,
					"email": email,
					"nickname": nickname,
					"role": role,
					"createdAt": createdAt,
					"lastLogin": lastLogin,
				})
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// Get user statistics (admin only)
func handleUserStats(c *gin.Context) {
	totalUsers := len(usersByEmail)
	adminUsers := 0
	activeUsers := 0 // Users who logged in within last 30 days
	
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	for _, u := range usersByEmail {
		if u.Role == "admin" {
			adminUsers++
		}
		if u.LastLogin != nil && u.LastLogin.After(thirtyDaysAgo) {
			activeUsers++
		}
	}
	
	// Also check Redis
	if redisClient != nil {
		ctx := context.Background()
		keys := redisClient.Keys(ctx, "user:*").Val()
		existingEmails := make(map[string]bool)
		
		for _, u := range usersByEmail {
			existingEmails[u.Email] = true
		}
		
		for _, key := range keys {
			email, _ := redisClient.HGet(ctx, key, "email").Result()
			if !existingEmails[email] { // Count only new users from Redis
				totalUsers++
				
				role, _ := redisClient.HGet(ctx, key, "role").Result()
				if role == "admin" {
					adminUsers++
				}
				
				lastLoginStr, _ := redisClient.HGet(ctx, key, "lastLogin").Result()
				if lastLoginStr != "" {
					if t, err := time.Parse(time.RFC3339, lastLoginStr); err == nil && t.After(thirtyDaysAgo) {
						activeUsers++
					}
				}
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"totalUsers": totalUsers,
		"adminUsers": adminUsers,
		"activeUsers": activeUsers,
		"registeredToday": 0, // TODO: implement if needed
	})
}

func handleRegister(c *gin.Context) {
	var req struct{ Email, Password string }
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if _, exists := getUserByEmail(req.Email); exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}
	h, err := hashPassword(req.Password)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register"}); return }
	
	emailLower := strings.ToLower(req.Email)
	u := &User{ 
		ID: newUUID(), 
		Email: emailLower, 
		Nickname: getNicknameFromEmail(emailLower),
		Password: h, 
		Role: getUserRole(emailLower),
		CreatedAt: time.Now(),
	}
	if err := saveUser(u); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"}); return }
	// Auto-login
	token, err := createSession(u.ID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"}); return }
	setSessionCookie(c, token)
	c.JSON(http.StatusOK, gin.H{"id": u.ID, "email": u.Email, "nickname": u.Nickname, "role": u.Role})
}

func handleLogin(c *gin.Context) {
	var req struct{ Email, Password string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	u, ok := getUserByEmail(req.Email)
	if !ok || !checkPassword(u.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	
	// Update last login time
	now := time.Now()
	u.LastLogin = &now
	saveUser(u)
	
	token, err := createSession(u.ID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"}); return }
	setSessionCookie(c, token)
	c.JSON(http.StatusOK, gin.H{"id": u.ID, "email": u.Email, "nickname": u.Nickname, "role": u.Role})
}

func handleLogout(c *gin.Context) {
	if cookie, err := c.Cookie("session_token"); err == nil {
		deleteSession(cookie)
	}
	clearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func handleMe(c *gin.Context) {
	uid, _ := c.Get("userId")
	if uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
		return
	}
	
	userID := uid.(string)
	// Find user by ID
	for _, u := range usersByEmail {
		if u.ID == userID {
			c.JSON(http.StatusOK, gin.H{
				"id": u.ID, 
				"email": u.Email, 
				"nickname": u.Nickname, 
				"role": u.Role,
				"createdAt": u.CreatedAt,
				"lastLogin": u.LastLogin,
			})
			return
		}
	}
	
	// Check Redis if not found in memory
	if redisClient != nil {
		ctx := context.Background()
		keys := redisClient.Keys(ctx, "user:*").Val()
		for _, key := range keys {
			if id, err := redisClient.HGet(ctx, key, "id").Result(); err == nil && id == userID {
				email, _ := redisClient.HGet(ctx, key, "email").Result()
				nickname, _ := redisClient.HGet(ctx, key, "nickname").Result()
				role, _ := redisClient.HGet(ctx, key, "role").Result()
				createdAtStr, _ := redisClient.HGet(ctx, key, "createdAt").Result()
				lastLoginStr, _ := redisClient.HGet(ctx, key, "lastLogin").Result()
				
				var createdAt time.Time
				var lastLogin *time.Time
				if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
					createdAt = t
				}
				if lastLoginStr != "" {
					if t, err := time.Parse(time.RFC3339, lastLoginStr); err == nil {
						lastLogin = &t
					}
				}
				
				c.JSON(http.StatusOK, gin.H{
					"id": userID, 
					"email": email, 
					"nickname": nickname, 
					"role": role,
					"createdAt": createdAt,
					"lastLogin": lastLogin,
				})
				return
			}
		}
	}
	
	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
} 