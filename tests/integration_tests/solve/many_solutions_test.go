package solve

import (
	"testing"
	"time"

	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
)

// Ruleset with many dependent primitives.
// Create selections for all, some of which are composite.
// The solver should create a solution for each selection.
// nolint:lll
func Test_CreateSolutionsBySelection_givenManyDependentSelections_shouldCreateSolutionForEach(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()
	from := time.Now()
	end := from.Add(1 * time.Hour)
	_ = creator.EnableTime(from, end)

	primitivesCount := 80
	primitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = primitivesCount
			oo.RandomMaxSliceSize = primitivesCount
		},
	)
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr(primitives...)
	_ = creator.Assume(orID)

	ruleset, _ := creator.Create()

	// Make half of the selections composite.
	selections := make([]puan.Selection, len(primitives))
	for i, primitive := range primitives {
		builder := puan.NewSelectionBuilder(primitive)
		if i < (primitivesCount / 2) {
			otherPrimitive := primitives[i*2]
			builder.WithSubSelectionID(otherPrimitive)
		}
		selections[i] = builder.Build()
	}

	query := puan.NewSolutionQueryBuilder().WithSelections(selections).WithRuleset(ruleset).Build()
	solutions, _ := solutionCreator.CreateSolutionsBySelection(query)

	assert.Len(t, solutions.SolutionsBySelection(), len(primitives))
	for _, selection := range selections {
		solution, err := solutions.GetSolutionBySelection(selection)

		assert.NoError(t, err)
		newSolutionAsserter(solution.Solution()).
			assertActive(t, selection.IDs()...)
	}
}

// Ruleset with many independent primitives.
// The solver should create a solution for each selection.
// nolint:lll
func Test_CreateSolutionsBySelection_givenManyIndependentSelections_shouldCreateSolutionForEach(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()
	from := time.Now()
	end := from.Add(1 * time.Hour)
	_ = creator.EnableTime(from, end)

	primitivesCount := 80
	primitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = primitivesCount
			oo.RandomMaxSliceSize = primitivesCount
		},
	)
	_ = creator.AddPrimitives(primitives...)

	ruleset, _ := creator.Create()

	selections := make([]puan.Selection, len(primitives))
	for i, primitive := range primitives {
		selections[i] = puan.NewSelectionBuilder(primitive).Build()
	}

	query := puan.NewSolutionQueryBuilder().WithSelections(selections).WithRuleset(ruleset).Build()
	solutions, _ := solutionCreator.CreateSolutionsBySelection(query)

	assert.Len(t, solutions.SolutionsBySelection(), len(primitives))
	for _, selection := range selections {
		solution, err := solutions.GetSolutionBySelection(selection)

		assert.NoError(t, err)
		newSolutionAsserter(solution.Solution()).
			assertActive(t, selection.IDs()...)
	}
}

// Ruleset with many dependent and independent primitives.
// The solver should create a solution for each selection.
// nolint:lll
func Test_CreateSolutionsBySelection_givenMixedSelections_shouldCreateSolutionForEach(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()
	from := time.Now()
	end := from.Add(1 * time.Hour)
	_ = creator.EnableTime(from, end)

	dependentPrimitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 40
			oo.RandomMaxSliceSize = 40
		},
	)
	independentPrimitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 40
			oo.RandomMaxSliceSize = 40
		},
	)

	var primitives []string
	primitives = append(primitives, dependentPrimitives...)
	primitives = append(primitives, independentPrimitives...)

	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr(dependentPrimitives...)
	_ = creator.Assume(orID)

	selections := make([]puan.Selection, len(primitives))
	for i, primitive := range primitives {
		selections[i] = puan.NewSelectionBuilder(primitive).Build()
	}

	ruleset, _ := creator.Create()

	query := puan.NewSolutionQueryBuilder().WithSelections(selections).WithRuleset(ruleset).Build()
	solutions, _ := solutionCreator.CreateSolutionsBySelection(query)

	assert.Len(t, solutions.SolutionsBySelection(), len(primitives))
	for _, selection := range selections {
		solution, err := solutions.GetSolutionBySelection(selection)

		assert.NoError(t, err)
		newSolutionAsserter(solution.Solution()).
			assertActive(t, selection.IDs()...)
	}
}
