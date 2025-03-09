package services

import (
	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type UserService struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewUserService(db *gorm.DB, validate *validator.Validate) *UserService {
	return &UserService{
		DB:       db,
		Validate: validate,
	}
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	var user *models.User
	err := s.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user *models.User
	err := s.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) CreateUser(user *models.User) error {
	err := s.Validate.Struct(user)
	if err != nil {
		return err
	}

	err = s.DB.Create(&user).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateUser(id int, user *models.User) error {
	err := s.Validate.Struct(user)
	if err != nil {
		return err
	}

	err = s.DB.Where("id = ?", id).Updates(&user).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(id int) error {
	err := s.DB.Where("id = ?", id).Delete(&models.User{}).Error
	if err != nil {
		return err
	}

	return nil
}
