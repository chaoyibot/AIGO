package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
)

type Service struct {
	userRepo  *repository.UserRepo
	jwtSecret []byte
}

type Claims struct {
	UserID    string `json:"user_id"`
	PublicKey string `json:"public_key,omitempty"`
	jwt.RegisteredClaims
}

func NewService(userRepo *repository.UserRepo, jwtSecret string) *Service {
	return &Service{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) Register(publicKey string) (*model.User, error) {
	user := &model.User{
		ID:        uuid.New().String(),
		PublicKey: publicKey,
		Role:      "user",
		Status:    "active",
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	return user, nil
}

func (s *Service) GenerateAPIKey(userID, name string) (*model.APIKey, string, error) {
	rawKey := generateRandomKey()
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey := &model.APIKey{
		ID:      uuid.New().String(),
		UserID:  userID,
		KeyHash: keyHash,
		Name:    name,
	}
	if err := s.userRepo.CreateAPIKey(apiKey); err != nil {
		return nil, "", fmt.Errorf("generate api key: %w", err)
	}

	apiKey.KeyHash = "" // Don't expose hash
	return apiKey, rawKey, nil
}

func (s *Service) ExchangeToken(rawKey string) (string, *Claims, error) {
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey, err := s.userRepo.FindAPIKeyByHash(keyHash)
	if err != nil {
		return "", nil, fmt.Errorf("invalid api key: %w", err)
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return "", nil, fmt.Errorf("api key expired")
	}

	s.userRepo.UpdateAPIKeyLastUsed(apiKey.ID)

	user, err := s.userRepo.FindByID(apiKey.UserID)
	if err != nil {
		return "", nil, fmt.Errorf("user not found: %w", err)
	}

	claims := &Claims{
		UserID:    user.ID,
		PublicKey: user.PublicKey,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "aigo",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, fmt.Errorf("sign token: %w", err)
	}

	return tokenStr, claims, nil
}

func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func generateRandomKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
