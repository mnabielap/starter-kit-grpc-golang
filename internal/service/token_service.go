package service

import (
	"time"

	"starter-kit-grpc-golang/config"
	"starter-kit-grpc-golang/internal/models"
	"starter-kit-grpc-golang/internal/repository"
	"starter-kit-grpc-golang/pkg/utils"
)

type TokenService struct {
	repo repository.TokenRepository
	cfg  *config.Config
}

func NewTokenService(repo repository.TokenRepository, cfg *config.Config) *TokenService {
	return &TokenService{repo: repo, cfg: cfg}
}

// GenerateAuthTokens creates Access and Refresh tokens
func (s *TokenService) GenerateAuthTokens(user *models.User) (string, string, time.Time, time.Time, error) {
	// 1. Generate Access Token
	accessToken, accessExp, err := utils.GenerateToken(
		user.ID,
		user.Role,
		"access",
		s.cfg.JWT.AccessExpiration,
		s.cfg.JWT.Secret,
	)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	// 2. Generate Refresh Token
	refreshToken, refreshExp, err := utils.GenerateToken(
		user.ID,
		user.Role,
		"refresh",
		s.cfg.JWT.RefreshExpiration,
		s.cfg.JWT.Secret,
	)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	// 3. Save Refresh Token to DB
	err = s.SaveToken(refreshToken, user.ID, refreshExp, models.TokenTypeRefresh)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	return accessToken, refreshToken, accessExp, refreshExp, nil
}

func (s *TokenService) SaveToken(token, userID string, expires time.Time, tokenType string) error {
	tokenModel := &models.Token{
		Token:   token,
		UserID:  userID,
		Expires: expires,
		Type:    tokenType,
	}
	return s.repo.Create(tokenModel)
}

func (s *TokenService) VerifyToken(token string, tokenType string) (*models.Token, error) {
	return s.repo.FindByToken(token, tokenType)
}