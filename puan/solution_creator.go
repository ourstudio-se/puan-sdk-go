package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
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
) (SolutionEnvelope, error) {
	err := validateSelections(selections, ruleset)
	if err != nil {
		return SolutionEnvelope{}, err
	}

	dependantSelections, independentSelections :=
		categorizeSelections(selections, ruleset.independentVariables)

	envelope, err := c.calculateSolveSolution(
		dependantSelections,
		ruleset,
		from,
	)
	if err != nil {
		err = updateSolveError(err, ruleset, from)
		return SolutionEnvelope{}, err
	}

	dependentSolution := envelope.Solution()
	weightsTooLarge := envelope.WeightsTooLarge()

	independentSolution := calculateIndependentSolution(
		ruleset.independentVariables,
		independentSelections,
	)

	solution := dependentSolution.merge(independentSolution)

	return SolutionEnvelope{
		solution:        solution,
		weightsTooLarge: weightsTooLarge,
	}, nil
}

func (c *SolutionCreator) calculateSolveSolution(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (SolutionEnvelope, error) {
	query, err := newQuery(selections, ruleset, from)
	if err != nil {
		return SolutionEnvelope{}, err
	}

	tooLarge := query.weights.WeightsTooLarge()

	if tooLarge {
		return c.calculateMultiSolveSolution(selections, ruleset, from)
	}

	solution, err := c.Solve(query)
	if err != nil {
		return SolutionEnvelope{}, err
	}

	primitiveSolution := ruleset.RemoveSupportVariables(solution)

	return SolutionEnvelope{
		solution: primitiveSolution,
	}, nil
}

func (c *SolutionCreator) calculateMultiSolveSolution(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (SolutionEnvelope, error) {
	earlierSelections, laterSelections := selections.split()

	solutionFromLaterSelections, err := c.calculateSolveSolution(laterSelections, ruleset, from)
	if err != nil {
		return SolutionEnvelope{}, err
	}

	updatedRuleset, err := c.newRulesetWithAssumedSolution(
		ruleset,
		laterSelections,
		solutionFromLaterSelections.Solution(),
	)
	if err != nil {
		return SolutionEnvelope{}, err
	}

	return c.calculateSolveSolution(earlierSelections, updatedRuleset, from)
}

func (c *SolutionCreator) newRulesetWithAssumedSolution(
	ruleset Ruleset,
	selections Selections,
	solution Solution,
) (Ruleset, error) {
	newRuleset := ruleset.copy()

	selectedIDs := c.getSelectedIDs(selections, solution)

	for _, id := range selectedIDs {
		err := newRuleset.assume(id)
		if err != nil {
			return Ruleset{}, err
		}
	}

	return newRuleset, nil
}

func (c *SolutionCreator) getSelectedIDs(
	selections Selections,
	solution Solution,
) []string {
	var selectedIDs []string
	for _, id := range selections.ids() {
		isSelected := solution.isSelected(id)
		if isSelected {
			selectedIDs = append(selectedIDs, id)
		}
	}

	return selectedIDs
}

func calculateIndependentSolution(independentVariables []string, selections Selections) Solution {
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

func updateSolveError(
	err error,
	ruleset Ruleset,
	from *time.Time,
) error {
	solverFailed := errors.Is(err, puanerror.SolverFailed)
	if solverFailed {
		invalidTime := !ruleset.isValidFromTime(from)
		if invalidTime {
			return errors.Errorf(
				"%w: from '%s' is not valid for the ruleset",
				puanerror.InvalidArgument,
				from,
			)
		}
	}

	return err
}
