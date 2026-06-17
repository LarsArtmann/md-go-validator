package mdgovalidator

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/md-go-validator/pkg/testutil"
)

func TestDefaultContextConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	testutil.AssertZeroValue(t, "Timeout", cfg.Timeout, time.Duration(0))
	testutil.AssertZeroValue(t, "Deadline", cfg.Deadline, time.Time{})

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

	testutil.AssertContextNotNil(ctx, t)
	testutil.AssertContextCondition(ctx, t, false, "context should not be done immediately")
}

func TestContextConfigBuildWithParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	cfg := DefaultContextConfig().WithParent(parent)

	ctx, cancel := cfg.Build()
	defer cancel()

	testutil.AssertContextNotNil(ctx, t)

	parentCancel()

	testutil.AssertContextErr(
		ctx, t,
		context.Canceled,
		"context should be done after parent is cancelled",
	)
}

func TestContextConfigBuildTimeout(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig().WithTimeout(10 * time.Millisecond)

	ctx, cancel := cfg.Build()
	defer cancel()

	time.Sleep(20 * time.Millisecond)

	testutil.AssertContextErr(
		ctx, t,
		context.DeadlineExceeded,
		"context should be done after timeout",
	)
}

func TestContextConfigBranch(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	ctx, cancel := cfg.Branch()
	defer cancel()

	testutil.AssertContextNotNil(ctx, t)
	testutil.AssertContextCondition(ctx, t, false, "branch should not be done immediately")
}

func TestContextConfigBranchWithTimeout(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	ctx, cancel := cfg.BranchWithTimeout(10 * time.Millisecond)
	defer cancel()

	time.Sleep(20 * time.Millisecond)

	testutil.AssertContextErr(
		ctx, t,
		context.DeadlineExceeded,
		"context should be done after timeout",
	)
}

func TestContextConfigBranchWithDeadline(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()
	deadline := time.Now().Add(10 * time.Millisecond)

	ctx, cancel := cfg.BranchWithDeadline(deadline)
	defer cancel()

	time.Sleep(20 * time.Millisecond)

	testutil.AssertContextErr(
		ctx, t,
		context.DeadlineExceeded,
		"context should be done after deadline",
	)
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

	testutil.AssertContextErr(
		ctx, t,
		context.DeadlineExceeded,
		"context should be done after deadline",
	)
}
