package puan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var MinTime = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
var MaxTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

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
			actual := calculateNonOverlappingPeriods(tt.periods, MinTime, MaxTime)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_groupByPeriod(t *testing.T) {
	tests := []struct {
		name             string
		periodVariables  timeBoundVariables
		assumedVariables timeBoundVariables
		expectedGroups   map[string][]string
	}{
		{
			name: "assumed variable overlaps with multiple period variables",
			periodVariables: timeBoundVariables{
				{
					variable: "p1",
					period: period{
						from: newTestTime("2024-01-01T00:00:00Z"),
						to:   newTestTime("2024-01-10T00:00:00Z"),
					},
				},
				{
					variable: "p2",
					period: period{
						from: newTestTime("2024-01-10T00:00:00Z"),
						to:   newTestTime("2024-01-15T00:00:00Z"),
					},
				},
			},
			assumedVariables: timeBoundVariables{
				{
					variable: "v1",
					period: period{
						from: newTestTime("2024-01-05T00:00:00Z"),
						to:   newTestTime("2024-01-12T00:00:00Z"),
					},
				},
			},
			expectedGroups: map[string][]string{
				"p1|p2": {"v1"},
			},
		},
		{
			name: "multiple assumed variables with different overlaps",
			periodVariables: timeBoundVariables{
				{
					variable: "p1",
					period: period{
						from: newTestTime("2024-01-01T00:00:00Z"),
						to:   newTestTime("2024-01-10T00:00:00Z"),
					},
				},
				{
					variable: "p2",
					period: period{
						from: newTestTime("2024-01-10T00:00:00Z"),
						to:   newTestTime("2024-01-20T00:00:00Z"),
					},
				},
				{
					variable: "p3",
					period: period{
						from: newTestTime("2024-01-20T00:00:00Z"),
						to:   newTestTime("2024-01-30T00:00:00Z"),
					},
				},
			},
			assumedVariables: timeBoundVariables{
				{
					variable: "v1",
					period: period{
						from: newTestTime("2024-01-05T00:00:00Z"),
						to:   newTestTime("2024-01-08T00:00:00Z"),
					},
				},
				{
					variable: "v2",
					period: period{
						from: newTestTime("2024-01-15T00:00:00Z"),
						to:   newTestTime("2024-01-25T00:00:00Z"),
					},
				},
				{
					variable: "v3",
					period: period{
						from: newTestTime("2024-01-05T00:00:00Z"),
						to:   newTestTime("2024-01-08T00:00:00Z"),
					},
				},
			},
			expectedGroups: map[string][]string{
				"p1":    {"v1", "v3"},
				"p2|p3": {"v2"},
			},
		},
		{
			name: "variable ends at the start of the next period",
			periodVariables: timeBoundVariables{
				{
					variable: "p1",
					period: period{
						from: newTestTime("2024-01-01T00:00:00Z"),
						to:   newTestTime("2024-01-10T00:00:00Z"),
					},
				},
				{
					variable: "p2",
					period: period{
						from: newTestTime("2024-01-10T00:00:00Z"),
						to:   newTestTime("2024-01-15T00:00:00Z"),
					},
				},
			},
			assumedVariables: timeBoundVariables{
				{
					variable: "v1",
					period: period{
						from: newTestTime("2024-01-05T00:00:00Z"),
						to:   newTestTime("2024-01-10T00:00:00Z"),
					},
				},
			},
			expectedGroups: map[string][]string{
				"p1": {"v1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := groupByPeriod(tt.periodVariables, tt.assumedVariables)

			// Convert temp keys to strings for easier comparison
			actualGroups := make(map[string][]string)
			for key, vars := range actual {
				actualGroups[string(key)] = vars
			}

			assert.Equal(t, tt.expectedGroups, actualGroups)
		})
	}
}

func Test_periodsOverlap(t *testing.T) {
	tests := []struct {
		name     string
		a        period
		b        period
		expected bool
	}{
		{
			name: "touching at end edge, should not overlap",
			a: period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
			b: period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-15T00:00:00Z"),
			},
			expected: false,
		},
		{
			name: "touching at start edge, should not overlap",
			a: period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-15T00:00:00Z"),
			},
			b: period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := periodsOverlap(tt.a, tt.b)
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
