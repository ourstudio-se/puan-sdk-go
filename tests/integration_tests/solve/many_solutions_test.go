//nolint:lll
package solve

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateManySolutions_givenOnlyDependentSelections(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC", "optD", "optE", "optF"}
	_ = creator.AddPrimitives(primitives...)

	orABC, _ := creator.SetOr("optA", "optB", "optC")
	_ = creator.Assume(orABC)

	orDEF, _ := creator.SetOr("optD", "optE", "optF")
	_ = creator.Assume(orDEF)

	ruleset, err := creator.Create()
	assert.NoError(t, err)

	selectionGroups := []puan.Selections{
		{
			puan.NewSelectionBuilder("optA").Build(),
			puan.NewSelectionBuilder("optD").Build(),
		},
		{
			puan.NewSelectionBuilder("optB").Build(),
			puan.NewSelectionBuilder("optF").Build(),
		},
		{
			puan.NewSelectionBuilder("optC").Build(),
			puan.NewSelectionBuilder("optE").Build(),
		},
	}
	query := puan.NewManySolutionQueries(
		selectionGroups,
		ruleset,
		nil,
		nil,
	)
	envelope, err := solutionCreator.CreateManySolutions(query)
	require.NoError(t, err)

	solutionForGroup := envelope.SolutionsBySelectionGroup()

	assert.Len(t, solutionForGroup, 3)

	group1 := solutionForGroup[0]
	assert.Equal(
		t,
		selectionGroups[0],
		group1.Selections(),
	)
	assert.Equal(
		t,
		puan.Solution{
			"optA": 1,
			"optB": 0,
			"optC": 0,
			"optD": 1,
			"optE": 0,
			"optF": 0,
		},
		group1.Solution(),
	)

	group2 := solutionForGroup[1]
	assert.Equal(
		t,
		selectionGroups[1],
		group2.Selections(),
	)
	assert.Equal(
		t,
		puan.Solution{
			"optA": 0,
			"optB": 1,
			"optC": 0,
			"optD": 0,
			"optE": 0,
			"optF": 1,
		},
		group2.Solution(),
	)

	group3 := solutionForGroup[2]
	assert.Equal(
		t,
		selectionGroups[2],
		group3.Selections(),
	)
	assert.Equal(t,
		puan.Solution{
			"optA": 0,
			"optB": 0,
			"optC": 1,
			"optD": 0,
			"optE": 1,
			"optF": 0,
		},
		group3.Solution(),
	)
}

func Test_CreateManySolutions_givenDependentAndIndependentSelections(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC", "optD"}
	_ = creator.AddPrimitives(primitives...)

	orABC, _ := creator.SetOr("optA", "optB")
	_ = creator.Assume(orABC)

	ruleset, err := creator.Create()
	assert.NoError(t, err)

	selectionGroups := []puan.Selections{
		{
			puan.NewSelectionBuilder("optA").Build(),
			puan.NewSelectionBuilder("optC").Build(),
		},
		{
			puan.NewSelectionBuilder("optB").Build(),
			puan.NewSelectionBuilder("optD").Build(),
		},
	}
	query := puan.NewManySolutionQueries(
		selectionGroups,
		ruleset,
		nil,
		nil,
	)
	envelope, err := solutionCreator.CreateManySolutions(query)
	require.NoError(t, err)

	solutionForGroup := envelope.SolutionsBySelectionGroup()

	assert.Len(t, solutionForGroup, 2)

	group1 := solutionForGroup[0]
	assert.Equal(
		t,
		selectionGroups[0],
		group1.Selections(),
	)
	assert.Equal(
		t,
		puan.Solution{
			"optA": 1,
			"optB": 0,
			"optC": 1,
			"optD": 0,
		},
		group1.Solution(),
	)

	group2 := solutionForGroup[1]
	assert.Equal(
		t,
		selectionGroups[1],
		group2.Selections(),
	)
	assert.Equal(
		t,
		puan.Solution{
			"optA": 0,
			"optB": 1,
			"optC": 0,
			"optD": 1,
		},
		group2.Solution(),
	)
}
