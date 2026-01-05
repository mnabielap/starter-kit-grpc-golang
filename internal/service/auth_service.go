package service

import (
	"errors"
	"time"

	"starter-kit-grpc-golang/config"
	"starter-kit-grpc-golang/internal/models"
	"starter-kit-grpc-golang/internal/repository"
	"starter-kit-grpc-golang/pkg/utils"
)

type AuthService interface {
	Login(email, password string) (*models.User, string, string, time.Time, time.Time, error)
	Register(name, email, password string) (*models.User, string, string, time.Time, time.Time, error)
	RefreshAuth(refreshToken string) (string, string, time.Time, time.Time, error)
	Logout(refreshToken string) error
	
	ForgotPassword(email string) error
	ResetPassword(token, newPassword string) error
	SendVerificationEmail(userID string) error
	VerifyEmail(token string) error
}

type authService struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.TokenRepository
	tokenService *TokenService
	emailService EmailService
	cfg          *config.Config
}

func NewAuthService(uRepo repository.UserRepository, tRepo repository.TokenRepository, tService *TokenService, eService EmailService, cfg *config.Config) AuthService {
	return &authService{
		userRepo:     uRepo,
		tokenRepo:    tRepo,
		tokenService: tService,
		emailService: eService,
		cfg:          cfg,
	}
}

func (s *authService) Register(name, email, password string) (*models.User, string, string, time.Time, time.Time, error) {
	if exists, _ := s.userRepo.ExistsByEmail(email); exists {
		return nil, "", "", time.Time{}, time.Time{}, errors.New("email already taken")
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", time.Time{}, time.Time{}, err
	}

	accessToken, refreshToken, accessExp, refreshExp, err := s.tokenService.GenerateAuthTokens(user)
	return user, accessToken, refreshToken, accessExp, refreshExp, err
}

func (s *authService) Login(email, password string) (*models.User, string, string, time.Time, time.Time, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || !user.ComparePassword(password) {
		return nil, "", "", time.Time{}, time.Time{}, errors.New("incorrect email or password")
	}

	accessToken, refreshToken, accessExp, refreshExp, err := s.tokenService.GenerateAuthTokens(user)
	return user, accessToken, refreshToken, accessExp, refreshExp, err
}

func (s *authService) Logout(refreshToken string) error {
	tokenDoc, err := s.tokenService.VerifyToken(refreshToken, models.TokenTypeRefresh)
	if err != nil {
		return errors.New("token not found")
	}
	return s.tokenRepo.Delete(tokenDoc)
}

func (s *authService) RefreshAuth(refreshTokenStr string) (string, string, time.Time, time.Time, error) {
	// 1. Verify existence in DB
	tokenDoc, err := s.tokenService.VerifyToken(refreshTokenStr, models.TokenTypeRefresh)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, errors.New("please authenticate")
	}

	// 2. Validate JWT Signature
	payload, err := utils.ValidateToken(refreshTokenStr, s.cfg.JWT.Secret)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, errors.New("invalid token")
	}

	// 3. Get User
	user, err := s.userRepo.FindByID(payload.UserID)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, errors.New("user not found")
	}

	// 4. Delete old token (Rotation)
	s.tokenRepo.Delete(tokenDoc)

	// 5. Generate new pair
	return s.tokenService.GenerateAuthTokens(user)
}

func (s *authService) ForgotPassword(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil // Return success to prevent email enumeration
	}

	expires := s.cfg.JWT.ResetPasswordExpiration
	resetToken, _, err := utils.GenerateToken(user.ID, user.Role, models.TokenTypeResetPassword, expires, s.cfg.JWT.Secret)
	if err != nil {
		return err
	}

	err = s.tokenService.SaveToken(resetToken, user.ID, time.Now().Add(expires), models.TokenTypeResetPassword)
	if err != nil {
		return err
	}

	return s.emailService.SendResetPasswordEmail(user.Email, resetToken)
}

func (s *authService) ResetPassword(tokenStr, newPassword string) error {
	tokenDoc, err := s.tokenService.VerifyToken(tokenStr, models.TokenTypeResetPassword)
	if err != nil {
		return errors.New("password reset failed")
	}

	user, err := s.userRepo.FindByID(tokenDoc.UserID)
	if err != nil {
		return errors.New("invalid user data")
	}

	user.Password = newPassword
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Consume all reset tokens for this user
	return s.tokenRepo.DeleteByUserIDAndType(user.ID, models.TokenTypeResetPassword)
}

func (s *authService) SendVerificationEmail(userID string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	expires := s.cfg.JWT.VerifyEmailExpiration
	verifyToken, _, err := utils.GenerateToken(user.ID, user.Role, models.TokenTypeVerifyEmail, expires, s.cfg.JWT.Secret)
	if err != nil {
		return err
	}

	err = s.tokenService.SaveToken(verifyToken, user.ID, time.Now().Add(expires), models.TokenTypeVerifyEmail)
	if err != nil {
		return err
	}

	return s.emailService.SendVerificationEmail(user.Email, verifyToken)
}

func (s *authService) VerifyEmail(tokenStr string) error {
	tokenDoc, err := s.tokenService.VerifyToken(tokenStr, models.TokenTypeVerifyEmail)
	if err != nil {
		return errors.New("email verification failed")
	}

	user, err := s.userRepo.FindByID(tokenDoc.UserID)
	if err != nil {
		return errors.New("email verification failed")
	}

	user.IsEmailVerified = true
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	return s.tokenRepo.DeleteByUserIDAndType(user.ID, models.TokenTypeVerifyEmail)
}