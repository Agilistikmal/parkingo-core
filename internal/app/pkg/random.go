package pkg

import (
	"math/rand"
	"time"
)

func RandomString(length int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}
