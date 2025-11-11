package central

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"seisami/server/centraldb"
	"seisami/server/types"
	"seisami/server/utils"
)

type ContextKey string

const UserContextKey ContextKey = "user"

var wsHandler func(http.ResponseWriter, *http.Request)

func SetWebSocketHandler(handler func(http.ResponseWriter, *http.Request)) {
	wsHandler = handler
}

func NewRouter(authService *AuthService, syncService *SyncService) *gin.Engine {
	router := gin.Default()

	corsMiddleware := cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization", "Upgrade", "Connection"},
		MaxAge:          12 * 60 * 60,
	})
	router.Use(corsMiddleware)

	h := &handler{authService: authService, syncService: syncService}

	router.GET("/ws", func(c *gin.Context) {

		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication token"})
			return
		}

		boardID := c.Query("board_id")
		if boardID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing board_id"})
			return
		}

		userID, err := validateToken(token, authService.jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		boardUUID, err := uuid.Parse(boardID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board_id"})
			return
		}

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID in token"})
			return
		}

		if err := h.syncService.ensureBoardAccess(c, boardUUID, userUUID); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Request.Header.Set("X-User-ID", userID)
		c.Request.Header.Set("X-Board-ID", boardID)
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
		protected.Use(authMiddleware(authService))
		{
			protected.GET("/desktop/start", h.desktopStart)
		}
	}

	// Sync endpoints
	sync := router.Group("/sync")
	sync.Use(authMiddleware(authService))
	{
		sync.POST("/init", h.initSyncState)
		sync.POST("/upload", h.uploadData)
		sync.GET("/pull/:table", h.pullData)
		sync.GET("/export", h.exportAllData)
		sync.GET("/export/:boardId", h.exportData)

		sync.POST("/init/cloud", h.initCloud)
		sync.GET("/cloud/status", h.getCloudStatus)

		sync.POST("/board", h.upsertBoard)
		sync.POST("/column", h.upsertColumn)
		sync.POST("/card", h.upsertCard)

		sync.POST("/state", h.updateSyncState)
		sync.GET("/state/:table", h.getSyncState)
	}

	boardRts := router.Group("/board")
	boardRts.Use(authMiddleware(authService))
	{
		boardRts.POST("/invite", h.inviteUser)
		boardRts.DELETE("/remove", h.removeUserFromBoard)
		boardRts.GET("/:boardId/members", h.getBoardMembers)
		boardRts.GET("/:boardId/metadata", h.getBoardMetadata)
	}

	updates := router.Group("/updates")
	{
		updates.GET("/latest", h.getLatestAppVersion)
		updates.POST("/publish", h.createNewAppVersion)
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

		c.Request = c.Request.WithContext(setUserContext(c.Request.Context(), &centraldb.User{ID: utils.StringToUUID(userID)}))
		c.Next()
	}
}

func setUserContext(ctx context.Context, user *centraldb.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

type handler struct {
	authService *AuthService
	syncService *SyncService
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

	result, err := h.authService.Signup(c.Request.Context(), req.Email, req.Password)
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

	result, err := h.authService.Signin(c.Request.Context(), req.Email, req.Password)
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

	token, expiresAt, err := h.authService.ForgotPassword(c.Request.Context(), req.Email)
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

	if err := h.authService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
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

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	code, expiresAt, err := h.authService.CreateDesktopLoginCode(c.Request.Context(), userID, state)
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

	token, err := h.authService.ExchangeDesktopLoginCode(c.Request.Context(), req.Code, req.State)
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

func (h *handler) exportData(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	boardID := c.Param("boardId")
	if boardID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "board id is required"})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	data, err := h.syncService.ExportBoardData(c.Request.Context(), uid, boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	boardData, _ := json.MarshalIndent(data, "", " ")
	fmt.Println(string(boardData))

	c.JSON(http.StatusCreated, gin.H{"message": "import successful", "data": data})
}

func (h *handler) exportAllData(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	data, err := h.syncService.ExportAllData(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	boardData, _ := json.MarshalIndent(data, "", " ")
	fmt.Println(string(boardData))

	c.JSON(http.StatusCreated, gin.H{"message": "import successful", "data": data})
}

func (h *handler) uploadData(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync service unavailable"})
		return
	}

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SyncOperation
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields", "data": err.Error()})
		return
	}

	if err := h.syncService.ProcessOperation(c.Request.Context(), userID, req); err != nil {
		log.Printf("sync upload failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process sync operation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":       "stored",
		"record_id":    req.RecordID,
		"operation_id": req.ID,
	})
}

