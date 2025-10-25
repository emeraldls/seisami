package central

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"seisami/server/centraldb"
)

type ContextKey string

const UserContextKey ContextKey = "user"

var wsHandler func(http.ResponseWriter, *http.Request)

func SetWebSocketHandler(handler func(http.ResponseWriter, *http.Request)) {
	wsHandler = handler
}

func NewRouter(service *AuthService) *gin.Engine {
	router := gin.Default()

	corsMiddleware := cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization", "Upgrade", "Connection"},
		MaxAge:          12 * 60 * 60,
	})
	router.Use(corsMiddleware)

	h := &handler{service: service}

	router.GET("/ws", func(c *gin.Context) {

		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication token"})
			return
		}

		userID, err := validateToken(token, service.jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Request.Header.Set("X-User-ID", userID)
		if wsHandler != nil {
			wsHandler(c.Writer, c.Request)
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "WebSocket handler not available"})
		}
	})

	auth := router.Group("/auth")
	{
		auth.POST("/signup", h.signup)
		auth.POST("/signin", h.signin)
		auth.POST("/forgot-password", h.forgotPassword)
		auth.POST("/reset-password", h.resetPassword)
		auth.POST("/desktop/exchange", h.desktopExchange)

		protected := auth.Group("")
		protected.Use(authMiddleware(service))
		{
			protected.GET("/desktop/start", h.desktopStart)
		}
	}

	// Sync endpoints
	sync := router.Group("/sync")
	sync.Use(authMiddleware(service))
	{
		sync.POST("/upload", h.uploadData)
		sync.GET("/status", h.getSyncStatus)
	}

	return router
}

func authMiddleware(service *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return service.jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		userID := claims.Subject

		c.Request = c.Request.WithContext(setUserContext(c.Request.Context(), &centraldb.User{ID: stringToUUID(userID)}))
		c.Next()
	}
}

func setUserContext(ctx context.Context, user *centraldb.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func stringToUUID(userID string) pgtype.UUID {
	parsed, _ := uuid.Parse(userID)
	return pgtype.UUID{Bytes: parsed, Valid: true}
}

type handler struct {
	service *AuthService
}

type emailPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *handler) signup(c *gin.Context) {
	var req emailPasswordRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	result, err := h.service.Signup(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyInUse):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to complete signup"})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *handler) signin(c *gin.Context) {
	var req emailPasswordRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	result, err := h.service.Signin(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to complete signin"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type forgotPasswordResponse struct {
	Message      string    `json:"message"`
	ResetToken   string    `json:"reset_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	ExpiresInSec int64     `json:"expires_in_seconds"`
}

func (h *handler) forgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	token, expiresAt, err := h.service.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to process password reset"})
		}
		return
	}

	c.JSON(http.StatusOK, forgotPasswordResponse{
		Message:      "password reset token generated",
		ResetToken:   token,
		ExpiresAt:    expiresAt,
		ExpiresInSec: int64(time.Until(expiresAt).Seconds()),
	})
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"password"`
}

func (h *handler) resetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and password are required"})
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, ErrInvalidResetToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to reset password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func (h *handler) desktopStart(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing state param"})
		return
	}

	userID, err := h.service.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	code, expiresAt, err := h.service.CreateDesktopLoginCode(c.Request.Context(), userID, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create desktop login code"})
		return
	}

	log.Printf("Generated code: %s", code)

	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"expires": expiresAt.Unix(),
	})
}

type desktopExchangeRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func (h *handler) desktopExchange(c *gin.Context) {
	var req desktopExchangeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	if req.Code == "" || req.State == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or state"})
		return
	}

	token, err := h.service.ExchangeDesktopLoginCode(c.Request.Context(), req.Code, req.State)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidOrExpiredCode):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to complete exchange"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func validateToken(tokenString string, jwtSecret []byte) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.Subject, nil
}

// Data Sync Handlers
type syncUploadRequest struct {
	Boards         []map[string]interface{} `json:"boards"`
	Columns        []map[string]interface{} `json:"columns"`
	Cards          []map[string]interface{} `json:"cards"`
	Transcriptions []map[string]interface{} `json:"transcriptions"`
}

func (h *handler) uploadData(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req syncUploadRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	// Process upload in background and stream progress updates
	go h.processSyncUpload(c.Request.Context(), userID, req)

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "in_progress",
		"message": "sync upload started",
	})
}

func (h *handler) processSyncUpload(ctx context.Context, userID string, req syncUploadRequest) {
	// This will be called asynchronously
	// For now, we'll log the request and process it
	log.Printf("Starting data sync for user %s: %d boards, %d columns, %d cards, %d transcriptions",
		userID, len(req.Boards), len(req.Columns), len(req.Cards), len(req.Transcriptions))

	// TODO: Implement actual database inserts with conflict handling
	// Track duplicates, errors, and processed items
	// Emit WebSocket progress events to client
}

func (h *handler) getSyncStatus(c *gin.Context) {
	userID, err := h.service.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// TODO: Return sync status from cache/database
	c.JSON(http.StatusOK, gin.H{
		"status":  "idle",
		"message": "no active sync",
	})
}
