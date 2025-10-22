package central

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// NewRouter wires the auth handlers into a standard http.ServeMux.
func NewRouter(service *AuthService) http.Handler {
	h := &handler{service: service}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/signup", h.signup)
	mux.HandleFunc("/auth/signin", h.signin)
	mux.HandleFunc("/auth/forgot-password", h.forgotPassword)
	mux.HandleFunc("/auth/reset-password", h.resetPassword)
	mux.HandleFunc("/auth/desktop/start", h.desktopStart)
	mux.HandleFunc("/auth/desktop/exchange", h.desktopExchange)

	return mux
}

type handler struct {
	service *AuthService
}

type emailPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *handler) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req emailPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	defer r.Body.Close()

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.service.Signup(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyInUse):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "unable to complete signup")
		}
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *handler) signin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req emailPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	defer r.Body.Close()

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.service.Signin(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "unable to complete signin")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
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

func (h *handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	defer r.Body.Close()

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	token, expiresAt, err := h.service.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "unable to process password reset")
		}
		return
	}

	writeJSON(w, http.StatusOK, forgotPasswordResponse{
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

func (h *handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	defer r.Body.Close()

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "token and password are required")
		return
	}

	if err := h.service.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, ErrInvalidResetToken):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "unable to reset password")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *handler) desktopStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		writeError(w, http.StatusBadRequest, "missing state param")
		return
	}

	// userID, err := h.service.GetUserIDFromContext(r.Context())
	// if err != nil || userID == "" {
	// 	writeError(w, http.StatusUnauthorized, "unauthorized")
	// 	return
	// }

	userID := "b1d42325-9d7c-4fbf-8b9c-31b82e9ac649"

	code, expiresAt, err := h.service.CreateDesktopLoginCode(r.Context(), userID, state)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create desktop login code")
		return
	}

	log.Printf("Generated code: %s", code)

	redirectURL := fmt.Sprintf("seisami://auth/callback?code=%s&expires=%d", code, expiresAt.Unix())
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

type desktopExchangeRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func (h *handler) desktopExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req desktopExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	defer r.Body.Close()

	if req.Code == "" || req.State == "" {
		writeError(w, http.StatusBadRequest, "missing code or state")
		return
	}

	token, err := h.service.ExchangeDesktopLoginCode(r.Context(), req.Code, req.State)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidOrExpiredCode):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "unable to complete exchange")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
