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
	testutil.AssertZeroValue(t, "MaxFiles", cfg.MaxFiles, 0)
	testutil.AssertZeroValue(t, "MaxBlocksPerFile", cfg.MaxBlocksPerFile, 0)

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

	testutil.AssertContextNotNil(t, ctx)
	testutil.AssertContextCondition(t, ctx, false, "context should not be done immediately")
}

func TestContextConfigBuildWithParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	cfg := DefaultContextConfig().WithParent(parent)

	ctx, cancel := cfg.Build()
	defer cancel()

	testutil.AssertContextNotNil(t, ctx)

	parentCancel()

	testutil.AssertContextErr(
		t,
		ctx,
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
		t,
		ctx,
		context.DeadlineExceeded,
		"context should be done after timeout",
	)
}

func TestContextConfigBranch(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	ctx, cancel := cfg.Branch()
	defer cancel()

	testutil.AssertContextNotNil(t, ctx)
	testutil.AssertContextCondition(t, ctx, false, "branch should not be done immediately")
}

func TestContextConfigBranchWithTimeout(t *testing.T) {
	t.Parallel()

	cfg := DefaultContextConfig()

	ctx, cancel := cfg.BranchWithTimeout(10 * time.Millisecond)
	defer cancel()

	time.Sleep(20 * time.Millisecond)

	testutil.AssertContextErr(
		t,
		ctx,
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
		t,
		ctx,
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
		t,
		ctx,
		context.DeadlineExceeded,
		"context should be done after deadline",
	)
}
