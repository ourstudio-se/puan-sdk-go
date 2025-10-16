package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type Ruleset struct {
	polyhedron           *pldag.Polyhedron
	selectableVariables  []string
	dependantVariables   []string
	independentVariables []string
	preferredVariables   []string
	periodVariables      timeBoundVariables
}

// For when creating a rule set from a serialized representation
// When setting up new rule sets, use RuleSetCreator instead
func HydrateRuleSet(
	aMatrix [][]int,
	bVector []int,
	dependentVariables []string,
	independentVariables []string,
	selectableVariables []string,
	preferredVariables []string,
) *Ruleset {
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	return &Ruleset{
		polyhedron:           polyhedron,
		selectableVariables:  selectableVariables,
		dependantVariables:   dependentVariables,
		independentVariables: independentVariables,
		preferredVariables:   preferredVariables,
	}
}

func (r *Ruleset) Polyhedron() *pldag.Polyhedron {
	return r.polyhedron
}

func (r *Ruleset) SelectableVariables() []string {
	return r.selectableVariables
}

func (r *Ruleset) DependantVariables() []string {
	return r.dependantVariables
}

func (r *Ruleset) IndependentVariables() []string {
	return r.independentVariables
}

func (r *Ruleset) PreferredVariables() []string {
	return r.preferredVariables
}

func (r *Ruleset) RemoveSupportVariables(solution Solution) (Solution, error) {
	nonSupportVariables := []string{}
	nonSupportVariables = append(nonSupportVariables, r.periodVariables.ids()...)
	nonSupportVariables = append(nonSupportVariables, r.selectableVariables...)

	return solution.Extract(nonSupportVariables...)
}

func (r *Ruleset) FindPeriodInSolution(solution Solution) (Period, error) {
	var period *Period
	for _, periodVariable := range r.periodVariables {
		if isSet := solution[periodVariable.variable]; isSet == 1 {
			if period != nil {
				return Period{},
					errors.Errorf(
						"multiple periods found: %v and %v",
						period,
						periodVariable.period,
					)
			}
			period = &periodVariable.period
		}
	}

	if period == nil {
		return Period{}, errors.New("period not found for solution")
	}

	return *period, nil
}

func (r *Ruleset) copy() *Ruleset {
	aMatrix := make([][]int, len(r.polyhedron.A()))
	copy(aMatrix, r.polyhedron.A())

	bVector := make([]int, len(r.polyhedron.B()))
	copy(bVector, r.polyhedron.B())

	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)

	dependantVariableIDs := make([]string, len(r.dependantVariables))
	copy(dependantVariableIDs, r.dependantVariables)

	independentVariablesIDs := make([]string, len(r.independentVariables))
	copy(independentVariablesIDs, r.independentVariables)

	selectableVariables := make([]string, len(r.selectableVariables))
	copy(selectableVariables, r.selectableVariables)

	preferredIDs := make([]string, len(r.preferredVariables))
	copy(preferredIDs, r.preferredVariables)

	periodVariables := make([]timeBoundVariable, len(r.periodVariables))
	copy(periodVariables, r.periodVariables)

	return &Ruleset{
		polyhedron:           polyhedron,
		selectableVariables:  selectableVariables,
		dependantVariables:   dependantVariableIDs,
		independentVariables: independentVariablesIDs,
		preferredVariables:   preferredIDs,
		periodVariables:      periodVariables,
	}
}

type querySpecification struct {
	ruleset    *Ruleset
	selections weights.Selections
}

func (r *Ruleset) newQuerySpecification(
	selections Selections,
	from *time.Time,
) (*querySpecification, error) {
	ruleset := r.copy()

	weightSelections, err := ruleset.newWeightSelections(selections)
	if err != nil {
		return nil, err
	}

	if from != nil {
		err = ruleset.forbidPassedPeriods(*from)
		if err != nil {
			return nil, err
		}
	}

	return &querySpecification{
		ruleset:    ruleset,
		selections: weightSelections,
	}, nil
}

