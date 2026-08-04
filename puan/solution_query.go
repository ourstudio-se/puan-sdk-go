package puan

import (
	"time"
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
