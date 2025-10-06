package puan

import (
	"sort"
	"strings"
	"time"
)

func calculateNonOverlappingPeriods(
	periods []period,
	startTime time.Time,
	endTime time.Time,
) []period {
	if len(periods) == 0 {
		return nil
	}

	periodEdges := make(map[time.Time]bool)
	periodEdges[startTime] = true
	periodEdges[endTime] = true
	for _, period := range periods {
		periodEdges[period.from] = true
		periodEdges[period.to] = true
	}

	var sortedPeriodEdges []time.Time
	for t := range periodEdges {
		sortedPeriodEdges = append(sortedPeriodEdges, t)
	}
	sort.Slice(sortedPeriodEdges, func(i, j int) bool {
		return sortedPeriodEdges[i].Before(sortedPeriodEdges[j])
	})

	var nonOverlappingPeriods []period

	for i := range len(sortedPeriodEdges) - 1 {
		period := period{
			from: sortedPeriodEdges[i],
			to:   sortedPeriodEdges[i+1],
		}
		nonOverlappingPeriods = append(nonOverlappingPeriods, period)
	}

	return nonOverlappingPeriods
}

// '|' separated list of variable names
type periodVariables string

func newPeriodVariables(variables []string) periodVariables {
	sort.Strings(variables)
	value := strings.Join(variables, "|")
	return periodVariables(value)
}

func (p periodVariables) variables() []string {
	return strings.Split(string(p), "|")
}

// periodsOverlap checks if two periods overlap (excluding touching at edges)
func periodsOverlap(a, b period) bool {
	return a.from.Before(b.to) && a.to.After(b.from)
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
			if periodsOverlap(assumedVar.period, periodVar.period) {
				overlappingPeriodVars = append(overlappingPeriodVars, periodVar.variable)
			}
		}

		key := newPeriodVariables(overlappingPeriodVars)

		grouped[key] = append(grouped[key], assumedVar.variable)
	}

	return grouped
}
