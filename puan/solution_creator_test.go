package puan

import (
	"fmt"
	"testing"

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
	ruleset, primitives := setupSaturatableRuleset(t)
	client := &fakeSolverClient{}
	creator := NewSolutionCreator(client)

	currentSelections := selectionsFor(t, primitives[:3])
	nextSelections := selectionsFor(t, primitives[3:5])
	query, err := NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	require.NoError(t, err)

	actual, err := creator.calculateNextDependentSolutions(query)

	require.NoError(t, err)
	assertSolutionExistsForEachSelection(t, actual, nextSelections)
	assert.Equal(t, 1, client.multiSolveCalls)
	assert.Zero(t, client.solveCalls)
}

func Test_SolutionCreator_calculateNextDependentSolutions_givenOversizedWeights_shouldSolveOneByOne(
	t *testing.T,
) {
	ruleset, primitives := setupSaturatableRuleset(t)
	client := &fakeSolverClient{}
	creator := NewSolutionCreator(client)

	// Select many for current to enforce saturation.
	currentSelections := selectionsFor(t, primitives[:25])
	nextSelections := selectionsFor(t, primitives[25:27])
	query, err := NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	require.NoError(t, err)

	actual, err := creator.calculateNextDependentSolutions(query)

	require.NoError(t, err)
	assertSolutionExistsForEachSelection(t, actual, nextSelections)
	assert.Zero(t, client.multiSolveCalls)
	// each selection is split once, 2 calls per next selection, therefore 4 in total.
	assert.Equal(t, 4, client.solveCalls)
}

func Test_SolutionCreator_CreateNextSolutions_givenNoNextSelections_shouldReturnEmptyEnvelope(
	t *testing.T,
) {
	ruleset, primitives := setupSaturatableRuleset(t)
	client := &fakeSolverClient{}
	creator := NewSolutionCreator(client)

	query, err := NewNextSolutionsQuery(
		selectionsFor(t, primitives[:3]),
		nil,
		ruleset,
		nil,
		nil,
	)
	require.NoError(t, err)

	envelope, err := creator.CreateNextSolutions(query)

	require.NoError(t, err)
	assert.Empty(t, envelope.SolutionsBySelection())
}

func selectionsFor(t *testing.T, primitives []string) Selections {
	t.Helper()
	selections := make(Selections, len(primitives))
	for i, primitive := range primitives {
		selections[i] = NewSelectionBuilder(primitive).Build()
	}

	return selections
}

func assertSolutionExistsForEachSelection(
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

func setupSaturatableRuleset(t *testing.T) (Ruleset, []string) {
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

type fakeSolverClient struct {
	solveCalls      int
	multiSolveCalls int
}

func (c *fakeSolverClient) Solve(_ *SolverQuery) (Solution, error) {
	c.solveCalls++

	return Solution{}, nil
}

func (c *fakeSolverClient) SolveWithManyWeights(
	query *MultiWeightSolverQuery,
) ([]Solution, error) {
	c.multiSolveCalls++

	solutions := make([]Solution, len(query.WeightGroups()))
	for i := range solutions {
		solutions[i] = Solution{}
	}

	return solutions, nil
}
