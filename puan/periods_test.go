package puan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewPeriod(t *testing.T) {
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	actual, err := NewPeriod(from, to)

	assert.NoError(t, err)
	assert.Equal(t, from, actual.From())
	assert.Equal(t, to, actual.To())
}

func Test_NewPeriod_givenFromAfterTo_shouldReturnError(t *testing.T) {
	from := newTestTime("2024-01-31T00:00:00Z")
	to := newTestTime("2024-01-01T00:00:00Z")

	_, err := NewPeriod(from, to)

	assert.Error(t, err)
}

func Test_NewPeriod_givenFromEqualToTo_shouldReturnError(t *testing.T) {
	from := newTestTime("2024-01-31T00:00:00Z")
	to := from

	_, err := NewPeriod(from, to)

	assert.Error(t, err)
}

func Test_Period_overlaps(t *testing.T) {
	tests := []struct {
		name     string
		a        Period
		b        Period
		expected bool
	}{
		{
			name: "touching at end edge, should not overlap",
			a: Period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
			b: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-15T00:00:00Z"),
			},
			expected: false,
		},
		{
			name: "touching at start edge, should not overlap",
			a: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-15T00:00:00Z"),
			},
			b: Period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
			expected: false,
		},
		{
			name: "overlapping",
			a: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-20T00:00:00Z"),
			},
			b: Period{
				from: newTestTime("2024-01-15T00:00:00Z"),
				to:   newTestTime("2024-01-25T00:00:00Z"),
			},
			expected: true,
		},
		{
			name: "not overlapping",
			a: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-20T00:00:00Z"),
			},
			b: Period{
				from: newTestTime("2024-01-25T00:00:00Z"),
				to:   newTestTime("2024-01-30T00:00:00Z"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.a.overlaps(tt.b)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_Period_contains(t *testing.T) {
	tests := []struct {
		name     string
		period   Period
		other    Period
		expected bool
	}{
		{
			name: "period contains another with exact same boundaries",
			period: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			expected: true,
		},
		{
			name: "period fully contains another smaller period",
			period: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-20T00:00:00Z"),
			},
			expected: true,
		},
		{
			name: "period contains another at start edge",
			period: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-15T00:00:00Z"),
			},
			expected: true,
		},
		{
			name: "period contains another at end edge",
			period: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-15T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			expected: true,
		},
		{
			name: "period does not contain another - extends before start",
			period: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-20T00:00:00Z"),
			},
			expected: false,
		},
		{
			name: "period does not contain another - extends after end",
			period: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-20T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			expected: false,
		},
		{
			name: "period does not contain another - completely before",
			period: Period{
				from: newTestTime("2024-01-15T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
			expected: false,
		},
		{
			name: "period does not contain another - completely after",
			period: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-15T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			expected: false,
		},
		{
			name: "period does not contain another - other is larger",
			period: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-20T00:00:00Z"),
			},
			other: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-31T00:00:00Z"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.period.contains(tt.other)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_calculateCompletePeriods(t *testing.T) {
	tests := []struct {
		name     string
		periods  []Period
		expected []Period
	}{
		{
			name: "given single period",
			periods: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-31T00:00:00Z"),
				},
			},
			expected: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-31T00:00:00Z"),
				},
			},
		},
		{
			name: "given no overlaps or gaps",
			periods: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-31T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-31T00:00:00Z"),
					to:   newTestTime("2024-02-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-02-15T00:00:00Z"),
					to:   newTestTime("2024-02-28T00:00:00Z"),
				},
			},
			expected: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-31T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-31T00:00:00Z"),
					to:   newTestTime("2024-02-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-02-15T00:00:00Z"),
					to:   newTestTime("2024-02-28T00:00:00Z"),
				},
			},
		},
		{
			name: "given overlap",
			periods: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-10T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
			},
			expected: []Period{
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
			},
		},
		{
			name: "given gap",
			periods: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-15T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-20T00:00:00Z"),
					to:   newTestTime("2024-01-30T00:00:00Z"),
				},
			},
			expected: []Period{
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
			},
		},
		{
			name: "given overlap and gap",
			periods: []Period{
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
			expected: []Period{
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
			actual := calculateCompletePeriods(tt.periods)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_getSortedPeriodEdges(t *testing.T) {
	tests := []struct {
		name     string
		periods  []Period
		expected []time.Time
	}{
		{
			name: "single period",
			periods: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-31T00:00:00Z"),
				},
			},
			expected: []time.Time{
				newTestTime("2024-01-01T00:00:00Z"),
				newTestTime("2024-01-31T00:00:00Z"),
			},
		},
		{
			name: "multiple periods with unique edges",
			periods: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-10T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-15T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
			},
			expected: []time.Time{
				newTestTime("2024-01-01T00:00:00Z"),
				newTestTime("2024-01-10T00:00:00Z"),
				newTestTime("2024-01-15T00:00:00Z"),
				newTestTime("2024-01-20T00:00:00Z"),
			},
		},
		{
			name: "multiple periods with shared edges",
			periods: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-10T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-10T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
			},
			expected: []time.Time{
				newTestTime("2024-01-01T00:00:00Z"),
				newTestTime("2024-01-10T00:00:00Z"),
				newTestTime("2024-01-20T00:00:00Z"),
			},
		},
		{
			name:     "empty input",
			periods:  []Period{},
			expected: nil,
		},
		{
			name: "multiple periods randomly ordered",
			periods: []Period{
				{
					from: newTestTime("2024-01-10T00:00:00Z"),
					to:   newTestTime("2024-01-20T00:00:00Z"),
				},
				{
					from: newTestTime("2024-01-10T00:00:00Z"),
					to:   newTestTime("2024-01-01T00:00:00Z"),
				},
				{
					from: newTestTime("2023-12-20T00:00:00Z"),
					to:   newTestTime("2023-12-10T00:00:00Z"),
				},
			},
			expected: []time.Time{
				newTestTime("2023-12-10T00:00:00Z"),
				newTestTime("2023-12-20T00:00:00Z"),
				newTestTime("2024-01-01T00:00:00Z"),
				newTestTime("2024-01-10T00:00:00Z"),
				newTestTime("2024-01-20T00:00:00Z"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getSortedPeriodEdges(tt.periods)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_toPeriods(t *testing.T) {
	tests := []struct {
		name     string
		edges    []time.Time
		expected []Period
	}{
		{
			name: "two edges create one period",
			edges: []time.Time{
				newTestTime("2024-01-01T00:00:00Z"),
				newTestTime("2024-01-31T00:00:00Z"),
			},
			expected: []Period{
				{
					from: newTestTime("2024-01-01T00:00:00Z"),
					to:   newTestTime("2024-01-31T00:00:00Z"),
				},
			},
		},
		{
			name: "multiple edges create consecutive periods",
			edges: []time.Time{
				newTestTime("2024-01-01T00:00:00Z"),
				newTestTime("2024-01-10T00:00:00Z"),
				newTestTime("2024-01-15T00:00:00Z"),
				newTestTime("2024-01-20T00:00:00Z"),
			},
			expected: []Period{
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
			},
		},
		{
			name: "single edge creates no periods",
			edges: []time.Time{
				newTestTime("2024-01-01T00:00:00Z"),
			},
			expected: nil,
		},
		{
			name:     "no edges create no periods",
			edges:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := toPeriods(tt.edges)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_groupByPeriods(t *testing.T) {
	periodVariables := TimeBoundVariables{
		{
			variable: "p1",
			period: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
		},
		{
			variable: "p2",
			period: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-20T00:00:00Z"),
			},
		},
		{
			variable: "p3",
			period: Period{
				from: newTestTime("2024-01-20T00:00:00Z"),
				to:   newTestTime("2024-01-30T00:00:00Z"),
			},
		},
	}
	assumedVariables := TimeBoundVariables{
		{
			variable: "x",
			period: Period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-08T00:00:00Z"),
			},
		},
		{
			variable: "y",
			period: Period{
				from: newTestTime("2024-01-15T00:00:00Z"),
				to:   newTestTime("2024-01-25T00:00:00Z"),
			},
		},
		{
			variable: "z",
			period: Period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-08T00:00:00Z"),
			},
		},
		{
			variable: "w",
			period: Period{
				from: newTestTime("2024-01-07T00:00:00Z"),
				to:   newTestTime("2024-01-17T00:00:00Z"),
			},
		},
	}

	actual, err := groupByPeriods(periodVariables, assumedVariables)

	assert.NoError(t, err)

	expected := map[idsString][]string{
		"p1":    {"x", "z"},
		"p2|p3": {"y"},
		"p1|p2": {"w"},
	}
	assert.Equal(t, expected, actual)
}

func Test_findContainingPeriodIDs(t *testing.T) {
	tests := []struct {
		name             string
		periodVariables  TimeBoundVariables
		comparisonPeriod Period
		expected         idsString
	}{
		{
			name: "assumed variable overlaps with multiple period variables",
			periodVariables: TimeBoundVariables{
				{
					variable: "p1",
					period: Period{
						from: newTestTime("2024-01-01T00:00:00Z"),
						to:   newTestTime("2024-01-10T00:00:00Z"),
					},
				},
				{
					variable: "p2",
					period: Period{
						from: newTestTime("2024-01-10T00:00:00Z"),
						to:   newTestTime("2024-01-15T00:00:00Z"),
					},
				},
			},
			comparisonPeriod: Period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-12T00:00:00Z"),
			},
			expected: "p1|p2",
		},
		{
			name: "variable ends at the start of a period",
			periodVariables: TimeBoundVariables{
				{
					variable: "p1",
					period: Period{
						from: newTestTime("2024-01-01T00:00:00Z"),
						to:   newTestTime("2024-01-10T00:00:00Z"),
					},
				},
				{
					variable: "p2",
					period: Period{
						from: newTestTime("2024-01-10T00:00:00Z"),
						to:   newTestTime("2024-01-15T00:00:00Z"),
					},
				},
			},
			comparisonPeriod: Period{
				from: newTestTime("2024-01-05T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
			expected: "p1",
		},
		{
			name: "variable starts at the end of a period",
			periodVariables: TimeBoundVariables{
				{
					variable: "p1",
					period: Period{
						from: newTestTime("2024-01-01T00:00:00Z"),
						to:   newTestTime("2024-01-10T00:00:00Z"),
					},
				},
				{
					variable: "p2",
					period: Period{
						from: newTestTime("2024-01-10T00:00:00Z"),
						to:   newTestTime("2024-01-15T00:00:00Z"),
					},
				},
			},
			comparisonPeriod: Period{
				from: newTestTime("2024-01-10T00:00:00Z"),
				to:   newTestTime("2024-01-20T00:00:00Z"),
			},
			expected: "p2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := findContainingPeriodIDs(tt.periodVariables, tt.comparisonPeriod)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_findContainingPeriodIDs_givenComparisonPeriodOutsideOfPeriods_shouldReturnError(
	t *testing.T,
) {
	periodVariables := TimeBoundVariables{
		{
			variable: "p1",
			period: Period{
				from: newTestTime("2024-01-01T00:00:00Z"),
				to:   newTestTime("2024-01-10T00:00:00Z"),
			},
		},
	}
	comparisonPeriod := Period{
		from: newTestTime("2024-01-15T00:00:00Z"),
		to:   newTestTime("2024-01-20T00:00:00Z"),
	}

	_, err := findContainingPeriodIDs(periodVariables, comparisonPeriod)

	assert.Error(t, err)
}

func Test_isEqual_givenEqual_shouldReturnTrue(t *testing.T) {
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	period := Period{
		from: from,
		to:   to,
	}

	other := Period{
		from: from,
		to:   to,
	}

	assert.True(t, period.isEqual(other))
}

func Test_isEqual_givenNotEqual_shouldReturnFalse(t *testing.T) {
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	period := Period{
		from: from,
		to:   to,
	}

	otherFrom := from.Add(time.Second)

	other := Period{
		from: otherFrom,
		to:   to,
	}

	assert.False(t, period.isEqual(other))
}

func newTestTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
