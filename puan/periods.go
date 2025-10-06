package puan

import (
	"sort"
	"time"
)

var MinTime = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
var MaxTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

func calculateNonOverlappingPeriods(periods []period) []period {
	if len(periods) == 0 {
		return nil
	}

	periodEdges := make(map[time.Time]bool)
	periodEdges[MinTime] = true
	periodEdges[MaxTime] = true
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
