package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Weights_Concat(t *testing.T) {
	w1 := Weights{
		"x": 1,
		"y": 2,
	}

	w2 := Weights{
		"z": 3,
	}

	want := Weights{
		"x": 1,
		"y": 2,
		"z": 3,
	}

	actual := w1.Concat(w2)
	assert.Equal(t, want, actual)
}
