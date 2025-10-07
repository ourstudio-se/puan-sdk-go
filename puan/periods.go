package puan

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type Period struct {
	from time.Time
	to   time.Time
}

func NewPeriod(from, to time.Time) (Period, error) {
	if from.After(to) {
		return Period{}, errors.New("from time must be before to time")
	}

	return Period{
		from: from.Truncate(time.Second),
		to:   to.Truncate(time.Second),
	}, nil
}

func (p Period) From() time.Time {
	return p.from
}

func (p Period) To() time.Time {
	return p.to
}

// Checks overlap, excluding edges
func (p Period) Overlaps(other Period) bool {
	return p.from.Before(other.to) && p.to.After(other.from)
}

// Checks if period contains another, including edges
func (p Period) Contains(other Period) bool {
	return !other.from.Before(p.from) && !other.to.After(p.to)
}

type timeBoundVariables []timeBoundVariable

type timeBoundVariable struct {
	variable string
	period   Period
}

func (p timeBoundVariables) periods() []Period {
	periods := make([]Period, len(p))
	for i, periodVariable := range p {
		periods[i] = periodVariable.period
	}
	return periods
}

func (p timeBoundVariables) ids() []string {
	ids := make([]string, len(p))
	for i, periodVariable := range p {
		ids[i] = periodVariable.variable
	}
	return ids
}

// find all periods without caps or overlaps
// Input:
// |---|...................
// .......|------|.........
// ...................|---|
// .........|------|.......
// Output:
// |---|--|-|----|-|--|---|
func calculateCompletePeriods(
	periods []Period,
) []Period {
	if len(periods) == 0 {
		return nil
	}

	edges := make(map[time.Time]bool)
	for _, period := range periods {
		edges[period.from] = true
		edges[period.to] = true
	}

	var sortedEdges []time.Time
	for t := range edges {
		sortedEdges = append(sortedEdges, t)
	}
	sort.Slice(sortedEdges, func(i, j int) bool {
		return sortedEdges[i].Before(sortedEdges[j])
	})

	var completePeriods []Period

	for i := range len(sortedEdges) - 1 {
		period := Period{
			from: sortedEdges[i],
			to:   sortedEdges[i+1],
		}
		completePeriods = append(completePeriods, period)
	}

	return completePeriods
}

// '|' separated list of variable names
// Need to have the variables serialized since it is used as a key in a map
type periodVariables string

func newPeriodVariables(variables []string) periodVariables {
	sort.Strings(variables)
	value := strings.Join(variables, "|")
	return periodVariables(value)
}

func (p periodVariables) variables() []string {
	return strings.Split(string(p), "|")
}

func groupByPeriods(
	periods timeBoundVariables,
	assumedVariables timeBoundVariables,
) map[periodVariables][]string {
	grouped := make(map[periodVariables][]string)

	// For each assumed variable, find which period variables it overlaps with
	for _, assumedVar := range assumedVariables {
		var overlappingPeriodVars []string

		for _, periodVar := range periods {
			if assumedVar.period.Overlaps(periodVar.period) {
				overlappingPeriodVars = append(overlappingPeriodVars, periodVar.variable)
			}
		}

		key := newPeriodVariables(overlappingPeriodVars)

		grouped[key] = append(grouped[key], assumedVar.variable)
	}

	return grouped
}
