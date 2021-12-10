package client

import (
	"math/rand"
	"time"
)

var runeChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomString generate random string for messate sending
func GenerateRandomString(n int) string {
	b := make([]rune, n+1)
	for i := range b {
		b[i] = runeChars[rand.Intn(len(runeChars))]
	}
	return string(b)
}
