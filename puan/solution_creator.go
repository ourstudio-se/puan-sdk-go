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
	ruleset *Ruleset,
	from *time.Time,
) (Solution, error) {
	err := validateSelections(selections, ruleset)
	if err != nil {
		return nil, err
	}

	query, err := newQuery(selections, ruleset, from)
	if err != nil {
		return nil, err
	}

	return c.solve(query, selections, ruleset)
}

func (c *SolutionCreator) solve(
	query *Query,
	selections Selections,
	ruleset *Ruleset,
) (Solution, error) {
	solution, err := c.Solve(query)
	if err != nil {
		return Solution{}, err
	}

	primitiveSolution, err := ruleset.RemoveSupportVariables(solution)
	if err != nil {
		return Solution{}, err
	}

	independentSelections := createIndependentSelections(
		selections,
		ruleset.independentVariables,
	)

	for _, independentSelection := range independentSelections {
		primitiveSolution[independentSelection.id] = independentSelection.toSolutionValue()
	}

	return primitiveSolution, nil
}

func validateSelections(selections Selections, ruleset *Ruleset) error {
	for _, selection := range selections {
		if !utils.ContainsAll(ruleset.selectableVariables, selection.ids()) {
			return errors.Errorf(
				"invalid selection: %v",
				selection,
			)
		}

		if utils.ContainsAny(selection.subSelectionIDs, ruleset.independentVariables) {
			return errors.Errorf(
				"independent variables cannot be part of a composite selection: %v",
				selection,
			)
		}
	}

	return nil
}

func newQuery(selections Selections, ruleset *Ruleset, from *time.Time) (*Query, error) {
	impactingSelections := calculateImpactingSelections(selections, ruleset.independentVariables)

	specification, err := ruleset.newQuerySpecification(impactingSelections, from)
	if err != nil {
		return nil, err
	}

	weights := weights.Calculate(
		specification.ruleset.selectableVariables,
		specification.weightSelections,
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

func calculateImpactingSelections(
	selections Selections,
	independentVariables []string,
) Selections {
	var dependantSelections Selections
	for _, selection := range selections {
		s := extractDependantSelection(selection, independentVariables)
		if s != nil {
			dependantSelections = append(dependantSelections, *s)
		}
	}

	extendedSelections := dependantSelections.modifySelections()
	impactingSelections := getImpactingSelections(extendedSelections)

	return impactingSelections
}

func extractDependantSelection(selection Selection, independentVariables []string) *Selection {
	if utils.Contains(independentVariables, selection.id) {
		return nil
	}

	return &selection
}

func createIndependentSelections(
	selections Selections,
	independentVariableIDs []string,
) IndependentSelections {
	var independentSelections IndependentSelections
	for _, id := range independentVariableIDs {
		selection := extractIndependentSelection(selections, id)
		if selection != nil {
			independentSelections = append(
				independentSelections,
				selection.toIndependentSelection(),
			)
		}
	}

	return independentSelections
}

func extractIndependentSelection(
	selections Selections,
	independentVariableID string,
) *Selection {
	// reverse loop for prioritizing the latest selection action
	for i := len(selections) - 1; i >= 0; i-- {
		selection := selections[i]
		if selection.id == independentVariableID {
			return &selection
		}
	}

	return nil
}
