package database

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase() *gorm.DB {
	logrus.Info("Connecting to database")

	db, err := gorm.Open(postgres.Open(viper.GetString("postgres.dsn")))
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Connected to database")

	return db
}
