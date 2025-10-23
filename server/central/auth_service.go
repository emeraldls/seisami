package central

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"seisami/server/centraldb"
)

var (
	ErrEmailAlreadyInUse    = errors.New("email already registered")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidResetToken    = errors.New("invalid or expired reset token")
	ErrInvalidOrExpiredCode = errors.New("invalid or expired code")
)

type AuthService struct {
	queries              *centraldb.Queries
	jwtSecret            []byte
	jwtExpiration        time.Duration
	resetTokenExpiration time.Duration
}

type AuthResult struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func NewAuthService(queries *centraldb.Queries, cfg Config) *AuthService {
	return &AuthService{
		queries:              queries,
		jwtSecret:            []byte(cfg.JWTSecret),
		jwtExpiration:        cfg.JWTExpiration,
		resetTokenExpiration: cfg.ResetTokenExpiration,
	}
}

func (s *AuthService) Signup(ctx context.Context, email, password string) (AuthResult, error) {
	if _, err := s.queries.GetUserByEmail(ctx, email); err == nil {
		return AuthResult{}, ErrEmailAlreadyInUse
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return AuthResult{}, fmt.Errorf("lookup user: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	id := uuid.New()
	user, err := s.queries.CreateUser(ctx, centraldb.CreateUserParams{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		Email:        email,
		PasswordHash: string(hashed),
	})
	if err != nil {
		return AuthResult{}, fmt.Errorf("create user: %w", err)
	}

	userID, err := uuidFromPgtype(user.ID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("read user id: %w", err)
	}

	token, err := s.generateToken(userID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResult{Token: token, UserID: userID, Email: user.Email}, nil
}

func (s *AuthService) Signin(ctx context.Context, email, password string) (AuthResult, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, fmt.Errorf("lookup user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	userID, err := uuidFromPgtype(user.ID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("read user id: %w", err)
	}

	token, err := s.generateToken(userID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return AuthResult{Token: token, UserID: userID, Email: user.Email}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, time.Time, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, ErrUserNotFound
		}
		return "", time.Time{}, fmt.Errorf("lookup user: %w", err)
	}

	token := uuid.NewString()
	expiresAt := time.Now().Add(s.resetTokenExpiration)

	if err := s.queries.SetPasswordResetToken(ctx, centraldb.SetPasswordResetTokenParams{
		ID: user.ID,
		ResetToken: pgtype.Text{
			String: token,
			Valid:  true,
		},
		ResetTokenExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
	}); err != nil {
		return "", time.Time{}, fmt.Errorf("set reset token: %w", err)
	}

	return token, expiresAt, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.queries.GetUserByResetToken(ctx, pgtype.Text{String: token, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidResetToken
		}
		return fmt.Errorf("lookup user by reset token: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.queries.UpdatePassword(ctx, centraldb.UpdatePasswordParams{
		ID:           user.ID,
		PasswordHash: string(hashed),
	}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// uuidFromPgtype converts a pgtype.UUID to its string representation.
func uuidFromPgtype(id pgtype.UUID) (string, error) {
	if !id.Valid {
		return "", fmt.Errorf("uuid is not valid")
	}

	parsed, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return "", err
	}

	return parsed.String(), nil
}

func (s *AuthService) CreateDesktopLoginCode(ctx context.Context, userID, state string) (string, time.Time, error) {
	code, err := generateSecureCode(32)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(1 * time.Minute)

	id, err := uuid.Parse(userID)
	if err != nil {
		return "", time.Time{}, err
	}

	_, err = s.queries.CreateDesktopLoginCode(ctx, centraldb.CreateDesktopLoginCodeParams{
		Code:      code,
		UserID:    pgtype.UUID{Bytes: id, Valid: true},
		State:     state,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})

	if err != nil {
		return "", time.Time{}, err
	}

	return code, expiresAt, nil
}

func generateSecureCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AuthService) ExchangeDesktopLoginCode(ctx context.Context, code, state string) (string, error) {
	dbCode, err := s.queries.ConsumeDesktopLoginCode(ctx, centraldb.ConsumeDesktopLoginCodeParams{
		Code:  code,
		State: state,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidOrExpiredCode
		}
		return "", err
	}

	user, err := s.queries.GetUserByID(ctx, dbCode.UserID)
	if err != nil {
		return "", err
	}

	userId, err := uuidFromPgtype(user.ID)
	if err != nil {
		return "", err
	}

	token, err := s.generateToken(userId)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) GetUserIDFromContext(ctx context.Context) (string, error) {
	user, ok := ctx.Value(UserContextKey).(*centraldb.User)
	if !ok || user == nil {
		return "", fmt.Errorf("no user in context")
	}

	userId, err := uuidFromPgtype(user.ID)
	if err != nil {
		return "", err
	}
	return userId, nil
}
