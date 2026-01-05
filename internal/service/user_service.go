package service

import (
	"errors"

	"starter-kit-grpc-golang/internal/models"
	"starter-kit-grpc-golang/internal/repository"
	"starter-kit-grpc-golang/pkg/utils"
)

type UserService interface {
	CreateUser(name, email, password, role string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	GetUsers(filters map[string]interface{}, page, limit int32, sort string) ([]models.User, int64, error)
	UpdateUser(id string, req UpdateUserDTO) (*models.User, error)
	DeleteUser(id string) error
}

type userService struct {
	repo repository.UserRepository
}

type UpdateUserDTO struct {
	Name     string
	Email    string
	Password string
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(name, email, password, role string) (*models.User, error) {
	if exists, _ := s.repo.ExistsByEmail(email); exists {
		return nil, errors.New("email already taken")
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: password, // Hashed by GORM hook
		Role:     role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserByID(id string) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *userService) GetUsers(filters map[string]interface{}, page, limit int32, sort string) ([]models.User, int64, error) {
	paginationScope := &utils.PaginationScope{
		Page:  page,
		Limit: limit,
		Sort:  sort,
	}

	return s.repo.FindAll(filters, paginationScope)
}

func (s *userService) UpdateUser(id string, req UpdateUserDTO) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if req.Email != "" && req.Email != user.Email {
		if exists, _ := s.repo.ExistsByEmail(req.Email); exists {
			return nil, errors.New("email already taken")
		}
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Password != "" {
		user.Password = req.Password // Will be hashed by GORM hook
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) DeleteUser(id string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("user not found")
	}
	return s.repo.Delete(id)
}