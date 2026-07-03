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

func (r *Ruleset) dependentSelectableVariables() []string {
	return utils.Without(r.selectableVariables, r.independentVariables)
}

func (r *Ruleset) RemoveSupportVariables(solution Solution) Solution {
	nonSupportVariables := []string{}
	nonSupportVariables = append(nonSupportVariables, r.selectableVariables...)
	nonSupportVariables = append(nonSupportVariables, r.periodVariables.ids()...)

	return solution.Extract(nonSupportVariables...)
}

func (r *Ruleset) RemoveSupportVariablesForMany(solutions []Solution) []Solution {
	cleaned := make([]Solution, len(solutions))
	for i, solution := range solutions {
		cleaned[i] = r.RemoveSupportVariables(solution)
	}

	return cleaned
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
	aMatrix := r.polyhedron.A()
	aMatrixCopy := make([][]int, len(r.polyhedron.A()))
	for i := range aMatrix {
		aMatrixCopy[i] = make([]int, len(aMatrix[i]))
		copy(aMatrixCopy[i], aMatrix[i])
	}

	bVector := make([]int, len(r.polyhedron.B()))
	copy(bVector, r.polyhedron.B())

	polyhedron := pldag.NewPolyhedron(aMatrixCopy, bVector)

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

func (r *Ruleset) modifyForQuery(
	selections Selections,
	from *time.Time,
	to *time.Time,
) (Ruleset, error) {
	ruleset := r.copy()

	if err := ruleset.setCompositeSelectionConstraints(selections); err != nil {
		return Ruleset{}, err
	}

	if from != nil {
		err := ruleset.forbidPassedPeriods(*from)
		if err != nil {
			return Ruleset{}, err
		}
	}

	if to != nil {
		err := ruleset.forbidFuturePeriods(*to)
		if err != nil {
			return Ruleset{}, err
		}
	}

	return ruleset, nil
}

// Add constraints for composite selections to get a single ID.
// This is needed to set weights
func (r *Ruleset) setCompositeSelectionConstraints(
	selections Selections,
) error {
	for _, selection := range selections {
		if selection.IsComposite() {
			ids := selection.IDs()
			if err := r.setCompositeSelectionConstraint(ids); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Ruleset) newWeightSelections(selections Selections) (weights.Selections, error) {
	weightSelections := make(weights.Selections, len(selections))
	for i, selection := range selections {
		weightSelection, err := r.newWeightSelection(selection)
		if err != nil {
			return nil, err
		}

		weightSelections[i] = weightSelection
	}

	return weightSelections, nil
}

func (r *Ruleset) newWeightSelection(selection Selection) (weights.Selection, error) {
	id, err := r.getWeightSelectionID(selection)
	if err != nil {
		return weights.Selection{}, err
	}

	weightSelection, err := weights.NewSelection(id, weights.Action(selection.action))
	if err != nil {
		return weights.Selection{}, err
	}

	return weightSelection, nil
}

func (r *Ruleset) getWeightSelectionID(selection Selection) (string, error) {
	if selection.IsComposite() {
		constraint, err := newCompositeSelectionConstraint(selection.IDs())
		if err != nil {
			return "", err
		}

		missing := !r.constraintExists(constraint)
		if missing {
			return "", errors.Errorf(
				"Weight selection ID not found for: %v. Is the ruleset prepared for the selection?",
				selection.IDs(),
			)
		}

		return constraint.ID(), nil
	}

	return selection.id, nil
}

func (r *Ruleset) setCompositeSelectionConstraint(ids []string) error {
	constraint, err := newCompositeSelectionConstraint(ids)
	if err != nil {
		return err
	}

	err = r.setConstraintIfNotExist(constraint)
	if err != nil {
		return err
	}

	return nil
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
	passedPeriods := r.periodVariables.earlierThan(from)
	passedPeriodIDs := passedPeriods.ids()

	if len(passedPeriodIDs) == 0 {
		return nil
	}

	return r.assumeNot(passedPeriodIDs...)
}

func (r *Ruleset) forbidFuturePeriods(to time.Time) error {
	futurePeriods := r.periodVariables.laterThan(to)
	futurePeriodIDs := futurePeriods.ids()

	if len(futurePeriodIDs) == 0 {
		return nil
	}

	return r.assumeNot(futurePeriodIDs...)
}

func (r *Ruleset) assume(id string) error {
	constraint := pldag.NewAssumedConstraint(id)
	return r.setAuxiliaryConstraint(constraint)
}

func (r *Ruleset) assumeNot(ids ...string) error {
	notID, err := pldag.NewAtMostConstraint(ids, 0)
	if err != nil {
		return err
	}

	if err = r.setConstraint(notID); err != nil {
		return err
	}

	return r.assume(notID.ID())
}

func (r *Ruleset) isValidFromTime(from *time.Time) bool {
	if r.timeDisabled() {
		return true
	}

	if from == nil {
		return true
	}

	for _, periodVariable := range r.periodVariables {
		isValid := !from.After(periodVariable.period.To())
		if isValid {
			return true
		}
	}

	return false
}

func (r *Ruleset) timeDisabled() bool {
	return len(r.periodVariables) == 0
}
