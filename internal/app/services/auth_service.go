package services

import (
	"context"
	"encoding/json"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService struct {
	JWTService   *JWTService
	GoogleConfig *oauth2.Config
}

type GoogleUserInfo struct {
	// define the fields of GoogleUserInfo here
}

func NewAuthService(jwtService *JWTService) *AuthService {
	return &AuthService{
		JWTService: jwtService,
		GoogleConfig: &oauth2.Config{
			ClientID:     viper.GetString("oauth.google.id"),
			ClientSecret: viper.GetString("oauth.google.secret"),
			RedirectURL:  viper.GetString("oauth.google.redirect_url"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (s *AuthService) GetGoogleAuthURL(state string) string {
	return s.GoogleConfig.AuthCodeURL(state)
}

func (s *AuthService) GetGoogleToken(code string) (*oauth2.Token, error) {
	return s.GoogleConfig.Exchange(context.Background(), code)
}

func (s *AuthService) GetGoogleUserInfo(token *oauth2.Token) (map[string]interface{}, error) {
	client := s.GoogleConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}
