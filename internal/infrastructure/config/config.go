package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Load() {
	logrus.Info("Loading config file")

	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Config file loaded")
}
