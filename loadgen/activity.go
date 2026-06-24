package loadgen

import (
	"context"
	"math/rand"
	"time"
)

func SimpleActivity(ctx context.Context) error {
	randomSleep := time.Duration(rand.Intn(500)+500) * time.Millisecond
	time.Sleep(randomSleep)
	return nil
}
