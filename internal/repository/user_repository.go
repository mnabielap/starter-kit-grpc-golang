package repository

import (
	"strings"

	"starter-kit-grpc-golang/internal/models"
	"starter-kit-grpc-golang/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(filters map[string]interface{}, pagination *utils.PaginationScope) ([]models.User, int64, error) {
	var users []models.User
	var totalRows int64

	query := r.db.Model(&models.User{})

	// --- 1. SEARCH LOGIC (Replicated from REST API) ---
	if search, ok := filters["search"].(string); ok && search != "" {
		scope, _ := filters["scope"].(string)
		searchPattern := "%" + strings.ToLower(search) + "%"

		switch scope {
		case "name":
			query = query.Where("lower(name) LIKE ?", searchPattern)
		case "email":
			query = query.Where("lower(email) LIKE ?", searchPattern)
		case "id":
			// Strict ID search if valid UUID
			if _, err := uuid.Parse(search); err == nil {
				query = query.Where("id = ?", search)
			} else {
				query = query.Where("1 = 0") // Invalid UUID returns nothing
			}
		default: // "all" or empty
			// OR Logic: Name OR Email OR ID
			subQuery := r.db.Where("lower(name) LIKE ?", searchPattern).
				Or("lower(email) LIKE ?", searchPattern)

			if _, err := uuid.Parse(search); err == nil {
				subQuery = subQuery.Or("id = ?", search)
			}
			query = query.Where(subQuery)
		}
	}

	// --- 2. FILTER LOGIC ---
	if role, ok := filters["role"].(string); ok && role != "" {
		query = query.Where("role = ?", role)
	}

	// --- 3. COUNT TOTAL ---
	query.Count(&totalRows)

	// --- 4. SORTING & PAGINATION ---
	allowedSortFields := map[string]bool{
		"id":         true,
		"name":       true,
		"email":      true,
		"role":       true,
		"created_at": true,
	}

	err := query.
		Scopes(pagination.SortScope(allowedSortFields)).
		Scopes(pagination.Paginate()).
		Find(&users).Error

	return users, totalRows, err
}

func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}