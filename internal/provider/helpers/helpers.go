package helpers

import (
	"context"
	"time"
)

func WaitUntil(ctx context.Context, fn func(context.Context) (bool, error)) error {
	// immediately check the condition
	done, err := fn(ctx)
	if err != nil {
		return err
	}
	if done {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		// check every 5s
		case <-time.After(2 * time.Second):
			done, err := fn(ctx)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}

	return nil
}
