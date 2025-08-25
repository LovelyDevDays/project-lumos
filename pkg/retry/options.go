package retry

import "time"

type retryOptions struct {
	maxRetries int
	backoff    time.Duration
	maxBackoff time.Duration
	retryable  func(error) bool
}

var defaultOptions = retryOptions{
	maxRetries: 3,
	backoff:    1 * time.Second,
	maxBackoff: 30 * time.Second,
	retryable:  func(err error) bool { return true },
}

type Option func(*retryOptions)

func WithMaxRetries(maxRetries int) Option {
	return func(opts *retryOptions) {
		opts.maxRetries = maxRetries
	}
}

func WithBackoff(backoff time.Duration) Option {
	return func(opts *retryOptions) {
		opts.backoff = backoff
	}
}

func WithMaxBackoff(maxBackoff time.Duration) Option {
	return func(opts *retryOptions) {
		opts.maxBackoff = maxBackoff
	}
}

func WithRetryable(retryable func(error) bool) Option {
	return func(opts *retryOptions) {
		if retryable == nil {
			opts.retryable = func(err error) bool { return true }
		} else {
			opts.retryable = retryable
		}
	}
}
