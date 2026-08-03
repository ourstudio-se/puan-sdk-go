//nolint:lll
package solve

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateNextSolutions_shouldRemoveExistingSelections(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC", "optD", "optE", "optF"}
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr("optA", "optB")

	_ = creator.Assume(orID)

	cImpliesD, _ := creator.SetImply("optC", "optD")

	_ = creator.Assume(cImpliesD)

	ruleset, err := creator.Create()
	assert.NoError(t, err)

	currentSelections := puan.Selections{
		puan.NewSelectionBuilder("optA").Build(),
		puan.NewSelectionBuilder("optE").Build(),
		puan.NewSelectionBuilder("optC").Build(),
	}
	removeA := puan.NewSelectionBuilder("optA").WithAction(puan.REMOVE).Build()
	nextSelections := puan.Selections{
		removeA,
	}
	query := puan.NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	envelope, err := solutionCreator.CreateNextSolutions(query)
	require.NoError(t, err)

	solution, err := envelope.GetSolutionBySelection(removeA)
	require.NoError(t, err)

	asserter := newSolutionAsserter(solution.Solution())
	asserter.assertInactive(t, "optA")
	asserter.assertActive(t, "optB")
	asserter.assertActive(t, "optC")
	asserter.assertActive(t, "optD")
	asserter.assertActive(t, "optE")
	asserter.assertInactive(t, "optF")
}

func Test_CreateNextSolutions_shouldAddNewSelections(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC", "optD", "optE", "optF"}
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr("optA", "optB")

	_ = creator.Assume(orID)

	cImpliesD, _ := creator.SetImply("optC", "optD")

	_ = creator.Assume(cImpliesD)

	ruleset, err := creator.Create()
	assert.NoError(t, err)

	currentSelections := puan.Selections{
		puan.NewSelectionBuilder("optA").Build(),
		puan.NewSelectionBuilder("optE").Build(),
		puan.NewSelectionBuilder("optC").Build(),
	}
	addB := puan.NewSelectionBuilder("optB").Build()
	nextSelections := puan.Selections{
		addB,
	}
	query := puan.NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	envelope, err := solutionCreator.CreateNextSolutions(query)
	require.NoError(t, err)

	solution, err := envelope.GetSolutionBySelection(addB)
	require.NoError(t, err)

	asserter := newSolutionAsserter(solution.Solution())
	asserter.assertActive(t, "optA")
	asserter.assertActive(t, "optB")
	asserter.assertActive(t, "optC")
	asserter.assertActive(t, "optD")
	asserter.assertActive(t, "optE")
	asserter.assertInactive(t, "optF")
}

func Test_CreateNextSolutions2_shouldCreateSolutionsForAllNextSelections(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC", "optD", "optE", "optF"}
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr("optA", "optB")

	_ = creator.Assume(orID)

	cImpliesD, _ := creator.SetImply("optC", "optD")

	_ = creator.Assume(cImpliesD)

	ruleset, _ := creator.Create()

	currentSelections := puan.Selections{
		puan.NewSelectionBuilder("optA").Build(),
		puan.NewSelectionBuilder("optE").Build(),
		puan.NewSelectionBuilder("optC").Build(),
	}
	nextSelections := puan.Selections{
		puan.NewSelectionBuilder("optA").WithAction(puan.REMOVE).Build(),
		puan.NewSelectionBuilder("optB").Build(),
		puan.NewSelectionBuilder("optC").WithAction(puan.REMOVE).Build(),
		puan.NewSelectionBuilder("optD").Build(),
		puan.NewSelectionBuilder("optE").WithAction(puan.REMOVE).Build(),
		puan.NewSelectionBuilder("optF").Build(),
	}
	query := puan.NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	envelope, _ := solutionCreator.CreateNextSolutions(query)

	assert.Len(t, envelope.SolutionsBySelection(), len(nextSelections))

	for _, nextSelection := range nextSelections {
		_, err := envelope.GetSolutionBySelection(nextSelection)
		require.NoError(t, err)
	}
}
