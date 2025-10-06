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

// | separated list of variable names
type periodVariables string

func (p periodVariables) variables() []string {
	return strings.Split(string(p), "|")
}

// periodsOverlap checks if two periods overlap (excluding touching at edges)
func periodsOverlap(a, b period) bool {
	return a.from.Before(b.to) && b.from.Before(a.to)
}

func groupByPeriod(
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

		// Sort to ensure consistent key
		sort.Strings(overlappingPeriodVars)

		// Create the key
		key := periodVariables(strings.Join(overlappingPeriodVars, "|"))

		// Add the assumed variable to the group
		grouped[key] = append(grouped[key], assumedVar.variable)
	}

	return grouped
}
