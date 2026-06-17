package mdgovalidator

import (
	"context"
	"time"
)

// ContextWrapper is a function that wraps a context with additional behavior.
type ContextWrapper func(context.Context) (context.Context, context.CancelFunc)

// ContextConfig holds configuration for context behavior in validation flows.
// This enables proper context propagation, timeouts, and cancellation support
// throughout the validation pipeline.
//
// Note: file and block limits are NOT context concerns — they live on
// FileValidator (WithMaxFiles / WithMaxBlocks), which performs the actual work.
type ContextConfig struct {
	// Timeout is the maximum duration for validation operations.
	// If zero, no timeout is applied.
	Timeout time.Duration

	// Deadline is the absolute deadline for validation.
	// If zero, no deadline is set.
	Deadline time.Time

	// Parent is the parent context for propagation.
	// If nil, context.Background() is used as base.
	//nolint:containedctx // Context config intentionally wraps parent context for propagation
	Parent context.Context
}

// DefaultContextConfig returns a default context configuration with sensible defaults.
func DefaultContextConfig() ContextConfig {
	return ContextConfig{
		Timeout:  0, // No timeout by default
		Deadline: time.Time{},
		Parent:   nil,
	}
}

// WithTimeout returns a new ContextConfig with the specified timeout.
func (c ContextConfig) WithTimeout(timeout time.Duration) ContextConfig {
	c.Timeout = timeout

	return c
}

// WithDeadline returns a new ContextConfig with the specified deadline.
func (c ContextConfig) WithDeadline(deadline time.Time) ContextConfig {
	c.Deadline = deadline

	return c
}

// WithParent returns a new ContextConfig with the specified parent context.
func (c ContextConfig) WithParent(parent context.Context) ContextConfig {
	c.Parent = parent

	return c
}

// wrapContextWithCancel wraps a context with a derived context and chains
// the cancel functions so both are called when the returned cancel is invoked.
func wrapContextWithCancel(
	ctx context.Context,
	cancel context.CancelFunc,
	wrap ContextWrapper,
) (context.Context, context.CancelFunc) {
	derivedCtx, derivedCancel := wrap(ctx)
	originalCancel := cancel

	return derivedCtx, func() {
		derivedCancel()
		originalCancel()
	}
}

// buildContextWrapper creates a wrapper function for context.WithCancel chaining.
func buildContextWrapper[T any](
	value T,
	wrapFn func(context.Context, T) (context.Context, context.CancelFunc),
) ContextWrapper {
	return func(parent context.Context) (context.Context, context.CancelFunc) {
		return wrapFn(parent, value)
	}
}

// Build creates a context.Context from this configuration.
// It chains context.WithCancel, context.WithTimeout, and context.WithDeadline
// based on the configuration values.
func (c ContextConfig) Build() (context.Context, context.CancelFunc) {
	parent := c.getParent()

	// Start with a cancelable context
	ctx, cancel := context.WithCancel(parent)

	// Apply timeout if set
	if c.Timeout > 0 {
		ctx, cancel = wrapContextWithCancel(
			ctx,
			cancel,
			buildContextWrapper(c.Timeout, context.WithTimeout),
		)
	}

	// Apply deadline if set
	if !c.Deadline.IsZero() {
		ctx, cancel = wrapContextWithCancel(
			ctx,
			cancel,
			buildContextWrapper(c.Deadline, context.WithDeadline),
		)
	}

	return ctx, cancel
}

// Branch creates a new branch context for parallel operations.
// This is useful when validating multiple files concurrently,
// where each file should get its own context that can be cancelled independently.
func (c ContextConfig) Branch() (context.Context, context.CancelFunc) {
	parent := c.getParent()

	// Create a new cancelable context branching from the parent
	return context.WithCancel(parent)
}

// BranchWithTimeout creates a new branch context with a timeout for parallel operations.
func (c ContextConfig) BranchWithTimeout(
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	parent := c.getParent()

	if timeout > 0 {
		return context.WithTimeout(parent, timeout)
	}

	return context.WithCancel(parent)
}

// BranchWithDeadline creates a new branch context with a deadline for parallel operations.
func (c ContextConfig) BranchWithDeadline(
	deadline time.Time,
) (context.Context, context.CancelFunc) {
	parent := c.getParent()

	if !deadline.IsZero() {
		return context.WithDeadline(parent, deadline)
	}

	return context.WithCancel(parent)
}

// getParent returns the parent context, defaulting to context.Background if nil.
func (c ContextConfig) getParent() context.Context {
	if c.Parent != nil {
		return c.Parent
	}

	return context.Background()
}
