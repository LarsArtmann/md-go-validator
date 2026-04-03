package mdgovalidator

import (
	"context"
	"testing"
	"time"
)

func TestDefaultContextConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	if cfg.Timeout != 0 {
		t.Errorf("expected timeout 0, got %v", cfg.Timeout)
	}
	if !cfg.Deadline.IsZero() {
		t.Errorf("expected zero deadline, got %v", cfg.Deadline)
	}
	if cfg.MaxFiles != 0 {
		t.Errorf("expected maxFiles 0, got %d", cfg.MaxFiles)
	}
	if cfg.MaxBlocksPerFile != 0 {
		t.Errorf("expected maxBlocksPerFile 0, got %d", cfg.MaxBlocksPerFile)
	}
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

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// Context should not be done immediately
	select {
	case <-ctx.Done():
		t.Fatal("context should not be done immediately")
	default:
	}
}

func TestContextConfigBuildWithParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	cfg := DefaultContextConfig().WithParent(parent)
	ctx, cancel := cfg.Build()
	defer cancel()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

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

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// Branch should not be done immediately
	select {
	case <-ctx.Done():
		t.Fatal("branch should not be done immediately")
	default:
	}
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

	time.Sleep(60 * time.Millisecond)

	assertContextDeadlineExceeded(t, ctx, "context should be done after deadline")
}

//nolint:revive // Test helper function, t must be first for consistency with testing package
func assertContextDeadlineExceeded(t *testing.T, ctx context.Context, msg string) {
	t.Helper()
	select {
	case <-ctx.Done():
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got %v", ctx.Err())
		}
	default:
		t.Fatal(msg)
	}
}
