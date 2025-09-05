package puan

import (
	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/utils"
)

type RuleSetCreator struct {
	pldag              *pldag.Model
	preferredVariables []string
}

type RuleSet struct {
	polyhedron         *pldag.Polyhedron
	primitiveVariables []string
	variables          []string
	preferredVariables []string
}

func NewRuleSetCreator() *RuleSetCreator {
	return &RuleSetCreator{
		pldag: pldag.New(),
	}
}

func (c RuleSetCreator) PLDAG() *pldag.Model {
	return c.pldag
}

func (c RuleSetCreator) SetPreferreds(id ...string) error {
	err := c.validatePreferredIDs(id)
	if err != nil {
		return err
	}

	c.preferredVariables = append(c.preferredVariables, id...)

	return nil
}

func (c RuleSetCreator) validatePreferredIDs(ids []string) error {
	if utils.ContainsDuplicates(ids) {
		return errors.New("duplicated preferred variables")
	}

	if utils.ContainsAny(c.preferredVariables, ids) {
		return errors.New("preferred variable already added")
	}

	missingIDs := !utils.ContainsAll(c.pldag.Variables(), ids)
	if missingIDs {
		return errors.New("preferred variable not in model")
	}

	return nil
}

func (c RuleSetCreator) Create() *RuleSet {
	polyhedron := c.pldag.GeneratePolyhedron()
	variables := c.pldag.Variables()
	primitiveVariables := c.PLDAG().PrimitiveVariables()

	return &RuleSet{
		polyhedron:         polyhedron,
		primitiveVariables: primitiveVariables,
		variables:          variables,
		preferredVariables: c.preferredVariables,
	}
}

func (r *RuleSet) Polyhedron() *pldag.Polyhedron {
	return r.polyhedron
}

func (r *RuleSet) PrimitiveVariables() []string {
	return r.primitiveVariables
}

func (r *RuleSet) Variables() []string {
	return r.variables
}

func (r *RuleSet) PreferredVariables() []string {
	return r.preferredVariables
}

type querySpecification struct {
	ruleSet           *RuleSet
	selectionIDLookUp map[Selection]string
}

func (r *RuleSet) NewQuery(selections Selections) (*Query, error) {
	impactingSelections := getImpactingSelections(selections)
	specification, err := r.newQuerySpecification(impactingSelections)
	if err != nil {
		return nil, err
	}

	selectedIDs, err := getSelectedIDs(impactingSelections, specification.selectionIDLookUp)
	if err != nil {
		return nil, err
	}

	objective := CalculateObjective(specification.ruleSet.primitiveVariables, selectedIDs, specification.ruleSet.preferredVariables)

	return NewQuery(specification.ruleSet.polyhedron, specification.ruleSet.variables, objective), nil
}

func getSelectedIDs(selections Selections, idLookUp map[Selection]string) ([]string, error) {
	ids := make([]string, len(selections))
	for i, selection := range selections {
		id, ok := idLookUp[selection]
		if !ok {
			return nil, errors.New("selection not found")
		}

		ids[i] = id
	}

	return ids, nil
}

func (r *RuleSet) CalculateSelectedIDs(selections Selections) ([]string, error) {
	impactingSelections := getImpactingSelections(selections)

	var selectedIDs []string
	for _, selection := range impactingSelections {
		hasSubselection := selection.subSelectionID != nil
		if hasSubselection {
			auxiliaryID, err := r.prepareCompositeSelection(selection.id, *selection.subSelectionID)
			if err != nil {
				return nil, err
			}
			r.variables = append(r.variables, auxiliaryID)
			selectedIDs = append(selectedIDs, auxiliaryID)
		} else {
			selectedIDs = append(selectedIDs, selection.id)
		}
	}

	return selectedIDs, nil
}

func (r *RuleSet) copy() *RuleSet {
	aMatrix := make([][]int, len(r.polyhedron.A()))
	copy(aMatrix, r.polyhedron.A())

	bVector := make([]int, len(r.polyhedron.B()))
	copy(bVector, r.polyhedron.B())

	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)

	variableIDs := make([]string, len(r.variables))
	copy(variableIDs, r.variables)

	primitiveVariables := make([]string, len(r.primitiveVariables))
	copy(primitiveVariables, r.primitiveVariables)

	preferredIDs := make([]string, len(r.preferredVariables))
	copy(preferredIDs, r.preferredVariables)

	return &RuleSet{
		polyhedron:         polyhedron,
		primitiveVariables: primitiveVariables,
		variables:          variableIDs,
		preferredVariables: preferredIDs,
	}

}

func (r *RuleSet) newQuerySpecification(selections Selections) (*querySpecification, error) {
	ruleSet := r.copy()
	selectionIDLookUp := make(map[Selection]string)
	for _, selection := range selections {
		if selection.isComposite() {
			id, err := ruleSet.prepareCompositeSelection(selection.id, *selection.subSelectionID)
			if err != nil {
				return nil, err
			}

			selectionIDLookUp[selection] = id
		} else {
			selectionIDLookUp[selection] = selection.id
		}
	}

	return &querySpecification{
		ruleSet:           ruleSet,
		selectionIDLookUp: selectionIDLookUp,
	}, nil
}

func (r *RuleSet) prepareCompositeSelection(id, subSelectionID string) (string, error) {
	idIndex, err := utils.IndexOf(r.variables, id)
	if err != nil {
		return "", err
	}

	subSelectionIndex, err := utils.IndexOf(r.variables, subSelectionID)
	if err != nil {
		return "", err
	}

	constraint, err := pldag.NewAtLeastConstraint([]string{id, subSelectionID}, 2)
	if err != nil {
		return "", err
	}

	r.polyhedron.AddEmptyColumn()

	supportImpliesPrimitives := make([]int, len(r.variables)+1)
	primitivesImpliesSupport := make([]int, len(r.variables)+1)
	auxiliarySupportsImpliesPrimitive, auxiliaryPrimitivesImpliesSupport := constraint.ToAuxiliaryConstraintsWithSupport()

	supportImpliesPrimitives[idIndex] = auxiliarySupportsImpliesPrimitive.Coefficients()[id]
	supportImpliesPrimitives[subSelectionIndex] = auxiliarySupportsImpliesPrimitive.Coefficients()[subSelectionID]
	supportImpliesPrimitives[len(supportImpliesPrimitives)-1] = auxiliarySupportsImpliesPrimitive.Coefficients()[constraint.ID()]
	supportImpliesPrimitivesBias := auxiliarySupportsImpliesPrimitive.Bias()
	r.polyhedron.Extend(supportImpliesPrimitives, supportImpliesPrimitivesBias)

	primitivesImpliesSupport[idIndex] = auxiliaryPrimitivesImpliesSupport.Coefficients()[id]
	primitivesImpliesSupport[subSelectionIndex] = auxiliaryPrimitivesImpliesSupport.Coefficients()[subSelectionID]
	primitivesImpliesSupport[len(primitivesImpliesSupport)-1] = auxiliaryPrimitivesImpliesSupport.Coefficients()[constraint.ID()]
	primitivesImpliesSupportBias := auxiliaryPrimitivesImpliesSupport.Bias()
	r.polyhedron.Extend(primitivesImpliesSupport, primitivesImpliesSupportBias)

	return constraint.ID(), nil
}
