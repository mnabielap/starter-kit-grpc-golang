package repository

import (
	"starter-kit-grpc-golang/internal/models"

	"gorm.io/gorm"
)

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db}
}

func (r *tokenRepository) Create(token *models.Token) error {
	return r.db.Create(token).Error
}

func (r *tokenRepository) FindByToken(tokenStr string, tokenType string) (*models.Token, error) {
	var token models.Token
	err := r.db.Where("token = ? AND type = ? AND blacklisted = ?", tokenStr, tokenType, false).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepository) DeleteByUserIDAndType(userID string, tokenType string) error {
	return r.db.Where("user_id = ? AND type = ?", userID, tokenType).Delete(&models.Token{}).Error
}

func (r *tokenRepository) Delete(token *models.Token) error {
	return r.db.Delete(token).Error
}