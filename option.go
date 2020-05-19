/*
 * @Author: jinde.zgm
 * @Date: 2020-05-15 11:05:41
 * @Description: coroutine pool option.
 */

package gopool

import (
	"context"
	"errors"
	"math"
	"time"
)

const (
	// DefaultPoolSize is the default capacity for a default coroutine pool.
	DefaultPoolSize = math.MaxInt32

	// DefaultIdleTimeout is the interval time to clean up goroutines.
	DefaultIdleTimeout = time.Second
)

var (
	// ErrInvalidIdleTimeout will be returned when setting a negative number as the periodic duration to purge goroutines.
	ErrInvalidIdleTimeout = errors.New("invalid idle timeout for pool")

	// ErrInvalidPoolSize will be returned when setting a negative number as pool capacity.
	ErrInvalidPoolSize = errors.New("invalid size for pool")
)

// Option represents the optional function.
type Option func(opts *Options)

// Options contains all options which will be applied when instantiating a ants pool.
type Options struct {
	// Coroutine pool name.
	Name string

	// Coroutine pool capacity.
	Capacity int32

	// Coroutine pool base context.
	Context context.Context

	// IdleTimeout is the maximum amount of time an idle(keep-alive) coroutine will remain idle before closing itself.
	// Zero means no DefaultIdleTimeout.
	IdleTimeout time.Duration
}

// NewOptions create options by Option array。
func NewOptions(options ...Option) (*Options, error) {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}

	return opts, opts.Check()
}

// Check option validation.
func (opts *Options) Check() error {
	if opts.Capacity < 0 {
		return ErrInvalidPoolSize
	} else if 0 == opts.Capacity {
		opts.Capacity = DefaultPoolSize
	}

	if opts.IdleTimeout < 0 {
		return ErrInvalidIdleTimeout
	} else if opts.IdleTimeout == 0 {
		opts.IdleTimeout = DefaultIdleTimeout
	}

	return nil
}

// WithName create option to sets up coroutine name.
func WithName(name string) Option {
	return func(opts *Options) { opts.Name = name }
}

// WithCapacity create option to sets up coroutine pool capacity.
func WithCapacity(capacity int32) Option {
	return func(opts *Options) { opts.Capacity = capacity }
}

// WithContext create option to sets up coroutine base context.
func WithContext(ctx context.Context) Option {
	return func(opts *Options) { opts.Context = ctx }
}

// WithIdleTimeout create option to sets up the interval time of cleaning up goroutines.
func WithIdleTimeout(timeout time.Duration) Option {
	return func(opts *Options) { opts.IdleTimeout = timeout }
}
