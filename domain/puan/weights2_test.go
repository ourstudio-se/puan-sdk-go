package puan

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/stretchr/testify/assert"
)

func Test_weights2(t *testing.T) {
	model := pldag.New()
	primitives := make([]string, 60)
	for i := range primitives {
		primitives[i] = fmt.Sprintf("%s%s", faker.Word(), faker.Word())
	}

	model.SetPrimitives(primitives...)

	xors := make([]XORWithPreference2, 10)
	for i := range xors {
		xors[i] = XORWithPreference2{
			PreferredID: primitives[i],
			IDs:         []string{primitives[i], primitives[i+1], primitives[i+2]},
		}
	}

	objective := CalculateObjective2(primitives, primitives[:len(primitives)-10], xors)

	max := math.MinInt
	for _, v := range objective {
		if v > max {
			max = v
		}
	}

	maxInt32 := math.MaxInt64
	remaining := maxInt32 - max

	assert.True(t, remaining > 0)
	assert.Equal(t, 0, objective.sum())
}
