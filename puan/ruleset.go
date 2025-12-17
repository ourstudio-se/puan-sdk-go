package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

type Ruleset struct {
	polyhedron           *pldag.Polyhedron
	selectableVariables  []string
	dependentVariables   []string
	independentVariables []string
	preferredVariables   []string
	periodVariables      TimeBoundVariables
}

// For when creating a rule set from a serialized representation
// When setting up new rule sets, use RulesetCreator instead
func HydrateRuleSet(
	aMatrix [][]int,
	bVector []int,
	dependentVariables []string,
	independentVariables []string,
	selectableVariables []string,
	preferredVariables []string,
	periodVariables TimeBoundVariables,
) (Ruleset, error) {
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	return newRuleset(
		polyhedron,
		selectableVariables,
		dependentVariables,
		independentVariables,
		preferredVariables,
		periodVariables,
	)
}

func newRuleset(
	polyhedron *pldag.Polyhedron,
	selectableVariables []string,
	dependentVariables []string,
	independentVariables []string,
	preferredVariables []string,
	periodVariables TimeBoundVariables,
) (Ruleset, error) {
	if polyhedron == nil {
		return Ruleset{}, errors.Errorf(
			"%w: polyhedron cannot be nil",
			puanerror.InvalidArgument,
		)
	}

	err := validateVariables(
		selectableVariables,
		dependentVariables,
		independentVariables,
		preferredVariables,
		periodVariables.ids(),
	)
	if err != nil {
		return Ruleset{}, err
	}

	return Ruleset{
		polyhedron:           polyhedron,
		selectableVariables:  selectableVariables,
		dependentVariables:   dependentVariables,
		independentVariables: independentVariables,
		preferredVariables:   preferredVariables,
		periodVariables:      periodVariables,
	}, nil
}

func validateVariables(selectable, dependent, independent, preferreds, periods []string) error {
	if utils.ContainsAny(dependent, independent) {
		return errors.Errorf(
			"%w: dependent and independent variables cannot share variables",
			puanerror.InvalidArgument,
		)
	}

	var combined []string
	combined = append(combined, dependent...)
	combined = append(combined, independent...)

	if len(combined) == 0 {
		return errors.Errorf(
			"%w: dependent and independent variables cannot both be empty",
			puanerror.InvalidArgument,
		)
	}

	if !utils.ContainsAll(combined, selectable) {
		return errors.Errorf(
			"%w: selectable variables must exist in dependent or independent variables",
			puanerror.InvalidArgument,
		)
	}

	if !utils.ContainsAll(dependent, preferreds) {
		return errors.Errorf(
			"%w: preferred variables must exist in dependent variables",
			puanerror.InvalidArgument,
		)
	}

	if !utils.ContainsAll(dependent, periods) {
		return errors.Errorf(
			"%w: period variables must exist in dependent variables",
			puanerror.InvalidArgument,
		)
	}

	return nil
}

func (r *Ruleset) Polyhedron() *pldag.Polyhedron {
	return r.polyhedron
}

func (r *Ruleset) SelectableVariables() []string {
	return r.selectableVariables
}

func (r *Ruleset) PreferredVariables() []string {
	return r.preferredVariables
}

func (r *Ruleset) DependentVariables() []string {
	return r.dependentVariables
}

func (r *Ruleset) IndependentVariables() []string {
	return r.independentVariables
}

func (r *Ruleset) PeriodVariables() TimeBoundVariables {
	return r.periodVariables
}

func (r *Ruleset) RemoveSupportVariables(solution Solution) Solution {
	nonSupportVariables := []string{}
	nonSupportVariables = append(nonSupportVariables, r.selectableVariables...)
	nonSupportVariables = append(nonSupportVariables, r.periodVariables.ids()...)

	return solution.Extract(nonSupportVariables...)
}

func (r *Ruleset) FindPeriodInSolution(solution Solution) (Period, error) {
	var period *Period
	for _, periodVariable := range r.periodVariables {
		if isSet := solution[periodVariable.variable]; isSet == 1 {
			if period != nil {
				return Period{},
					errors.Errorf(
						"%w: multiple periods found: %v and %v",
						puanerror.InvalidArgument,
						period,
						periodVariable.period,
					)
			}
			period = &periodVariable.period
		}
	}

	if period == nil {
		return Period{}, errors.Errorf(
			"%w: period not found in solution",
			puanerror.NotFound,
		)
	}

	return *period, nil
}

func (r *Ruleset) copy() Ruleset {
	aMatrix := make([][]int, len(r.polyhedron.A()))
	copy(aMatrix, r.polyhedron.A())

	bVector := make([]int, len(r.polyhedron.B()))
	copy(bVector, r.polyhedron.B())

	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)

	dependantVariableIDs := make([]string, len(r.dependentVariables))
	copy(dependantVariableIDs, r.dependentVariables)

	independentVariablesIDs := make([]string, len(r.independentVariables))
	copy(independentVariablesIDs, r.independentVariables)

	selectableVariables := make([]string, len(r.selectableVariables))
	copy(selectableVariables, r.selectableVariables)

	preferredIDs := make([]string, len(r.preferredVariables))
	copy(preferredIDs, r.preferredVariables)

	periodVariables := make([]TimeBoundVariable, len(r.periodVariables))
	copy(periodVariables, r.periodVariables)

	return Ruleset{
		polyhedron:           polyhedron,
		selectableVariables:  selectableVariables,
		dependentVariables:   dependantVariableIDs,
		independentVariables: independentVariablesIDs,
		preferredVariables:   preferredIDs,
		periodVariables:      periodVariables,
	}
}

type querySpecification struct {
	ruleset    Ruleset
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
	constraint, err := pldag.NewAtLeastConstraint(dedupedIDs, len(dedupedIDs))
	if err != nil {
		return pldag.Constraint{}, err
	}

	return constraint, nil
}

func (r *Ruleset) setConstraintIfNotExist(constraint pldag.Constraint) error {
	if r.constraintExists(constraint) {
		return nil
	}

	return r.setConstraint(constraint)
}

func (r *Ruleset) constraintExists(constraint pldag.Constraint) bool {
	return utils.Contains(r.dependentVariables, constraint.ID())
}

func (r *Ruleset) setConstraint(constraint pldag.Constraint) error {
	r.polyhedron.AddEmptyColumn()
	r.dependentVariables = append(r.dependentVariables, constraint.ID())

	supportImpliesConstraint, constraintImpliesSupport :=
		constraint.ToAuxiliaryConstraintsWithSupport()

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
	row := make([]int, len(r.dependentVariables))

	for id, value := range coefficients {
		idIndex, err := utils.IndexOf(r.dependentVariables, id)
		if err != nil {
			return nil, errors.Errorf(
				"%w: variable %s not found in dependent variables",
				puanerror.NotFound,
				id,
			)
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

	if err = r.setConstraint(constraint); err != nil {
		return err
	}

	return r.assume(constraint.ID())
}

func (r *Ruleset) assume(id string) error {
	constraint := pldag.NewAssumedConstraint(id)
	return r.setAuxiliaryConstraint(constraint)
}
