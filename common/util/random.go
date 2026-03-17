package util

import (
	"fmt"
	"math/rand"
	"time"
)

// RandomBookingCode generates a random booking code
func RandomBookingCode() string {
	//nolint:gosec // non-cryptographic random is fine for booking codes
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("PIN-%d-%06d", time.Now().Year(), r.Intn(999999))
}
