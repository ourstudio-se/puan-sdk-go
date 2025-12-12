package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

// WEIGHT_SATURATION_LIMIT is set to 2^55.
// The limit is set lower than 2^63 to allow for some headroom
// and enable early detection before reaching the overflow limit.
// Note: Weights for selected variables increase exponentially with the number of selections.
// The highest weight is calculated as 2^(n-1) * (c + 1),
// where n is the number of selections and c is a constant
// derived from the sum of none selected primitives, preferred, and period weights.
const WEIGHT_SATURATION_LIMIT = 36028797018963968

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
) (SolutionEnvelope, error) {
	err := validateSelections(selections, ruleset)
	if err != nil {
		return SolutionEnvelope{}, err
	}

	dependantSelections, independentSelections :=
		categorizeSelections(selections, ruleset.independentVariables)

	dependentSolution, weightsToLarge, err := c.findDependentSolution(
		dependantSelections,
		ruleset,
		from,
	)
	if err != nil {
		return SolutionEnvelope{}, err
	}

	independentSolution := findIndependentSolution(ruleset.independentVariables, independentSelections)

	solution := dependentSolution.merge(independentSolution)

	return SolutionEnvelope{
		Solution:       solution,
		WeightsToLarge: weightsToLarge,
	}, nil
}

func (c *SolutionCreator) findDependentSolution(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (Solution, bool, error) {
	query, err := newQuery(selections, ruleset, from)
	if err != nil {
		return nil, false, err
	}

	isToLarge := query.weights.MaxWeight() > WEIGHT_SATURATION_LIMIT

	solution, err := c.Solve(query)
	if err != nil {
		return nil, false, err
	}

	primitiveSolution := ruleset.RemoveSupportVariables(solution)

	return primitiveSolution, isToLarge, nil
}

func findIndependentSolution(independentVariables []string, selections Selections) Solution {
	solution := make(Solution, len(independentVariables))
	for _, variable := range independentVariables {
		solution[variable] = independentSolutionValue(variable, selections)
	}

	return solution
}

func independentSolutionValue(variableID string, selections Selections) int {
	// reverse loop for prioritizing the latest selection action
	for i := len(selections) - 1; i >= 0; i-- {
		selection := selections[i]
		if selection.id == variableID {
			if selection.action == ADD {
				return 1
			}

			return 0
		}
	}

	return 0
}

func validateSelections(selections Selections, ruleset Ruleset) error {
	for _, selection := range selections {
		if !utils.ContainsAll(ruleset.selectableVariables, selection.ids()) {
			return errors.Errorf(
				"%w: selection contains non-selectable variables: %v",
				puanerror.InvalidArgument,
				selection,
			)
		}

		hasSubSelection := len(selection.subSelectionIDs) > 0
		if hasSubSelection {
			if utils.ContainsAny(selection.ids(), ruleset.independentVariables) {
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

	dependentSelectableVariables := utils.Without(
		specification.ruleset.selectableVariables,
		specification.ruleset.independentVariables,
	)

	weights := weights.Calculate(
		dependentSelectableVariables,
		specification.selections,
		specification.ruleset.preferredVariables,
		specification.ruleset.periodVariables.ids(),
	)

	query := NewQuery(
		specification.ruleset.polyhedron,
		specification.ruleset.dependentVariables,
		weights,
	)

	return query, nil
}
