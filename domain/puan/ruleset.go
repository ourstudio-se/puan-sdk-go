package puan

import (
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

func (c RuleSetCreator) SetPreferred(id ...string) {
	c.preferredVariables = append(c.preferredVariables, id...)
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

func (r *RuleSet) NewQuery(selections Selections) (*Query, error) {
	selectedIDs, err := r.CalculateSelectedIDs(selections)
	if err != nil {
		return nil, err
	}

	objective := CalculateObjective(r.primitiveVariables, selectedIDs, r.preferredVariables)

	return NewQuery(r.polyhedron, r.variables, objective), nil
}

func (r *RuleSet) CalculateSelectedIDs(selections Selections) ([]string, error) {
	impactingSelections := selections.getImpactingSelections()
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

	r.polyhedron.IncrementMatrixRows()

	supportImpliesPrimitives := make([]int, len(r.variables)+1)
	primitivesImpliesSupport := make([]int, len(r.variables)+1)
	auxiliarySupportsImpliesPrimitive, auxiliaryPrimitivesImpliesSupport := constraint.ToAuxiliaryConstraintsWithSupport()

	supportImpliesPrimitives[idIndex] = auxiliarySupportsImpliesPrimitive.Coefficients()[id]
	supportImpliesPrimitives[subSelectionIndex] = auxiliarySupportsImpliesPrimitive.Coefficients()[subSelectionID]
	supportImpliesPrimitives[len(supportImpliesPrimitives)-1] = auxiliarySupportsImpliesPrimitive.Coefficients()[constraint.ID()]
	supportImpliesPrimitivesBias := auxiliarySupportsImpliesPrimitive.Bias()
	r.polyhedron.Append(supportImpliesPrimitives, supportImpliesPrimitivesBias)

	primitivesImpliesSupport[idIndex] = auxiliaryPrimitivesImpliesSupport.Coefficients()[id]
	primitivesImpliesSupport[subSelectionIndex] = auxiliaryPrimitivesImpliesSupport.Coefficients()[subSelectionID]
	primitivesImpliesSupport[len(primitivesImpliesSupport)-1] = auxiliaryPrimitivesImpliesSupport.Coefficients()[constraint.ID()]
	primitivesImpliesSupportBias := auxiliaryPrimitivesImpliesSupport.Bias()
	r.polyhedron.Append(primitivesImpliesSupport, primitivesImpliesSupportBias)

	return constraint.ID(), nil
}
