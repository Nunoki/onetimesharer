package randomizer

import (
	"math/rand"
	"time"
)

// String returns a random string of the defined length
func String(length uint8) string {
	rand.Seed(time.Now().UnixMilli())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, length)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
