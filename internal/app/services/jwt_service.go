package services

import (
	"time"

	"github.com/agilistikmal/parkingo-core/internal/app/pkg"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type JWTService struct {
	secretKey string
}

func NewJWTService() *JWTService {
	return &JWTService{
		secretKey: viper.GetString("jwt.secret_key"),
	}
}

func (s *JWTService) GenerateToken(userID int, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = pkg.GetCurrentTime().Add(duration).Unix()

	return token.SignedString([]byte(s.secretKey))
}

func (s *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
}

func (s *JWTService) ExtractTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func (s *JWTService) GetUserIDFromToken(tokenString string) (int, error) {
	claims, err := s.ExtractTokenClaims(tokenString)
	if err != nil {
		return 0, err
	}

	userID := int(claims["user_id"].(float64))
	return userID, nil
}