func (h *handler) initCloud(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	err = h.syncService.initCloud(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful"})
}

func (h *handler) getCloudStatus(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	status, err := h.syncService.getCloudStatus(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successful", "data": gin.H{"status": status}})
}

func (h *handler) getSyncState(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "table name is required"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	syncState, err := h.syncService.getSyncState(c, id, tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successful", "data": syncState})
}

func (h *handler) updateSyncState(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body types.SyncStatePayload

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	err = h.syncService.updateSyncState(c, id, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured updating sync state: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successful"})
}

// This will defintely run only once, when a user just setup cloud
func (h *handler) upsertBoard(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	var payload boardPayload

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	jBytes, _ := json.MarshalIndent(payload, "", " ")
	fmt.Println(string(jBytes))

	validate := validator.New()
	if err := validate.Struct(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields: " + err.Error()})
		return
	}

	err = h.syncService.upsertBoard(c, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful"})
}

func (h *handler) upsertColumn(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	var payload columnPayload

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields: " + err.Error()})
		return
	}

	err = h.syncService.upsertColumn(c, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful"})
}

func (h *handler) upsertCard(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	_, err = uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	var payload cardPayload

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields: " + err.Error()})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldnt parse user id to uuid: " + err.Error()})
		return
	}

	err = h.syncService.upsertCard(c, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful"})
}

func (h *handler) initSyncState(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse id: " + err.Error()})
		return
	}

	err = h.syncService.initializeSyncState(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful"})

}

func (h *handler) pullData(c *gin.Context) {

	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table name is required"})
		return
	}

	since := c.Query("since")

	var sinceTimestamp int64

	if since != "" {
		sinceTimestamp, err = strconv.ParseInt(since, 10, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid since parameter"})
			return
		}
	}

	operations, err := h.syncService.PullOperations(c.Request.Context(), userID, tableName, sinceTimestamp)
	if err != nil {
		log.Printf("sync pull failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to pull sync data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"table":      tableName,
		"count":      len(operations),
		"operations": operations,
	})
}

// Implement notification service for things like this

func (h *handler) inviteUser(c *gin.Context) {
	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var payload boardMemberActionPayload

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields: " + err.Error()})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldnt parse user id to uuid: " + err.Error()})
		return
	}

	err = h.syncService.inviteUserToBoard(c, id, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful"})
}

func (h *handler) removeUserFromBoard(c *gin.Context) {
	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var payload boardMemberActionPayload

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields: " + err.Error()})
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldnt parse user id to uuid: " + err.Error()})
		return
	}

	err = h.syncService.removeUserFromBoard(c, id, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful"})

}

func (h *handler) getBoardMembers(c *gin.Context) {
	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	boardID := c.Param("boardId")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "board id is required"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldnt parse user id to uuid: " + err.Error()})
		return
	}

	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldnt parse board id to uuid: " + err.Error()})
		return
	}

	boardMembers, err := h.syncService.getBoardMembers(c, boardUUID, userUUID)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful", "data": boardMembers})

}

func (h *handler) getBoardMetadata(c *gin.Context) {
	userID, err := h.authService.GetUserIDFromContext(c.Request.Context())
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	boardID := c.Param("boardId")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "board id is required"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id: " + err.Error()})
		return
	}

	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board id: " + err.Error()})
		return
	}

	metadata, err := h.syncService.getBoardMetadata(c, boardUUID, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successful", "data": metadata})
}

func (h *handler) getLatestAppVersion(c *gin.Context) {

	version, err := h.syncService.getAppLatestVersion(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successful", "data": version})
}

type createAppVersionPayload struct {
	ReleaseURL string `json:"release_url" validate:"required,url"`
	VersionKey string `json:"version_key" validate:"required"`
	Notes      string `json:"notes" validate:"required"`
}

func (h *handler) createNewAppVersion(c *gin.Context) {
	var payload createAppVersionPayload

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	validate := validator.New()
	if err := validate.Struct(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields: " + err.Error()})
		return
	}

	versionKey := h.authService.versionKey

	if payload.VersionKey != versionKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.syncService.createAppNewVersion(c, payload.ReleaseURL, payload.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "successful"})
}
