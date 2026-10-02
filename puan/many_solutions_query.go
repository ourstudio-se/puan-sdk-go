package puan

import (
	"time"
)

type ManySolutionsQuery struct {
	selectionGroups []Selections
	ruleset         Ruleset
	from            *time.Time
	to              *time.Time
}

func NewManySolutionsQuery(
	selectionGroups []Selections,
	ruleset Ruleset,
	from *time.Time,
	to *time.Time,
) ManySolutionsQuery {
	return ManySolutionsQuery{
		selectionGroups: selectionGroups,
		ruleset:         ruleset,
		from:            from,
		to:              to,
	}
}

func (query ManySolutionsQuery) asSolverQuery() (*MultiWeightSolverQuery, error) {
	preparedRuleset, err := query.prepareRuleset()
	if err != nil {
		return nil, err
	}

	weightGroups, err := calculateWeightGroups(
		preparedRuleset,
		query.selectionGroups,
	)
	if err != nil {
		return nil, err
	}

	solverQuery := NewMultiWeightSolverQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weightGroups,
	)

	return solverQuery, nil
}

func (q ManySolutionsQuery) prepareRuleset() (Ruleset, error) {
	var allSelections Selections
	for _, selectionGroup := range q.selectionGroups {
		allSelections = append(allSelections, selectionGroup...)
	}
	unorderedUniqueSelections := allSelections.dedupeUnordered()

	preparedRuleset, err := q.ruleset.modifyForQuery(
		unorderedUniqueSelections,
		q.from,
		q.to,
	)
	if err != nil {
		return Ruleset{}, err
	}

	return preparedRuleset, nil
}
