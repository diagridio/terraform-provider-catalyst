package helpers

import (
	"context"
	"time"
)

const DefaultTimeout = 5 * time.Minute

type WaitOptions struct {
	Timeout time.Duration
}

type WaitOption func(*WaitOptions)

// WithTimeout sets a custom timeout for the wait operation.
func WithTimeout(timeout time.Duration) WaitOption {
	return func(opts *WaitOptions) {
		opts.Timeout = timeout
	}
}

// WaitUntil waits for the given function to return true or an error.
func WaitUntil(ctx context.Context, fn func(context.Context) (bool, error), opts ...WaitOption) error {
	options := &WaitOptions{
		Timeout: DefaultTimeout,
	}

	for _, opt := range opts {
		opt(options)
	}

	// Add a timeout to the context only if it doesn't already have a deadline
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// immediately check the condition
	done, err := fn(ctx)
	if err != nil {
		return err
	}
	if done {
		return nil
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := fn(ctx)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}
