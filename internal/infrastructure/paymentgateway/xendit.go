package paymentgateway

import (
	"github.com/spf13/viper"
	"github.com/xendit/xendit-go/v6"
)

type Xendit struct {
	Client *xendit.APIClient
}

func NewXendit() *Xendit {
	client := xendit.NewClient(viper.GetString("xendit.secret_key"))

	return &Xendit{
		Client: client,
	}
}
