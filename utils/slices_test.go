package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/fake"
)

func Test_Without(t *testing.T) {
	sliceA := []int{1, 2, 3, 4, 5}
	sliceB := []int{2, 4}
	want := []int{1, 3, 5}

	actual := Without(sliceA, sliceB)

	assert.Equal(t, want, actual)
}

func Test_Without_nilComparable_shouldReturnSameSlice(t *testing.T) {
	sliceA := fake.New[[]int]()
	actual := Without(sliceA, nil)

	assert.Equal(t, sliceA, actual)
}

func Test_Without_nilInput_shouldReturnEmptySlice(t *testing.T) {
	sliceB := fake.New[[]int]()
	actual := Without(nil, sliceB)
	var want []int

	assert.Equal(t, want, actual)
}

func Test_ContainsDuplicates_shouldReturnTrue(t *testing.T) {
	slice := []string{"a", "b", "c", "a"}
	actual := ContainsDuplicates(slice)

	assert.Equal(t, true, actual)
}

func Test_ContainsDuplicates_shouldReturnFalse(t *testing.T) {
	slice := []string{"a", "b", "c"}
	actual := ContainsDuplicates(slice)

	assert.Equal(t, false, actual)
}
