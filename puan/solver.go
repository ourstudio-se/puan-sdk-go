package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type solverClient interface {
	Solve(query *Query) (Solution, error)
}

type solver struct {
	client     solverClient
	ruleset    *RuleSet
	selections Selections
	from       *time.Time
}

func newSolver(client solverClient, ruleset *RuleSet, selections Selections, from *time.Time) *solver {
	return &solver{
		client:     client,
		ruleset:    ruleset,
		selections: selections,
		from:       from,
	}
}

func Solve(
	client solverClient,
	ruleset *RuleSet,
	selections Selections,
	from *time.Time,
) (Solution, error) {
	s := newSolver(
		client,
		ruleset,
		selections,
		from,
	)

	query, err := s.newQuery()
	if err != nil {
		return Solution{}, err
	}

	solution, err := s.solve(query)
	if err != nil {
		return Solution{}, err
	}

	return solution, nil
}

func (s *solver) newQuery() (*Query, error) {
	impactingSelections, err := s.getImpactingSelections()
	if err != nil {
		return nil, err
	}

	specification, err := s.ruleset.newQuerySpecification(impactingSelections, s.from)
	if err != nil {
		return nil, err
	}

	weights := calculateWeights(
		specification.ruleSet.selectableVariables,
		specification.querySelections,
		specification.ruleSet.preferredVariables,
		specification.ruleSet.periodVariables.ids(),
	)

	query := NewQuery(
		specification.ruleSet.polyhedron,
		specification.ruleSet.dependantVariables,
		weights,
	)

	return query, nil
}

func (s *solver) solve(query *Query) (Solution, error) {
	solution, err := s.client.Solve(query)
	if err != nil {
		return Solution{}, err
	}

	independentSelections := extractIndependentSelections(
		s.selections,
		s.ruleset.independentVariables,
	)

	for _, independentSelection := range independentSelections {
		solution[independentSelection.id] = independentSelection.toSolutionValue()
	}

	return solution, nil
}

func (s *solver) getImpactingSelections() (Selections, error) {
	err := s.validateSelections()
	if err != nil {
		return nil, err
	}

	impactingSelections := calculateImpactingSelections(
		s.selections,
		s.ruleset.independentVariables,
	)

	return impactingSelections, nil
}

func (s *solver) validateSelections() error {
	independentVariables := s.ruleset.independentVariables
	for _, selection := range s.selections {
		if !utils.ContainsAll(s.ruleset.selectableVariables, selection.ids()) {
			return errors.Errorf(
				"invalid selection: %v",
				selection,
			)
		}

		if utils.ContainsAny(selection.subSelectionIDs, independentVariables) {
			return errors.Errorf(
				"independent variables cannot be part of a composite selection: %v",
				selection,
			)
		}
	}

	return nil
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

func extractIndependentSelections(selections Selections, independentVariableIDs []string) IndependentSelections {
	var independentSelections IndependentSelections
	for _, id := range independentVariableIDs {
		selection := extractIndependentSelection(selections, id)
		if selection != nil {
			independentSelections = append(independentSelections, *selection)
		}
	}

	return independentSelections
}

func extractIndependentSelection(selections Selections, independentVariableID string) *IndependentSelection {
	// reverse loop for prioritizing the latest selection action
	for i := len(selections) - 1; i >= 0; i-- {
		selection := selections[i]
		if selection.id == independentVariableID {
			independentSelection := selection.toIndependentSelection()

			return &independentSelection
		}
	}

	return nil
}
