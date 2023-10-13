package database

import (
	"testing"
	"time"
)

func Test_truncateToStart(t *testing.T) {
	tests := []struct {
		name string
		args time.Time
		want time.Time
	}{
		{
			name: "UTC",
			args: time.Date(2023, 1, 1, 17, 34, 59, 0, time.UTC),
			want: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "local",
			args: time.Date(1918, 10, 30, 12, 1, 9, 0, time.Local),
			want: time.Date(1918, 10, 30, 0, 0, 0, 0, time.Local),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateToStart(tt.args); got.Compare(tt.want) != 0 {
				t.Errorf("truncateToStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_truncateToEnd(t *testing.T) {
	tests := []struct {
		name string
		args time.Time
		want time.Time
	}{
		{
			name: "UTC",
			args: time.Date(2023, 1, 1, 17, 34, 59, 0, time.UTC),
			want: time.Date(2023, 1, 1, 23, 59, 59, 0, time.UTC),
		},
		{
			name: "local",
			args: time.Date(1918, 10, 30, 12, 1, 9, 0, time.Local),
			want: time.Date(1918, 10, 30, 23, 59, 59, 0, time.Local),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateToEnd(tt.args); got.Compare(tt.want) != 0 {
				t.Errorf("truncateToStart() = %v, want %v", got, tt.want)
			}
		})
	}
}
