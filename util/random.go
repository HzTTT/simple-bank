package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghigklmnopqrstuvwsyz"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(len(alphabet))]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandOwner() string {
	return RandomString(6)
}

func RandMoney() int64 {
	return RandomInt(0, 1000)
}

func RandCurrency() string {
	currencys := []string{"EUR", "USD", "RMB"}
	return currencys[rand.Intn(len(currencys))]
}

// RandomEmail generates a random email
func RandomEmail() string {
    return fmt.Sprintf("%s@email.com", RandomString(6))
}