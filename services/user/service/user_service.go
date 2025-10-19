package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/terkoizmy/golearn/internal/util"
	"github.com/terkoizmy/golearn/services/user/domain"
	"github.com/terkoizmy/golearn/services/user/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type UserService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	ValidateToken(ctx context.Context, token string) (string, error)
}

type userService struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewUserService(repo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *userService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	// Hash the password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: hashedPassword,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

func (s *userService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// Get user by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if !util.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := util.GenerateJWT(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Clear password before returning
	user.Password = ""

	return &domain.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

func (s *userService) ValidateToken(ctx context.Context, token string) (string, error) {
	claims, err := util.ValidateJWT(token, s.jwtSecret)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}
