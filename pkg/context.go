package mdgovalidator

import (
	"context"
	"time"
)

// ContextConfig holds configuration for context behavior in validation flows.
// This enables proper context propagation, timeouts, and cancellation support
// throughout the validation pipeline.
type ContextConfig struct {
	// Timeout is the maximum duration for validation operations.
	// If zero, no timeout is applied.
	Timeout time.Duration

	// Deadline is the absolute deadline for validation.
	// If zero, no deadline is set.
	Deadline time.Time

	// MaxFiles is the maximum number of files to process.
	// If zero, all files are processed.
	MaxFiles int

	// MaxBlocksPerFile is the maximum number of code blocks to process per file.
	// If zero, all blocks are processed.
	MaxBlocksPerFile int

	// Parent is the parent context for propagation.
	// If nil, context.Background() is used as base.
	Parent context.Context
}

// DefaultContextConfig returns a default context configuration with sensible defaults.
func DefaultContextConfig() ContextConfig {
	return ContextConfig{
		Timeout:          0, // No timeout by default
		Deadline:         time.Time{},
		MaxFiles:         0, // Unlimited
		MaxBlocksPerFile: 0, // Unlimited
		Parent:           nil,
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

// WithMaxFiles returns a new ContextConfig with the specified max files.
func (c ContextConfig) WithMaxFiles(maxFiles int) ContextConfig {
	c.MaxFiles = maxFiles
	return c
}

// WithMaxBlocksPerFile returns a new ContextConfig with the specified max blocks.
func (c ContextConfig) WithMaxBlocksPerFile(maxBlocks int) ContextConfig {
	c.MaxBlocksPerFile = maxBlocks
	return c
}

// WithParent returns a new ContextConfig with the specified parent context.
func (c ContextConfig) WithParent(parent context.Context) ContextConfig {
	c.Parent = parent
	return c
}

// Build creates a context.Context from this configuration.
// It chains context.WithCancel, context.WithTimeout, and context.WithDeadline
// based on the configuration values.
func (c ContextConfig) Build() (context.Context, context.CancelFunc) {
	parent := c.Parent
	if parent == nil {
		parent = context.Background()
	}

	// Start with a cancelable context
	ctx, cancel := context.WithCancel(parent)

	// Apply timeout if set
	if c.Timeout > 0 {
		var timeoutCancel context.CancelFunc
		ctx, timeoutCancel = context.WithTimeout(ctx, c.Timeout)
		// Chain the cancel functions
		originalCancel := cancel
		cancel = func() {
			timeoutCancel()
			originalCancel()
		}
	}

	// Apply deadline if set
	if !c.Deadline.IsZero() {
		var deadlineCancel context.CancelFunc
		ctx, deadlineCancel = context.WithDeadline(ctx, c.Deadline)
		// Chain the cancel functions
		originalCancel := cancel
		cancel = func() {
			deadlineCancel()
			originalCancel()
		}
	}

	return ctx, cancel
}

// Branch creates a new branch context for parallel operations.
// This is useful when validating multiple files concurrently,
// where each file should get its own context that can be cancelled independently.
func (c ContextConfig) Branch() (context.Context, context.CancelFunc) {
	parent := c.Parent
	if parent == nil {
		parent = context.Background()
	}

	// Create a new cancelable context branching from the parent
	return context.WithCancel(parent)
}

// BranchWithTimeout creates a new branch context with a timeout for parallel operations.
func (c ContextConfig) BranchWithTimeout(
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	parent := c.Parent
	if parent == nil {
		parent = context.Background()
	}

	if timeout > 0 {
		return context.WithTimeout(parent, timeout)
	}

	return context.WithCancel(parent)
}

// BranchWithDeadline creates a new branch context with a deadline for parallel operations.
func (c ContextConfig) BranchWithDeadline(
	deadline time.Time,
) (context.Context, context.CancelFunc) {
	parent := c.Parent
	if parent == nil {
		parent = context.Background()
	}

	if !deadline.IsZero() {
		return context.WithDeadline(parent, deadline)
	}

	return context.WithCancel(parent)
}
