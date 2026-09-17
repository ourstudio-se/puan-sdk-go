//nolint:lll
package solve

import (
	"testing"

	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateNextSolutions_givenNextRemoveSelection(
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
	query, _ := puan.NewNextSolutionsQuery(
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

func Test_CreateNextSolutions_givenNextAddSelection(
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
	query, _ := puan.NewNextSolutionsQuery(
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

func Test_CreateNextSolutions_shouldCreateSolutionsForAllNextSelections(
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
	query, _ := puan.NewNextSolutionsQuery(
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

func Test_CreateNextSolutions_givenCompositeNextSelection(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()
	_ = creator.AddPrimitives("packageA", "itemX", "itemY", "itemZ")

	xorItem1Item2, _ := creator.SetXor("itemX", "itemY")
	xorItem1Item3, _ := creator.SetXor("itemX", "itemZ")

	packageExactlyOneOfItem1Item2, _ := creator.SetImply("packageA", xorItem1Item2)
	packageExactlyOneOfItem1Item3, _ := creator.SetImply("packageA", xorItem1Item3)

	_ = creator.Assume(
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
	)

	ruleset, _ := creator.Create()

	addPackageAVariant1 := puan.NewSelectionBuilder("packageA").
		WithSubSelectionID("itemY").
		Build()
	currentSelections := puan.Selections{
		addPackageAVariant1,
	}

	addPackageAVariant2 := puan.NewSelectionBuilder("packageA").
		WithSubSelectionID("itemX").
		Build()
	nextSelections := puan.Selections{
		addPackageAVariant2,
	}

	query, _ := puan.NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	envelope, _ := solutionCreator.CreateNextSolutions(query)
	solutionForSelection, err := envelope.GetSolutionBySelection(addPackageAVariant2)
	require.NoError(t, err)

	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    1,
			"itemY":    0,
			"itemZ":    0,
		},
		solutionForSelection.Solution(),
	)
}

// Every weight group has 70 current selections plus one next selection, so all of them saturate
// and the batched solve has to fall back to solving each group on its own.
func Test_CreateNextSolutions_givenSaturatedWeights_shouldCreateSolutionForEach(
	t *testing.T,
) {
	creator, primitives := setupSaturatableRuleset(t)

	_ = creator.AddPrimitives("a", "b", "c")
	aImpliesB, _ := creator.SetImply("a", "b")
	aXorC, _ := creator.SetXor("a", "c")
	_ = creator.Assume(aImpliesB, aXorC)

	ruleset, _ := creator.Create()

	currentSelections := puan.Selections{puan.NewSelectionBuilder("a").Build()}
	for _, primitive := range primitives {
		currentSelections = append(
			currentSelections,
			puan.NewSelectionBuilder(primitive).Build(),
		)
	}

	removeA := puan.NewSelectionBuilder("a").WithAction(puan.REMOVE).Build()
	addB := puan.NewSelectionBuilder("b").Build()
	nextSelections := puan.Selections{removeA, addB}
	query, _ := puan.NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	envelope, err := solutionCreator.CreateNextSolutions(query)
	require.NoError(t, err)

	// The next selection outranks everything in current selection, so removing "a" wins.
	solutionWithoutA, err := envelope.GetSolutionBySelection(removeA)
	require.NoError(t, err)
	asserter := newSolutionAsserter(solutionWithoutA.Solution())
	asserter.assertInactive(t, "a", "b")
	asserter.assertActive(t, "c")

	// "b" conflicts with nothing, so the current "a" survives.
	solutionWithB, err := envelope.GetSolutionBySelection(addB)
	require.NoError(t, err)
	asserter = newSolutionAsserter(solutionWithB.Solution())
	asserter.assertActive(t, "a", "b")
	asserter.assertInactive(t, "c")
}

func Test_CreateNextSolutions_givenSaturatedWeightsAndIndependentNextSelection(
	t *testing.T,
) {
	creator, primitives := setupSaturatableRuleset(t)

	_ = creator.AddPrimitives("a", "b", "independent")
	aXorC, _ := creator.SetXor("a", "b")
	_ = creator.Assume(aXorC)

	ruleset, _ := creator.Create()

	currentSelections := puan.Selections{puan.NewSelectionBuilder("a").Build()}
	for _, primitive := range primitives {
		currentSelections = append(
			currentSelections,
			puan.NewSelectionBuilder(primitive).Build(),
		)
	}

	removeA := puan.NewSelectionBuilder("a").WithAction(puan.REMOVE).Build()
	addIndependent := puan.NewSelectionBuilder("independent").Build()
	nextSelections := puan.Selections{removeA, addIndependent}

	query, _ := puan.NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	envelope, err := solutionCreator.CreateNextSolutions(query)
	require.NoError(t, err)

	// The next selection outranks everything in current selection, so removing "a" wins.
	solutionWithoutA, err := envelope.GetSolutionBySelection(removeA)
	require.NoError(t, err)
	asserter := newSolutionAsserter(solutionWithoutA.Solution())
	asserter.assertInactive(t, "a", "independent")
	asserter.assertActive(t, "b")

	// "independent" conflicts with nothing, so the current "a" survives.
	solutionWithIndependent, err := envelope.GetSolutionBySelection(addIndependent)
	require.NoError(t, err)
	asserter = newSolutionAsserter(solutionWithIndependent.Solution())
	asserter.assertActive(t, "a", "independent")
	asserter.assertInactive(t, "b")
}

func setupSaturatableRuleset(t *testing.T) (*puan.RulesetCreator, []string) {
	t.Helper()
	creator := puan.NewRulesetCreator()
	primitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 70
			oo.RandomMaxSliceSize = 70
		},
	)
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr(primitives...)
	_ = creator.Assume(orID)

	return creator, primitives
}
