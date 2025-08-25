package retry

import (
	"context"
	"errors"
	"time"
)

func Do(ctx context.Context, fn func(ctx context.Context) error, opts ...Option) error {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	var (
		err     error
		backoff = min(options.backoff, options.maxBackoff)
	)

	err = fn(ctx)
	if err == nil {
		return nil
	} else if !options.retryable(err) {
		return err
	}

	timer := time.NewTimer(backoff)
	defer timer.Stop()

	for i := 0; i < options.maxRetries; i++ {
		timer.Reset(backoff)
		select {
		case <-ctx.Done():
			return errors.Join(ctx.Err(), err)
		case <-timer.C:
			backoff = min(backoff*2, options.maxBackoff)
		}

		err = fn(ctx)
		if err == nil {
			return nil
		} else if !options.retryable(err) {
			return err
		}
	}

	return err
}

func DoWithData[T any](ctx context.Context, fn func(ctx context.Context) (T, error), opts ...Option) (T, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	var (
		result  T
		err     error
		backoff = min(options.backoff, options.maxBackoff)
	)

	result, err = fn(ctx)
	if err == nil {
		return result, nil
	} else if !options.retryable(err) {
		return result, err
	}

	timer := time.NewTimer(backoff)
	defer timer.Stop()

	for i := 0; i < options.maxRetries; i++ {
		timer.Reset(backoff)
		select {
		case <-ctx.Done():
			return result, errors.Join(ctx.Err(), err)
		case <-timer.C:
			backoff = min(backoff*2, options.maxBackoff)
		}

		result, err = fn(ctx)
		if err == nil {
			return result, nil
		} else if !options.retryable(err) {
			return result, err
		}
	}

	return result, err
}
