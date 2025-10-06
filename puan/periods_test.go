package puan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_calculate_non_overlapping_periods(t *testing.T) {
	tests := []struct {
		name     string
		periods  []period
		expected []period
	}{
		{
			name: "given no overlaps or gaps",
			periods: []period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-31T00:00:00Z"),
				},
			},
			expected: []period{
				{
					from: MinTime,
					to:   newTestTime("2024-01-01T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-31T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-31T00:00:00Z"),
					to:   MaxTime,
				},
			},
		},
		{
			name: "given overlap",
			periods: []period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-10T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
			},
			expected: []period{
				{
					from: MinTime,
					to:   newTestTime("2024-01-01T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-10T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-10T00:00:00Z"),
					to:   newTestTime("2024-01-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-15T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-20T00:00:00Z"),
					to:   MaxTime,
				},
			},
		},
		{
			name: "given gap",
			periods: []period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-20T00:00:00Z"),
					to:   newTestTime("2024-01-30T00:00:00Z"),
				},
			},
			expected: []period{
				{
					from: MinTime,
					to:   newTestTime("2024-01-01T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-15T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-20T00:00:00Z"),
					to:   newTestTime("2024-01-30T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-30T00:00:00Z"),
					to:   MaxTime,
				},
			},
		},
		{
			name: "given overlap and gap",
			periods: []period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-10T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-25T00:00:00Z"),
					to:   newTestTime("2024-01-30T00:00:00Z"),
				},
			},
			expected: []period{
				{
					from: MinTime,
					to:   newTestTime("2024-01-01T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-10T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-10T00:00:00Z"),
					to:   newTestTime("2024-01-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-15T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-20T00:00:00Z"),
					to:   newTestTime("2024-01-25T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-25T00:00:00Z"),
					to:   newTestTime("2024-01-30T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-30T00:00:00Z"),
					to:   MaxTime,
				},
			},
		},
		{
			name:     "given no periods",
			periods:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := calculateNonOverlappingPeriods(tt.periods)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func newTestTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
