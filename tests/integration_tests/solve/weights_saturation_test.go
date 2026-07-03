package solve

import (
	"testing"
	"time"

	"github.com/go-faker/faker/v4/pkg/options"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Ruleset with a lot of primitives (103). All of them are dependent.
// We make 71 selections.
// The first selection (least prioritised) should be selected as none
// of the other selections are in conflict with it.
func Test_givenVeryManySelections_earliestSelectionShouldBeSelected(t *testing.T) {
	creator, primitives := rulesetCreatorForSaturationTest()

	_ = creator.AddPrimitives("a", "b", "c")
	aImpliesB, _ := creator.SetImply("a", "b")
	aXorC, _ := creator.SetXor("a", "c")
	_ = creator.Assume(aImpliesB, aXorC)

	ruleset, _ := creator.Create()

	selections := puan.Selections{}
	selections = append(
		selections,
		puan.NewSelectionBuilder("a").Build(),
	)
	for _, primitive := range primitives[:70] {
		selections = append(
			selections,
			puan.NewSelectionBuilder(primitive).Build(),
		)
	}

	envelope, _ := solutionCreator.Create(puan.SolutionQuery{Selections: selections, Ruleset: ruleset})

	asserter := newSolutionAsserter(envelope.Solution())
	asserter.assertActive(
		t,
		"a",
		"b",
	)
	asserter.assertInactive(
		t,
		"c",
	)
}

func rulesetCreatorForSaturationTest() (*puan.RulesetCreator, []string) {
	creator := puan.NewRulesetCreator()
	from := time.Now()
	end := from.Add(1 * time.Hour)
	_ = creator.EnableTime(from, end)

	primitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 100
			oo.RandomMaxSliceSize = 100
		},
	)
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr(primitives...)
	_ = creator.Assume(orID)

	// Create 10 assumes for different primitives in different periods
	assumeFrom := from
	for i := 0; i < 10; i++ {
		assumeEnd := assumeFrom.Add(5 * time.Minute)
		_ = creator.AssumeInPeriod(primitives[i], assumeFrom, assumeEnd)
		assumeFrom = assumeEnd
	}

	// Create 10 preferreds for different primitives in different periods
	preferFrom := from
	for i := 10; i < 20; i++ {
		preferEnd := preferFrom.Add(5 * time.Minute)
		preferredID, _ := creator.SetImply(orID, primitives[i])
		_ = creator.PreferInPeriod(preferredID, preferFrom, preferEnd)
		preferFrom = preferEnd
	}

	return creator, primitives
}
