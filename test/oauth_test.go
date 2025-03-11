package test

import (
	"testing"

	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/config"
	"github.com/sirupsen/logrus"
)

func TestOAuth(t *testing.T) {
	config.Load()

	jwtService := services.NewJWTService()
	authService := services.NewAuthService(jwtService)

	url := authService.GetGoogleAuthURL("agil.zip")
	logrus.Info(url)
}
