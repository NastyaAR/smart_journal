package services

import (
	"blockchain_project/internal/models"
	"blockchain_project/internal/repositories"
	"context"
)

type AuthService struct {
	userRepository *repositories.UserRepository
}

func NewAuthService(userRepository *repositories.UserRepository) *AuthService {
	return &AuthService{userRepository: userRepository}
}

func (s *AuthService) Authenticate(ctx context.Context, login, password string) (*models.User, error) {
	return s.userRepository.Authenticate(ctx, login, password)
}
