package puan

import (
	"fmt"
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SolutionCreator_groupSolutionsBySelection(t *testing.T) {
	creator := &SolutionCreator{}

	selection1 := NewSelectionBuilder("x").Build()
	selection2 := NewSelectionBuilder("y").WithAction(REMOVE).Build()
	solution1 := Solution{"x": 1, "y": 0}
	solution2 := Solution{"x": 0, "y": 1}

	got, err := creator.groupSolutionsBySelection(
		[]Solution{solution1, solution2},
		Selections{selection1, selection2},
	)

	assert.NoError(t, err)
	assert.Equal(t, []SolutionBySelection{
		{selection: selection1, solution: solution1},
		{selection: selection2, solution: solution2},
	}, got)
}

func Test_SolutionCreator_groupSolutionsBySelection_givenLengthMismatch_shouldReturnError(
	t *testing.T,
) {
	creator := &SolutionCreator{}

	got, err := creator.groupSolutionsBySelection(
		[]Solution{{"x": 1}},
		Selections{
			NewSelectionBuilder("x").Build(),
			NewSelectionBuilder("y").Build(),
		},
	)

	assert.Nil(t, got)
	assert.Error(t, err)
}

func Test_SolutionCreator_calculateNextDependentSolutions_givenSmallWeights_shouldSolveInOneBatch(
	t *testing.T,
) {
	ruleset, primitives := rulesetWithDependentPrimitives(t)
	client := &fakeSolverClient{}
	creator := NewSolutionCreator(client)

	currenSelections := selectionsFor(t, primitives[:3])
	nextSelections := selectionsFor(t, primitives[40:42])
	query, err := NewNextSolutionsQuery(
		currenSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	require.NoError(t, err)

	actual, err := creator.calculateNextDependentSolutions(query)

	require.NoError(t, err)
	assertSolutionForEachSelection(t, actual, nextSelections)
	assert.Equal(t, len(nextSelections), client.batchedGroupCount)
	assert.Zero(t, client.solveCalls)
}

func Test_SolutionCreator_calculateNextDependentSolutions_givenTooLargeWeights_shouldSolveOneByOne(
	t *testing.T,
) {
	ruleset, primitives := rulesetWithDependentPrimitives(t)
	client := &fakeSolverClient{}
	creator := NewSolutionCreator(client)

	currenSelections := selectionsFor(t, primitives[:60])
	nextSelections := selectionsFor(t, primitives[60:62])
	query, err := NewNextSolutionsQuery(
		currenSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	require.NoError(t, err)

	actual, err := creator.calculateNextDependentSolutions(query)

	require.NoError(t, err)
	assertSolutionForEachSelection(t, actual, nextSelections)
	assert.Empty(t, client.batchedGroupCount)
	assert.Positive(t, client.solveCalls)
}

type fakeSolverClient struct {
	solveCalls        int
	batchedGroupCount int
}

func (c *fakeSolverClient) Solve(_ *SolverQuery) (Solution, error) {
	c.solveCalls++

	return Solution{}, nil
}

func (c *fakeSolverClient) SolveWithManyWeights(
	query *MultiWeightSolverQuery,
) ([]Solution, error) {
	c.batchedGroupCount = len(query.WeightGroups())

	solutions := make([]Solution, len(query.WeightGroups()))
	for i := range solutions {
		solutions[i] = Solution{}
	}

	return solutions, nil
}

// All primitives are pulled into one assumed OR so that they end up dependent.
func rulesetWithDependentPrimitives(t *testing.T) (Ruleset, []string) {
	t.Helper()

	primitives := make([]string, 100)
	for i := range primitives {
		primitives[i] = fmt.Sprintf("p%d", i)
	}

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primitives...)

	orID, err := creator.SetOr(primitives...)
	require.NoError(t, err)
	require.NoError(t, creator.Assume(orID))

	ruleset, err := creator.Create()
	require.NoError(t, err)

	return ruleset, primitives
}

func selectionsFor(t *testing.T, primitives []string) Selections {
	t.Helper()
	selections := make(Selections, len(primitives))
	for i, primitive := range primitives {
		selections[i] = NewSelectionBuilder(primitive).Build()
	}

	return selections
}

// The solutions come back batched first and saturated last, so they are matched
// by selection rather than by position.
func assertSolutionForEachSelection(
	t *testing.T,
	solutions []SolutionBySelection,
	selections Selections,
) {
	t.Helper()

	require.Len(t, solutions, len(selections))

	envelope, err := NewSolutionsBySelectionEnvelope(solutions)
	require.NoError(t, err)

	for _, selection := range selections {
		_, err := envelope.GetSolutionBySelection(selection)
		assert.NoError(t, err, "no solution for selection %s", selection.ID())
	}
}

func Test_newWeightedSelections(t *testing.T) {
	a := NewSelectionBuilder("a").Build()
	b := NewSelectionBuilder("b").Build()
	weightGroups := []weights.Weights{{"a": 1}, {"b": 2}}

	actual, err := newWeightedSelections(Selections{a, b}, weightGroups)

	require.NoError(t, err)
	assert.Equal(t, weightedSelections{
		{selection: a, weights: weights.Weights{"a": 1}},
		{selection: b, weights: weights.Weights{"b": 2}},
	}, actual)
}

func Test_newWeightedSelections_givenLengthMismatch_shouldReturnError(t *testing.T) {
	selections := Selections{NewSelectionBuilder("a").Build()}

	actual, err := newWeightedSelections(selections, []weights.Weights{{"a": 1}, {"b": 2}})

	assert.Nil(t, actual)
	assert.ErrorIs(t, err, puanerror.InvalidArgument)
}

func Test_weightedSelections_splitBySaturation(t *testing.T) {
	fits := weights.Weights{"x": 1}
	saturates := weights.Weights{"x": weights.WEIGHTS_SATURATION_LIMIT + 1}

	aFits := weightedSelection{selection: NewSelectionBuilder("a").Build(), weights: fits}
	bSaturate := weightedSelection{selection: NewSelectionBuilder("b").Build(), weights: saturates}
	cFits := weightedSelection{selection: NewSelectionBuilder("c").Build(), weights: fits}
	dSaturate := weightedSelection{selection: NewSelectionBuilder("d").Build(), weights: saturates}

	type theory struct {
		name              string
		weighted          weightedSelections
		expectedBatchable weightedSelections
		expectedSaturated weightedSelections
	}

	theories := []theory{
		{
			name:     "nothing to split",
			weighted: nil,
		},
		{
			name:              "all fit",
			weighted:          weightedSelections{aFits, cFits},
			expectedBatchable: weightedSelections{aFits, cFits},
			expectedSaturated: nil,
		},
		{
			name:              "all saturated",
			weighted:          weightedSelections{bSaturate, dSaturate},
			expectedSaturated: weightedSelections{bSaturate, dSaturate},
			expectedBatchable: nil,
		},
		{
			name:              "mixed keeps the relative order of each half",
			weighted:          weightedSelections{aFits, bSaturate, cFits, dSaturate},
			expectedBatchable: weightedSelections{aFits, cFits},
			expectedSaturated: weightedSelections{bSaturate, dSaturate},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			batchable, saturated := tt.weighted.splitBySaturation()

			assert.Equal(t, tt.expectedBatchable, batchable)
			assert.Equal(t, tt.expectedSaturated, saturated)
		})
	}
}

func Test_weightedSelections_selections(t *testing.T) {
	a := NewSelectionBuilder("a").Build()
	b := NewSelectionBuilder("b").Build()
	weighted := weightedSelections{
		{selection: a, weights: weights.Weights{"a": 1}},
		{selection: b, weights: weights.Weights{"b": 2}},
	}

	actual := weighted.selections()

	assert.Equal(t, Selections{a, b}, actual)
}

func Test_weightedSelections_weightGroups(t *testing.T) {
	weighted := weightedSelections{
		{selection: NewSelectionBuilder("a").Build(), weights: weights.Weights{"a": 1}},
		{selection: NewSelectionBuilder("b").Build(), weights: weights.Weights{"b": 2}},
	}

	actual := weighted.weightGroups()

	assert.Equal(t, []weights.Weights{{"a": 1}, {"b": 2}}, actual)
}
