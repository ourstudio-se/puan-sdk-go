package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type SolverClient interface {
	Solve(query *Query) (Solution, error)
}

type SolutionCreator struct {
	SolverClient
}

func NewSolutionCreator(
	client SolverClient,
) *SolutionCreator {
	return &SolutionCreator{
		SolverClient: client,
	}
}

func (c *SolutionCreator) Create(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (Solution, error) {
	err := validateSelections(selections, ruleset)
	if err != nil {
		return nil, err
	}

	// nolint:lll
	dependantSelections, independentSelections := categorizeSelections(selections, ruleset.independentVariables)

	query, err := newQuery(dependantSelections, ruleset, from)
	if err != nil {
		return nil, err
	}

	return c.solve(query, independentSelections, ruleset)
}

func (c *SolutionCreator) solve(
	query *Query,
	independentSelections Selections,
	ruleset Ruleset,
) (Solution, error) {
	rawSolution, err := c.Solve(query)
	if err != nil {
		return Solution{}, err
	}

	solution, err := ruleset.RemoveSupportVariables(rawSolution)
	if err != nil {
		return Solution{}, err
	}

	// nolint:lll
	primitiveSolution := solution.applyIndependentVariables(ruleset.independentVariables, independentSelections)

	return primitiveSolution, nil
}

func validateSelections(selections Selections, ruleset Ruleset) error {
	for _, selection := range selections {
		if !utils.ContainsAll(ruleset.selectableVariables, selection.ids()) {
			return errors.Errorf(
				"invalid selection: %v",
				selection,
			)
		}

		hasSubSelection := len(selection.subSelectionIDs) > 0
		if hasSubSelection {
			if utils.ContainsAny(selection.ids(), ruleset.independentVariables) {
				return errors.Errorf(
					"independent variables cannot be part of a composite selections: %v",
					selection,
				)
			}
		}
	}

	return nil
}

func categorizeSelections(
	selections Selections,
	independentVariables []string,
) (Selections, Selections) {
	var dependantSelections Selections
	var independentSelections Selections

	for _, selection := range selections {
		isIndependent := utils.Contains(independentVariables, selection.id)
		if isIndependent {
			independentSelections = append(independentSelections, selection)
		} else {
			dependantSelections = append(dependantSelections, selection)
		}
	}

	return dependantSelections, independentSelections
}

func newQuery(selections Selections, ruleset Ruleset, from *time.Time) (*Query, error) {
	extendedSelections := selections.modifySelections()
	impactingSelections := getImpactingSelections(extendedSelections)

	specification, err := ruleset.newQuerySpecification(impactingSelections, from)
	if err != nil {
		return nil, err
	}

	weights := weights.Calculate(
		specification.ruleset.selectableVariables,
		specification.selections,
		specification.ruleset.preferredVariables,
		specification.ruleset.periodVariables.ids(),
	)

	query := NewQuery(
		specification.ruleset.polyhedron,
		specification.ruleset.dependantVariables,
		weights,
	)

	return query, nil
}
