package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateID creates a unique ID for entities
func GenerateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1000000))
}
