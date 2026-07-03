package puan

import (
	"time"

	"github.com/go-errors/errors"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

type SolutionQuery struct {
	selections Selections
	ruleset    Ruleset
	from       *time.Time
	to         *time.Time
}

func (query SolutionQuery) validate() error {
	if err := query.validateRuleset(); err != nil {
		return err
	}

	if err := query.validateTimestamps(); err != nil {
		return err
	}

	if err := query.validateSelections(); err != nil {
		return err
	}

	return nil
}

func (query SolutionQuery) validateRuleset() error {
	if query.ruleset.polyhedron == nil {
		return errors.Errorf("%w: ruleset is required", puanerror.InvalidArgument)
	}
	return nil
}

func (query SolutionQuery) validateTimestamps() error {
	if query.from != nil && query.to != nil {
		if query.from.After(*query.to) {
			return errors.Errorf(
				"%w: from '%s' must be before to '%s'",
				puanerror.InvalidArgument,
				query.from,
				query.to,
			)
		}
	}
	return nil
}

func (query SolutionQuery) validateSelections() error {
	for _, selection := range query.selections {
		if !utils.ContainsAll(query.ruleset.selectableVariables, selection.IDs()) {
			return errors.Errorf(
				"%w: selection contains non-selectable variables: %v",
				puanerror.InvalidArgument,
				selection,
			)
		}

		hasSubSelection := len(selection.subSelectionIDs) > 0
		if hasSubSelection {
			if utils.ContainsAny(selection.IDs(), query.ruleset.independentVariables) {
				return errors.Errorf(
					"%w: independent variables cannot be part of a composite selections: %v",
					puanerror.InvalidArgument,
					selection,
				)
			}
		}
	}

	return nil
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

func (b *SolutionQueryBuilder) Build() SolutionQuery {
	return SolutionQuery{
		selections: b.selections,
		ruleset:    b.ruleset,
		from:       b.from,
		to:         b.to,
	}
}
