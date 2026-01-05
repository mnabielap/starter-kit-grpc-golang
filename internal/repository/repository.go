package repository

import (
	"starter-kit-grpc-golang/internal/models"
	"starter-kit-grpc-golang/pkg/utils"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	FindAll(filters map[string]interface{}, pagination *utils.PaginationScope) ([]models.User, int64, error)
	ExistsByEmail(email string) (bool, error)
	Update(user *models.User) error
	Delete(id string) error
}

type TokenRepository interface {
	Create(token *models.Token) error
	FindByToken(token string, tokenType string) (*models.Token, error)
	DeleteByUserIDAndType(userID string, tokenType string) error
	Delete(token *models.Token) error
}