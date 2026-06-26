package mdgovalidator

import (
	"context"
	"time"
)

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

// chainCancel returns a CancelFunc that calls both cancelFuncs.
func chainCancel(c1, c2 context.CancelFunc) context.CancelFunc {
	return func() {
		c1()
		c2()
	}
}

// Build creates a context.Context from this configuration.
// It chains context.WithCancel, context.WithTimeout, and context.WithDeadline
// based on the configuration values.
func (c ContextConfig) Build() (context.Context, context.CancelFunc) {
	parent := c.getParent()

	ctx, cancel := context.WithCancel(parent)

	if c.Timeout > 0 {
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, c.Timeout)
		ctx = timeoutCtx
		cancel = chainCancel(cancel, timeoutCancel)
	}

	if !c.Deadline.IsZero() {
		deadlineCtx, deadlineCancel := context.WithDeadline(ctx, c.Deadline)
		ctx = deadlineCtx
		cancel = chainCancel(cancel, deadlineCancel)
	}

	return ctx, cancel
}

// getParent returns the parent context, defaulting to context.Background if nil.
func (c ContextConfig) getParent() context.Context {
	if c.Parent != nil {
		return c.Parent
	}

	return context.Background()
}
