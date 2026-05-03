package cron

import (
	"testing"
	"time"
)

func TestNextRun(t *testing.T) {
	// Use a fixed reference time: Monday 2025-01-06 10:00:00 UTC
	ref := time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		expression string
		wantHour   int
		wantMinute int
	}{
		{
			name:       "every 5 minutes",
			expression: "*/5 * * * *",
			wantHour:   10,
			wantMinute: 5,
		},
		{
			name:       "every hour at 0",
			expression: "0 * * * *",
			wantHour:   11,
			wantMinute: 0,
		},
		{
			name:       "every day at midnight",
			expression: "0 0 * * *",
			wantHour:   0,
			wantMinute: 0,
		},
		{
			name:       "specific minute 30",
			expression: "30 * * * *",
			wantHour:   10,
			wantMinute: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextRun(tt.expression, ref)
			if err != nil {
				t.Fatalf("NextRun() error = %v", err)
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("NextRun() hour = %d, want %d", got.Hour(), tt.wantHour)
			}
			if got.Minute() != tt.wantMinute {
				t.Errorf("NextRun() minute = %d, want %d", got.Minute(), tt.wantMinute)
			}
		})
	}
}

func TestNextRunInvalid(t *testing.T) {
	_, err := NextRun("invalid", time.Now())
	if err == nil {
		t.Error("NextRun() expected error for invalid expression")
	}
}

func TestNextRunEvery5Minutes(t *testing.T) {
	// If current time is 10:03, next run should be 10:05
	ref := time.Date(2025, 1, 6, 10, 3, 0, 0, time.UTC)
	got, err := NextRun("*/5 * * * *", ref)
	if err != nil {
		t.Fatalf("NextRun() error = %v", err)
	}
	if got.Minute() != 5 {
		t.Errorf("NextRun() minute = %d, want 5", got.Minute())
	}
}
