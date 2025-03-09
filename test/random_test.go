package test

import (
	"testing"

	"github.com/agilistikmal/parkingo-core/internal/app/pkg"
	"github.com/sirupsen/logrus"
)

func TestRandom(t *testing.T) {
	result := pkg.RandomString(8)
	logrus.Info(result)
}
