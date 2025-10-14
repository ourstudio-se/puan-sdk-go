package puan

import "github.com/ourstudio-se/puan-sdk-go/internal/utils"

type solverClient interface {
	Solve(query *Query) (Solution, error)
}

type solver struct {
	client     solverClient
	ruleset    *RuleSet
	selections Selections
}

func newSolver(client solverClient, ruleset *RuleSet, selections Selections) *solver {
	return &solver{
		client:     client,
		ruleset:    ruleset,
		selections: selections,
	}
}

func Solve(client solverClient, ruleset *RuleSet, selections Selections) (Solution, error) {
	s := newSolver(client, ruleset, selections)

	dependantSelections, independantSelections, err := ruleset.categorizeSelections(selections)
	if err != nil {
		return Solution{}, err
	}

}

func (s *solver) categorizeSelections() error {
	err := s.ruleset.validateSelectionIDs(s.selections.ids())
	if err != nil {
		return err
	}

	ependentSelections := extractDependantSelections(
		s.selections,
		s.ruleset.independentVariables,
	)
	independentSelections := extractIndependentSelections(
		s.selections,
		s.ruleset.independentVariables,
	)

}

func extractDependantSelections(selections Selections, freeVariables []string) Selections {
	var newSelections Selections
	for _, selection := range selections {
		s := extractDependantSelection(selection, freeVariables)
		if s != nil {
			newSelections = append(newSelections, *s)
		}
	}

	return newSelections
}

func extractDependantSelection(selection Selection, freeVariables []string) *Selection {
	if !utils.ContainsAny(selection.ids(), freeVariables) {
		return &selection
	}

	if utils.Contains(freeVariables, selection.id) {
		return nil
	}
	// kasta fel vi har en fri variabel i en composite selection i validate Selection

	newSubselectionIDs := utils.Without(selection.subSelectionIDs, freeVariables)
	withoutFreeVariables := newSelection(
		selection.action,
		selection.id,
		newSubselectionIDs,
	)

	return &withoutFreeVariables
}

func extractIndependentSelections(selections Selections, independentVariables []string) IndependentSelections {
	var independentSelections IndependentSelections
	for _, variable := range independentVariables {
		selection := extractIndependentSelection(selections, variable)
		if selection != nil {
			independentSelections = append(independentSelections, *selection)
		}
	}

	return independentSelections
}

func extractIndependentSelection(selections Selections, freeVariable string) *IndependentSelection {
	// reverse loop for prioritizing the latest selection action
	for i := len(selections) - 1; i >= 0; i-- {
		selection := selections[i]
		if utils.Contains(selection.ids(), freeVariable) {
			independentSelection := selection.toIndependentSelection()
			return &independentSelection
		}
	}

	return nil
}
