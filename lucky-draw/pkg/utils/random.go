package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

func GenerateNumber() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

func GenerateDrawID() string {
	return fmt.Sprintf("draw:%d", time.Now().UnixNano())
}