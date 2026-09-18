package puan

import (
	"time"

	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type SolutionQuery struct {
	selections Selections
	ruleset    Ruleset
	from       *time.Time
	to         *time.Time
}

func NewSolutionQuery(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
	to *time.Time,
) (SolutionQuery, error) {
	if err := validateRuleset(ruleset); err != nil {
		return SolutionQuery{}, err
	}

	if err := validateTimestamps(from, to); err != nil {
		return SolutionQuery{}, err
	}

	if err := validateSelections(ruleset, selections); err != nil {
		return SolutionQuery{}, err
	}

	return SolutionQuery{
		selections: selections,
		ruleset:    ruleset,
		from:       from,
		to:         to,
	}, nil
}

type SolutionQueryBuilder struct {
	selections Selections
	ruleset    Ruleset
	from       *time.Time
	to         *time.Time
}

func NewSolutionQueryBuilder() *SolutionQueryBuilder {
	return &SolutionQueryBuilder{}
}

func (b *SolutionQueryBuilder) fromQuery(
	query SolutionQuery,
) *SolutionQueryBuilder {
	b.selections = query.selections
	b.ruleset = query.ruleset
	b.from = query.from
	b.to = query.to
	return b
}

func (b *SolutionQueryBuilder) WithSelections(selections Selections) *SolutionQueryBuilder {
	b.selections = selections
	return b
}

func (b *SolutionQueryBuilder) WithRuleset(ruleset Ruleset) *SolutionQueryBuilder {
	b.ruleset = ruleset
	return b
}

func (b *SolutionQueryBuilder) WithFrom(from *time.Time) *SolutionQueryBuilder {
	b.from = from
	return b
}

func (b *SolutionQueryBuilder) WithTo(to *time.Time) *SolutionQueryBuilder {
	b.to = to
	return b
}

func (b *SolutionQueryBuilder) Build() (SolutionQuery, error) {
	return NewSolutionQuery(b.selections, b.ruleset, b.from, b.to)
}

type NextSolutionsQuery struct {
	currentSelections Selections
	nextSelections    Selections
	ruleset           Ruleset
	from              *time.Time
	to                *time.Time
}

func NewNextSolutionsQuery(
	currentSelections Selections,
	nextSelections Selections,
	ruleset Ruleset,
	from *time.Time,
	to *time.Time,
) (NextSolutionsQuery, error) {
	if err := validateRuleset(ruleset); err != nil {
		return NextSolutionsQuery{}, err
	}

	if err := validateTimestamps(from, to); err != nil {
		return NextSolutionsQuery{}, err
	}

	if err := validateSelections(ruleset, currentSelections); err != nil {
		return NextSolutionsQuery{}, err
	}

	if err := validateSelections(ruleset, nextSelections); err != nil {
		return NextSolutionsQuery{}, err
	}

	return NextSolutionsQuery{
		currentSelections: currentSelections,
		nextSelections:    nextSelections,
		ruleset:           ruleset,
		from:              from,
		to:                to,
	}, nil
}

func (nq NextSolutionsQuery) prepareRuleset() (Ruleset, error) {
	preparedRuleset, err := nq.ruleset.modifyForQuery(
		nq.currentSelections,
		nq.from,
		nq.to,
	)
	if err != nil {
		return Ruleset{}, err
	}

	if err = preparedRuleset.setCompositeSelectionConstraints(nq.nextSelections); err != nil {
		return Ruleset{}, err
	}

	return preparedRuleset, nil
}

func (nq NextSolutionsQuery) emptyNextSelections() bool {
	return len(nq.nextSelections) == 0
}

type nextSolutionQueryPartitioner struct {
	initialQuery NextSolutionsQuery
	weightGroups []weights.Weights
}

func newNextSolutionQueryPartitioner(
	query NextSolutionsQuery,
) (nextSolutionQueryPartitioner, error) {
	ruleset, err := query.prepareRuleset()
	if err != nil {
		return nextSolutionQueryPartitioner{}, err
	}

	weightGroups, err := calculateNextWeightGroups(
		ruleset,
		query.currentSelections,
		query.nextSelections,
	)
	if err != nil {
		return nextSolutionQueryPartitioner{}, err
	}

	return nextSolutionQueryPartitioner{
		initialQuery: query,
		weightGroups: weightGroups,
	}, nil
}

func (p nextSolutionQueryPartitioner) batchable() (NextSolutionsQuery, error) {
	allSelections := p.initialQuery.nextSelections.copy()
	var batchableSelections Selections
	for i, group := range p.weightGroups {
		if group.WeightsTooLarge() {
			continue
		}

		selection := allSelections[i]
		batchableSelections = append(batchableSelections, selection)
	}

	query, err := NewNextSolutionsQuery(
		p.initialQuery.currentSelections,
		batchableSelections,
		p.initialQuery.ruleset,
		p.initialQuery.from,
		p.initialQuery.to,
	)
	if err != nil {
		return NextSolutionsQuery{}, err
	}

	return query, nil
}

func (p nextSolutionQueryPartitioner) saturated() (NextSolutionsQuery, error) {
	allSelections := p.initialQuery.nextSelections.copy()
	var saturatedSelections Selections
	for i, group := range p.weightGroups {
		if group.WeightsTooLarge() {
			selection := allSelections[i]
			saturatedSelections = append(saturatedSelections, selection)
		}
	}

	query, err := NewNextSolutionsQuery(
		p.initialQuery.currentSelections,
		saturatedSelections,
		p.initialQuery.ruleset,
		p.initialQuery.from,
		p.initialQuery.to,
	)
	if err != nil {
		return NextSolutionsQuery{}, err
	}

	return query, nil
}
