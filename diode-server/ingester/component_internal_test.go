package ingester

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestParseMemory(t *testing.T) {
	tests := []struct {
		name     string
		info     string
		wantUsed int64
		wantMax  int64
		wantErr  bool
	}{
		{
			name: "real INFO memory section with maxmemory set",
			info: strings.Join([]string{
				"# Memory",
				"used_memory:104857600",
				"used_memory_human:100.00M",
				"used_memory_rss:120000000",
				"maxmemory:1073741824",
			}, "\r\n"),
			wantUsed: 104857600,
			wantMax:  1073741824,
		},
		{
			name:     "maxmemory missing defaults to 0",
			info:     "used_memory:42\n",
			wantUsed: 42,
			wantMax:  0,
		},
		{
			name:     "maxmemory:0 reported explicitly",
			info:     "used_memory:1024\nmaxmemory:0\n",
			wantUsed: 1024,
			wantMax:  0,
		},
		{
			name:     "leading whitespace tolerated",
			info:     "  used_memory: 42  \n  maxmemory: 100  \n",
			wantUsed: 42,
			wantMax:  100,
		},
		{
			name:    "missing used_memory errors",
			info:    "# Memory\nused_memory_rss:120000000\nmaxmemory:1024\n",
			wantErr: true,
		},
		{
			name:    "non-integer used_memory errors",
			info:    "used_memory:notanumber\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			used, max, err := parseMemory(tt.info)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got used=%d max=%d", used, max)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if used != tt.wantUsed {
				t.Fatalf("used: got %d, want %d", used, tt.wantUsed)
			}
			if max != tt.wantMax {
				t.Fatalf("max: got %d, want %d", max, tt.wantMax)
			}
		})
	}
}

// TestCheckRedisMemoryWatermark covers the decision branches without hitting
// Redis. Pre-seeding memCheckedAt skips the INFO call and lets us drive the
// cached used/max values directly.
func TestCheckRedisMemoryWatermark(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name         string
		pct          int
		used         int64
		max          int64
		expectReject bool
	}{
		{name: "disabled when pct is 0", pct: 0, used: 1 << 40, max: 1 << 30, expectReject: false},
		{name: "admit when maxmemory is 0 (unlimited)", pct: 80, used: 1 << 30, max: 0, expectReject: false},
		{name: "below threshold admits", pct: 80, used: 79, max: 100, expectReject: false},
		{name: "at threshold rejects", pct: 80, used: 80, max: 100, expectReject: true},
		{name: "above threshold rejects", pct: 80, used: 99, max: 100, expectReject: true},
		{name: "100 percent only rejects when full", pct: 100, used: 99, max: 100, expectReject: false},
		{name: "100 percent rejects at full", pct: 100, used: 100, max: 100, expectReject: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Component{
				config:       Config{RedisMemoryHighWatermarkPct: tt.pct},
				logger:       logger,
				memUsedBytes: tt.used,
				memMaxBytes:  tt.max,
				memCheckedAt: time.Now(), // skip the INFO call
			}
			err := c.checkRedisMemoryWatermark(context.Background())
			if tt.expectReject {
				if err == nil {
					t.Fatalf("expected rejection, got nil")
				}
				if got := status.Code(err); got != codes.ResourceExhausted {
					t.Fatalf("expected ResourceExhausted, got %s", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected admit, got error: %v", err)
			}
		})
	}
}
