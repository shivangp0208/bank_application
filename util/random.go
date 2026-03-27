package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var rGen *rand.Rand

var alphabets = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rGen = rand.New(rand.NewSource(time.Now().Unix()))
}

func GenerateRandomCurrency() string {
	currencies := []string{"INR", "USD", "EUR", "CAD"}
	return currencies[rGen.Intn(len(currencies))]
}

func GenerateRandomAmount() int64 {
	return int64(generateAmount(5000, 10000))
}

func generateAmount(min, max int) int {
	return min + rGen.Intn(max-min+1)
}

func GenerateRandomName(length int) string {
	return generateName(length)
}

func generateName(length int) string {
	var name strings.Builder
	for range length {
		randomIndex := rGen.Intn(26)
		name.WriteByte(alphabets[randomIndex])
	}
	return name.String()
}

func GenerateRandomID() uint64 {
	return uint64(rGen.Intn(100))
}

func GenerateRandomEmail() string {
	return fmt.Sprintf("%s@email.com", GenerateRandomName(6))
}