func (r *Ruleset) newWeightSelections(selections Selections) (weights.Selections, error) {
	weightSelections := make(weights.Selections, len(selections))
	for i, selection := range selections {
		weightSelection, err := r.newWeighSelection(selection)
		if err != nil {
			return nil, err
		}

		weightSelections[i] = weightSelection
	}

	return weightSelections, nil
}

func (r *Ruleset) newWeighSelection(selection Selection) (weights.Selection, error) {
	id, err := r.obtainQuerySelectionID(selection)
	if err != nil {
		return weights.Selection{}, err
	}

	weightSelection, err := weights.NewSelection(id, weights.Action(selection.action))
	if err != nil {
		return weights.Selection{}, err
	}

	return weightSelection, nil
}

func (r *Ruleset) obtainQuerySelectionID(selection Selection) (string, error) {
	if selection.isComposite() {
		return r.setCompositeSelectionConstraint(selection.ids())
	}

	return selection.id, nil
}

func (r *Ruleset) setCompositeSelectionConstraint(ids []string) (string, error) {
	constraint, err := newCompositeSelectionConstraint(ids)
	if err != nil {
		return "", err
	}

	err = r.setConstraintIfNotExist(constraint)
	if err != nil {
		return "", err
	}

	return constraint.ID(), nil
}

func newCompositeSelectionConstraint(ids []string) (pldag.Constraint, error) {
	dedupedIDs := utils.Dedupe(ids)
	return pldag.NewAtLeastConstraint(dedupedIDs, len(dedupedIDs))
}

func (r *Ruleset) setConstraintIfNotExist(constraint pldag.Constraint) error {
	if r.constraintExists(constraint) {
		return nil
	}

	return r.setConstraint(constraint)
}

func (r *Ruleset) constraintExists(constraint pldag.Constraint) bool {
	return utils.Contains(r.dependantVariables, constraint.ID())
}

// nolint:lll
func (r *Ruleset) setConstraint(constraint pldag.Constraint) error {
	r.polyhedron.AddEmptyColumn()
	r.dependantVariables = append(r.dependantVariables, constraint.ID())

	supportImpliesConstraint, constraintImpliesSupport := constraint.ToAuxiliaryConstraintsWithSupport()

	err := r.setAuxiliaryConstraint(supportImpliesConstraint)
	if err != nil {
		return err
	}

	err = r.setAuxiliaryConstraint(constraintImpliesSupport)
	if err != nil {
		return err
	}

	return nil
}

func (r *Ruleset) setAuxiliaryConstraint(constraint pldag.AuxiliaryConstraint) error {
	row, err := r.newRow(constraint.Coefficients())
	if err != nil {
		return err
	}

	r.polyhedron.Extend(row, constraint.Bias())

	return nil
}

func (r *Ruleset) newRow(coefficients pldag.Coefficients) ([]int, error) {
	row := make([]int, len(r.dependantVariables))

	for id, value := range coefficients {
		idIndex, err := utils.IndexOf(r.dependantVariables, id)
		if err != nil {
			return nil, err
		}

		row[idIndex] = value
	}

	return row, nil
}

func (r *Ruleset) forbidPassedPeriods(from time.Time) error {
	passedPeriods := r.periodVariables.passed(from)
	passedPeriodIDs := passedPeriods.ids()

	if len(passedPeriodIDs) == 0 {
		return nil
	}

	constraint, err := pldag.NewAtMostConstraint(passedPeriodIDs, 0)
	if err != nil {
		return err
	}

	if err := r.setConstraint(constraint); err != nil {
		return err
	}

	return r.assume(constraint.ID())
}

func (r *Ruleset) assume(id string) error {
	constraints := pldag.NewAssumedConstraints(id)
	for _, constraint := range constraints {
		err := r.setAuxiliaryConstraint(constraint)
		if err != nil {
			return err
		}
	}

	return nil
}
