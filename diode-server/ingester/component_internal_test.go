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

func TestParseUsedMemory(t *testing.T) {
	tests := []struct {
		name    string
		info    string
		want    int64
		wantErr bool
	}{
		{
			name: "real INFO memory section",
			info: strings.Join([]string{
				"# Memory",
				"used_memory:104857600",
				"used_memory_human:100.00M",
				"used_memory_rss:120000000",
				"maxmemory:1073741824",
			}, "\r\n"),
			want: 104857600,
		},
		{
			name: "leading whitespace tolerated",
			info: "  used_memory: 42  \n",
			want: 42,
		},
		{
			name:    "missing line",
			info:    "# Memory\nused_memory_rss:120000000\n",
			wantErr: true,
		},
		{
			name:    "non-integer value",
			info:    "used_memory:notanumber\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUsedMemory(tt.info)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

// TestCheckRedisMemoryWatermark covers the decision branches without hitting
// Redis. Pre-seeding memCheckedAt skips the INFO call and lets us drive the
// cached used-bytes value directly.
func TestCheckRedisMemoryWatermark(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name         string
		threshold    int64
		usedBytes    int64
		expectReject bool
	}{
		{name: "disabled when threshold is 0", threshold: 0, usedBytes: 1 << 40, expectReject: false},
		{name: "below threshold admits", threshold: 1000, usedBytes: 999, expectReject: false},
		{name: "at threshold rejects", threshold: 1000, usedBytes: 1000, expectReject: true},
		{name: "above threshold rejects", threshold: 1000, usedBytes: 2000, expectReject: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Component{
				config:       Config{RedisMemoryHighWatermarkBytes: tt.threshold},
				logger:       logger,
				memUsedBytes: tt.usedBytes,
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
