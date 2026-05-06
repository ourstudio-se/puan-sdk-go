package solve

import (
	"testing"
	"time"

	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// TODO: adjust test when weights to large is removed
func Test_WeightsTooLarge_givenSelectionsAtSaturation_weightsShouldBeTooLarge(t *testing.T) {
	ruleset, primitives := rulesetWithPrimitivesForSaturationTests()

	selections := puan.Selections{}
	for _, primitive := range primitives[:70] {
		selections = append(
			selections,
			puan.NewSelectionBuilder(primitive).Build(),
		)
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	weightsTooLarge := envelope.WeightsTooLarge()
	assert.False(
		t,
		weightsTooLarge,
	)
}

func rulesetWithPrimitivesForSaturationTests() (puan.Ruleset, []string) {
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

	ruleset, _ := creator.Create()

	return ruleset, primitives
}
