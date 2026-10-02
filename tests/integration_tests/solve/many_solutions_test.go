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
	query := puan.NewManySolutionsQuery(
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
		{},
	}
	query := puan.NewManySolutionsQuery(
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

	group3 := solutionForGroup[2]
	assert.Empty(t, group3.Selections())
	assert.Equal(
		t,
		puan.Solution{
			"optA": 1,
			"optB": 0,
			"optC": 0,
			"optD": 0,
		},
		group3.Solution(),
	)
}

func Test_CreateManySolutions_givenRemoveSelections(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC"}
	_ = creator.AddPrimitives(primitives...)

	aEquivB, _ := creator.SetEquivalent("optA", "optB")
	_ = creator.Assume(aEquivB)

	ruleset, err := creator.Create()
	assert.NoError(t, err)

	selectionGroups := []puan.Selections{
		{
			puan.NewSelectionBuilder("optA").Build(),
			puan.NewSelectionBuilder("optC").Build(),
		},
		{
			puan.NewSelectionBuilder("optA").Build(),
			puan.NewSelectionBuilder("optB").WithAction(puan.REMOVE).Build(),
		},
	}
	query := puan.NewManySolutionsQuery(
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
		puan.Solution{
			"optA": 1,
			"optB": 1,
			"optC": 1,
		},
		group1.Solution(),
	)

	group2 := solutionForGroup[1]
	assert.Equal(
		t,
		puan.Solution{
			"optA": 0,
			"optB": 0,
			"optC": 0,
		},
		group2.Solution(),
	)
}

// Test_optionalVariant_changeVariant and
// Test_optionalVariant_selectItemInAnotherVariant_shouldChangeVariant
// at the same time
func Test_CreateManySolutions_givenSubSelections(
	t *testing.T,
) {
	ruleset := optionalVariantsWithXORBetweenItemsLargeVariantPreferred()

	selections1 := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
	}
	selections2 := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
		puan.NewSelectionBuilder("itemY").Build(),
	}
	selectionGroups := []puan.Selections{
		selections1,
		selections2,
	}

	query := puan.NewManySolutionsQuery(
		selectionGroups,
		ruleset,
		nil,
		nil,
	)
	envelope, _ := solutionCreator.CreateManySolutions(query)
	solutionForGroup := envelope.SolutionsBySelectionGroup()

	assert.Len(t, solutionForGroup, 2)

	group1 := solutionForGroup[0]
	assert.Equal(
		t,
		selections1,
		group1.Selections(),
	)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    1,
			"itemY":    0,
			"itemZ":    0,
		},
		group1.Solution(),
	)

	group2 := solutionForGroup[1]
	assert.Equal(
		t,
		selections2,
		group2.Selections(),
	)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    1,
		},
		group2.Solution(),
	)
}
