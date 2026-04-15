package service

import (
	"errors"
	"marketplace-platform/models"
	"marketplace-platform/repository"
	"marketplace-platform/utils"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	merchantRepo *repository.MerchantRepository
}

func NewAuthService(userRepo *repository.UserRepository, merchantRepo *repository.MerchantRepository) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		merchantRepo: merchantRepo,
	}
}

// RegisterUser registers a new customer user.
func (s *AuthService) RegisterUser(email, password, name string, latitude, longitude float64, address string) (*models.User, string, error) {
	existing, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", errors.New("email già registrata")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		Latitude:  latitude,
		Longitude: longitude,
		Address:   address,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := utils.GenerateToken(user.ID, user.Email, "user")
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) RegisterMerchant(email, password, businessName string, merchantType models.MerchantType, latitude, longitude float64, address, city, phone string) (*models.Merchant, string, error) {
	existing, err := s.merchantRepo.GetByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", errors.New("email già registrata")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	merchant := &models.Merchant{
		Email:        email,
		Password:     string(hashedPassword),
		BusinessName: businessName,
		Type:         merchantType,
		Status:       models.MerchantStatusActive,
		Latitude:     latitude,
		Longitude:    longitude,
		Address:      address,
		City:         city,
		Phone:        phone,
	}

	if err := s.merchantRepo.Create(merchant); err != nil {
		return nil, "", err
	}

	token, err := utils.GenerateToken(merchant.ID, merchant.Email, "merchant")
	if err != nil {
		return nil, "", err
	}

	return merchant, token, nil
}

func (s *AuthService) LoginUser(email, password string) (*models.User, string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("credenziali non valide")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("credenziali non valide")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, "user")
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) LoginMerchant(email, password string) (*models.Merchant, string, error) {
	merchant, err := s.merchantRepo.GetByEmail(email)
	if err != nil {
		return nil, "", err
	}
	if merchant == nil {
		return nil, "", errors.New("credenziali non valide")
	}

	if merchant.Status != models.MerchantStatusActive {
		return nil, "", errors.New("account non attivo")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(merchant.Password), []byte(password)); err != nil {
		return nil, "", errors.New("credenziali non valide")
	}

	token, err := utils.GenerateToken(merchant.ID, merchant.Email, "merchant")
	if err != nil {
		return nil, "", err
	}

	return merchant, token, nil
}
