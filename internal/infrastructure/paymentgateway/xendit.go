package paymentgateway

import (
	"github.com/spf13/viper"
	"github.com/xendit/xendit-go/v6"
)

func NewXendit() *xendit.APIClient {
	return xendit.NewClient(viper.GetString("xendit.secret_key"))
}
