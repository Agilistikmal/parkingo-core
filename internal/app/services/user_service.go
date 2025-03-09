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

func (s *UserService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := s.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
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

func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	err := s.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		FullName: req.FullName,
		Role:     "USER",
		GoogleID: req.GoogleID,
	}

	err = s.DB.Create(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) UpdateUser(id int, req *models.UpdateUserRequest) (*models.User, error) {
	err := s.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	user, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	user.FullName = req.FullName
	user.Username = req.Username

	err = s.DB.Save(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(id int) error {
	err := s.DB.Where("id = ?", id).Delete(&models.User{}).Error
	if err != nil {
		return err
	}

	return nil
}
