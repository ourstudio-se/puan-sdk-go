package puan

import (
	"math/rand"
	"testing"
	"time"

	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
)

func Test_RulesetCreator_newTimeBoundVariable_givenTimeEnabled_andValidPeriod(t *testing.T) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRulesetCreator()
	err := creator.EnableTime(from, to)
	assert.NoError(t, err)

	variable, err := creator.newTimeBoundVariable(
		id,
		from,
		to,
	)

	assert.NoError(t, err)
	assert.Equal(t, id, variable.variable)
	assert.Equal(t, from, variable.period.From())
	assert.Equal(t, to, variable.period.To())
}

func Test_RulesetCreator_newTimeBoundVariable_givenTimeNotEnabled_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRulesetCreator()
	_, err := creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}

func Test_RulesetCreator_newTimeBoundVariable_givenTimeEnabled_andInvalidPeriod_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-31T00:00:00Z")
	to := newTestTime("2024-01-01T00:00:00Z")

	creator := NewRulesetCreator()
	err := creator.EnableTime(
		newTestTime("2024-01-01T00:00:00Z"),
		newTestTime("2024-02-01T00:00:00Z"),
	)
	assert.NoError(t, err)

	_, err = creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}

// nolint:lll
func Test_RulesetCreator_newTimeBoundVariable_givenTimeEnabled_andAssumedPeriodOutsideOfEnabledPeriod_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRulesetCreator()
	err := creator.EnableTime(
		from.Add(1*24*time.Hour),
		to.Add(-1*24*time.Hour),
	)
	assert.NoError(t, err)

	_, err = creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}

// Test_Create_givenDifferentModelingOrder_shouldReturnSamePolyhedron
// This test ensures that the order in which primitives and rules
// are added does not affect the resulting polyhedron
func Test_Create_givenDifferentModelingOrder_shouldReturnSamePolyhedron(
	t *testing.T,
) {
	primitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 10
			oo.RandomMaxSliceSize = 15
		},
	)

	creatorOne := NewRulesetCreator()
	_ = creatorOne.AddPrimitives(primitives...)

	id1One, _ := creatorOne.SetImply(primitives[0], primitives[1])
	id2One, _ := creatorOne.SetOr(primitives[2], primitives[3])
	id3One, _ := creatorOne.SetAnd(primitives[4], primitives[5])
	id4One, _ := creatorOne.SetXor(primitives[6], primitives[7], primitives[8])
	id5One, _ := creatorOne.SetImply(id3One, id2One)
	id6One, _ := creatorOne.SetNot(primitives[9])

	rootOne, _ := creatorOne.SetAnd(id1One, id2One, id3One, id4One, id5One, id6One)
	_ = creatorOne.Assume(rootOne)
	rulesetOne, _ := creatorOne.Create()

	shuffledPrimitives := append([]string(nil), primitives...)
	rand.Shuffle(len(shuffledPrimitives), func(i, j int) {
		shuffledPrimitives[i], shuffledPrimitives[j] = shuffledPrimitives[j], shuffledPrimitives[i]
	})

	// Create a second creator with the same primitives
	// and rules but in a different order.
	creatorTwo := NewRulesetCreator()
	_ = creatorTwo.AddPrimitives(shuffledPrimitives...)
	id1Two, _ := creatorTwo.SetAnd(primitives[4], primitives[5])
	id2Two, _ := creatorTwo.SetImply(primitives[0], primitives[1])
	id3Two, _ := creatorTwo.SetOr(primitives[2], primitives[3])
	id4Two, _ := creatorTwo.SetNot(primitives[9])
	id5Two, _ := creatorTwo.SetXor(primitives[6], primitives[7], primitives[8])
	id6Two, _ := creatorTwo.SetImply(id1Two, id3Two)

	rootTwo, _ := creatorTwo.SetAnd(id1Two, id2Two, id3Two, id4Two, id5Two, id6Two)
	_ = creatorTwo.Assume(rootTwo)
	rulesetTwo, _ := creatorTwo.Create()

	assert.Equalf(
		t,
		rulesetOne.dependentVariables,
		rulesetTwo.dependentVariables,
		"dependentVariables are not equal",
	)
	assert.Equalf(t, rulesetOne.polyhedron.A(), rulesetTwo.polyhedron.A(), "A matrices are not equal")
	assert.Equalf(t, rulesetOne.polyhedron.B(), rulesetTwo.polyhedron.B(), "B vectors are not equal")
}

func Test_RulesetCreator_AssumeInPeriod_givenSamePeriod_shouldUseAssume(t *testing.T) {
	creator := NewRulesetCreator()
	from, to := newTestTime("2024-01-01T00:00:00Z"), newTestTime("2024-01-31T23:59:59Z")
	err := creator.EnableTime(
		from,
		to,
	)
	assert.NoError(t, err)

	_ = creator.AddPrimitives("itemX")
	err = creator.AssumeInPeriod("itemX", from, to)
	assert.NoError(t, err)

	assert.Contains(t, creator.assumedVariables, "itemX")
	assert.NotContains(t, creator.timeBoundAssumedVariables.ids(), "itemX")
}

// nolint:lll
func Test_RulesetCreator_AssumeInPeriod_givenDifferentPeriod_shouldAddTimeBoundVariable(t *testing.T) {
	creator := NewRulesetCreator()
	from, to := newTestTime("2024-01-01T00:00:00Z"), newTestTime("2024-01-31T23:59:59Z")
	err := creator.EnableTime(
		from,
		to,
	)
	assert.NoError(t, err)

	newFrom := from.Add(time.Hour)
	_ = creator.AddPrimitives("itemX")
	err = creator.AssumeInPeriod("itemX", newFrom, to)
	assert.NoError(t, err)

	assert.NotContains(t, creator.assumedVariables, "itemX")
	assert.Contains(t, creator.timeBoundAssumedVariables.ids(), "itemX")
}
