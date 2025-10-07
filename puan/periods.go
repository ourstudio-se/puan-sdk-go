package puan

import (
	"sort"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type Period struct {
	from time.Time
	to   time.Time
}

func NewPeriod(from, to time.Time) (Period, error) {
	if !to.After(from) {
		return Period{},
			errors.Errorf(
				"from time %v must be before to time %v",
				from,
				to,
			)
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
func (p Period) overlaps(other Period) bool {
	return p.from.Before(other.to) && p.to.After(other.from)
}

// Checks if period contains another, including edges
func (p Period) contains(other Period) bool {
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

func (p timeBoundVariables) passed(timestamp time.Time) timeBoundVariables {
	return utils.Filter(p, func(periodVariable timeBoundVariable) bool {
		return periodVariable.period.to.Before(timestamp)
	})
}

// find all periods without caps or overlaps, sorted by start time
// Input:
// |----------------------|
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

	sortedEdges := getSortedPeriodEdges(periods)
	completePeriods := toPeriods(sortedEdges)

	return completePeriods
}

func getSortedPeriodEdges(periods []Period) []time.Time {
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

	return sortedEdges
}

func toPeriods(edges []time.Time) []Period {
	if len(edges) < 2 {
		return nil
	}

	periods := make([]Period, len(edges)-1)
	for i := range len(edges) - 1 {
		periods[i] = Period{
			from: edges[i],
			to:   edges[i+1],
		}
	}
	return periods
}

// '|' separated list of variable ids
// Need to have the variables serialized since it is used as a key in a map
type idsString string

func newIdsString(variables []string) (idsString, error) {
	if len(variables) == 0 {
		return "", errors.New("at least one variable is required")
	}

	sort.Strings(variables)
	value := strings.Join(variables, "|")
	return idsString(value), nil
}

func (p idsString) ids() []string {
	return strings.Split(string(p), "|")
}

func groupByPeriods(
	periodVariables timeBoundVariables,
	assumedVariables timeBoundVariables,
) (map[idsString][]string, error) {
	grouped := make(map[idsString][]string)

	for _, assumed := range assumedVariables {
		key, err := findContainingPeriodIDs(periodVariables, assumed.period)
		if err != nil {
			return nil, err
		}

		grouped[key] = append(grouped[key], assumed.variable)
	}

	return grouped, nil
}

func findContainingPeriodIDs(
	periodVariables timeBoundVariables,
	period Period,
) (idsString, error) {
	var overlapingPeriodIDs []string

	for _, periodVariable := range periodVariables {
		if periodVariable.period.overlaps(period) {
			overlapingPeriodIDs = append(overlapingPeriodIDs, periodVariable.variable)
		}
	}

	return newIdsString(overlapingPeriodIDs)
}
