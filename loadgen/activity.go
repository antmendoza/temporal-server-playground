package loadgen

import (
	"context"
	"time"
)

func SimpleActivity(ctx context.Context) error {
	time.Sleep(200 * time.Millisecond)
	return nil
}
