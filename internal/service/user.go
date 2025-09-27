package service

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/koyif/gophermart/internal/config"
	"github.com/koyif/gophermart/internal/domain"
	"github.com/koyif/gophermart/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

type UserRepository interface {
	CreateUser(login, hashedPassword string) (int64, error)
	User(login string) (*domain.User, error)
}

type UserService struct {
	config *config.Config
	repo   UserRepository
}

func NewUserService(repo UserRepository, config *config.Config) *UserService {
	return &UserService{
		repo:   repo,
		config: config,
	}
}

func (s *UserService) Register(login, password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		logger.Log.Warn("error while hashing password")
		return "", fmt.Errorf("error while hashing password: %w", err)
	}

	userId, err := s.repo.CreateUser(login, string(hashedPassword))
	if err != nil {
		return "", err
	}

	return generateJWTToken(userId, s.config.PrivateKey)
}

func (s *UserService) Login(login, password string) (string, error) {
	user, err := s.repo.User(login)
	if err != nil {
		if errors.Is(err, domain.ErrIncorrectCredentials) {
			logger.Log.Warn("incorrect login", logger.String("login", login))
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		logger.Log.Warn("incorrect password", logger.String("login", login))
		return "", domain.ErrIncorrectCredentials
	}

	return generateJWTToken(user.ID, s.config.PrivateKey)
}

func generateJWTToken(userId int64, privateKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userId, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("error while signing token: %w", err)
	}

	return signedToken, nil
}
