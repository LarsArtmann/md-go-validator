package mdgovalidator

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultContextConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	assertZeroValue(t, "Timeout", cfg.Timeout, time.Duration(0))
	assertZeroValue(t, "Deadline", cfg.Deadline, time.Time{})
	assertZeroValue(t, "MaxFiles", cfg.MaxFiles, 0)
	assertZeroValue(t, "MaxBlocksPerFile", cfg.MaxBlocksPerFile, 0)

	if cfg.Parent != nil {
		t.Error("expected nil parent")
	}
}

func TestContextConfigWithTimeout(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig().WithTimeout(5 * time.Second)

	if cfg.Timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", cfg.Timeout)
	}
}

func TestContextConfigWithDeadline(t *testing.T) {
	t.Parallel()

	deadline := time.Now().Add(10 * time.Second)
	cfg := DefaultContextConfig().WithDeadline(deadline)

	if !cfg.Deadline.Equal(deadline) {
		t.Errorf("expected deadline %v, got %v", deadline, cfg.Deadline)
	}
}

func TestContextConfigWithMaxFiles(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig().WithMaxFiles(100)

	if cfg.MaxFiles != 100 {
		t.Errorf("expected maxFiles 100, got %d", cfg.MaxFiles)
	}
}

func TestContextConfigWithMaxBlocksPerFile(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig().WithMaxBlocksPerFile(50)

	if cfg.MaxBlocksPerFile != 50 {
		t.Errorf("expected maxBlocksPerFile 50, got %d", cfg.MaxBlocksPerFile)
	}
}

func TestContextConfigWithMaxLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		setMax            func(ContextConfig) ContextConfig
		getMax            func(ContextConfig) int
		expected          int
		expectedFieldName string
	}{
		{
			name:              "WithMaxFiles",
			setMax:            func(c ContextConfig) ContextConfig { return c.WithMaxFiles(100) },
			getMax:            func(c ContextConfig) int { return c.MaxFiles },
			expected:          100,
			expectedFieldName: "maxFiles",
		},
		{
			name:              "WithMaxBlocksPerFile",
			setMax:            func(c ContextConfig) ContextConfig { return c.WithMaxBlocksPerFile(50) },
			getMax:            func(c ContextConfig) int { return c.MaxBlocksPerFile },
			expected:          50,
			expectedFieldName: "maxBlocksPerFile",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := tc.setMax(DefaultContextConfig())

			if got := tc.getMax(cfg); got != tc.expected {
				t.Errorf("expected %s %d, got %d", tc.expectedFieldName, tc.expected, got)
			}
		})
	}
}

func TestContextConfigWithParent(t *testing.T) {
	t.Parallel()

	parent := context.Background()
	cfg := DefaultContextConfig().WithParent(parent)

	if cfg.Parent != parent {
		t.Error("expected parent to be set")
	}
}

func TestContextConfigBuild(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig().WithTimeout(100 * time.Millisecond)

	ctx, cancel := cfg.Build()
	defer cancel()

	assertContextNotNil(t, ctx)
	assertContextNotDone(t, ctx, "context should not be done immediately")
}

func TestContextConfigBuildWithParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	cfg := DefaultContextConfig().WithParent(parent)

	ctx, cancel := cfg.Build()
	defer cancel()

	assertContextNotNil(t, ctx)

	// Cancelling parent should cancel the context chain
	parentCancel()

	select {
	case <-ctx.Done():
	default:
		t.Fatal("context should be done after parent is cancelled")
	}
}

func TestContextConfigBuildTimeout(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig().WithTimeout(10 * time.Millisecond)

	ctx, cancel := cfg.Build()
	defer cancel()

	time.Sleep(20 * time.Millisecond)

	assertContextDeadlineExceeded(t, ctx, "context should be done after timeout")
}

func TestContextConfigBranch(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	ctx, cancel := cfg.Branch()
	defer cancel()

	assertContextNotNil(t, ctx)
	assertContextNotDone(t, ctx, "branch should not be done immediately")
}

func TestContextConfigBranchWithTimeout(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	ctx, cancel := cfg.BranchWithTimeout(10 * time.Millisecond)
	defer cancel()

	time.Sleep(20 * time.Millisecond)

	assertContextDeadlineExceeded(t, ctx, "context should be done after timeout")
}

func TestContextConfigBranchWithDeadline(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()
	deadline := time.Now().Add(10 * time.Millisecond)

	ctx, cancel := cfg.BranchWithDeadline(deadline)
	defer cancel()

	time.Sleep(20 * time.Millisecond)

	assertContextDeadlineExceeded(t, ctx, "context should be done after deadline")
}

func TestContextConfigBuildChainedTimeoutAndDeadline(t *testing.T) {
	t.Parallel()

	deadline := time.Now().Add(50 * time.Millisecond)
	cfg := DefaultContextConfig().
		WithTimeout(100 * time.Millisecond).
		WithDeadline(deadline)

	ctx, cancel := cfg.Build()
	defer cancel()

	time.Sleep(70 * time.Millisecond)

	assertContextDeadlineExceeded(t, ctx, "context should be done after deadline")
}

//nolint:revive // Test helper function, t must be first for consistency with testing package
func assertContextNotNil(t *testing.T, ctx context.Context) {
	t.Helper()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
}

// assertContextCondition checks context state against expected condition.
func assertContextCondition(t *testing.T, ctx context.Context, expectDone bool, msg string) {
	t.Helper()

	done := isContextDone(ctx)
	if done == expectDone {
		return
	}
	t.Fatal(msg)
}

// assertContextNotDone checks that the context is NOT done.
// If the context is done, t.Fatal is called with msg.
func assertContextNotDone(t *testing.T, ctx context.Context, msg string) {
	assertContextCondition(t, ctx, false, msg)
}

// assertContextDone checks that the context IS done.
// If the context is not done, t.Fatal is called with msg.
func assertContextDone(t *testing.T, ctx context.Context, msg string) {
	assertContextCondition(t, ctx, true, msg)
}

// assertContextWithError checks that the context is done with a specific error.
// If the context is not done, t.Fatal is called with msg.
// If the context is done with a different error, t.Errorf is called.
func assertContextWithError(t *testing.T, ctx context.Context, expected error, msg string) {
	t.Helper()

	assertContextDone(t, ctx, msg)

	if !errors.Is(ctx.Err(), expected) {
		t.Errorf("expected %v, got %v", expected, ctx.Err())
	}
}

// assertContextDeadlineExceeded is a convenience wrapper for assertContextWithError with DeadlineExceeded.
func assertContextDeadlineExceeded(t *testing.T, ctx context.Context, msg string) {
	assertContextWithError(t, ctx, context.DeadlineExceeded, msg)
}

// isContextDone checks if the context is done without blocking.
func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// assertZeroValue checks that got equals expected using == comparison.
func assertZeroValue[T comparable](t *testing.T, name string, got, expected T) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %s %v, got %v", name, expected, got)
	}
}
