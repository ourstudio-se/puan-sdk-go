package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
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

func Test_Reverse(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	actual := Reverse(slice)
	expected := []int{5, 4, 3, 2, 1}

	assert.Equal(t, expected, actual)
}

func Test_Contains_shouldReturnTrue(t *testing.T) {
	slice := []string{"a", "b", "c"}
	actual := Contains(slice, "b")

	assert.True(t, actual)
}

func Test_Contains_shouldReturnFalse(t *testing.T) {
	slice := []string{"a", "b", "c"}
	actual := Contains(slice, "k")

	assert.False(t, actual)
}

func Test_IndexOf_givenElementExists(t *testing.T) {
	slice := []int{1, 2, 3}
	index, err := IndexOf(slice, 2)

	assert.NoError(t, err)
	assert.Equal(t, 1, index)
}

func Test_IndexOf_givenElementNotExists(t *testing.T) {
	slice := []int{1, 2, 3}
	_, err := IndexOf(slice, 4)

	assert.Error(t, err)
}

func Test_ContainsAny_givenSharedValues(t *testing.T) {
	sliceA := []string{"a", "b", "c"}
	sliceB := []string{"b", "c", "d"}
	actual := ContainsAny(sliceA, sliceB)
	assert.True(t, actual)
}

func Test_ContainsAny_givenNoSharedValues(t *testing.T) {
	sliceA := []string{"a", "b", "c"}
	sliceB := []string{"d", "e", "g"}
	actual := ContainsAny(sliceA, sliceB)
	assert.False(t, actual)
}

func Test_ContainsAll_givenAllValues(t *testing.T) {
	sliceA := []string{"a", "b", "c", "d"}
	sliceB := []string{"a", "b", "c"}
	actual := ContainsAll(sliceA, sliceB)
	assert.True(t, actual)
}

func Test_ContainsAll_givenMissingValues(t *testing.T) {
	sliceA := []string{"a", "b", "c"}
	sliceB := []string{"d", "e", "g"}
	actual := ContainsAll(sliceA, sliceB)
	assert.False(t, actual)
}

func Test_Dedupe_givenDuplicates(t *testing.T) {
	slice := []string{"a", "b", "c", "a"}
	actual := Dedupe(slice)
	expected := []string{"a", "b", "c"}

	assert.Equal(t, expected, actual)
}

func Test_Dedupe_givenNoDuplicates(t *testing.T) {
	slice := []string{"a", "b", "c"}
	actual := Dedupe(slice)
	expected := []string{"a", "b", "c"}

	assert.Equal(t, expected, actual)
}

func Test_Dedupe_givenEmptySlice(t *testing.T) {
	slice := []string{}
	actual := Dedupe(slice)
	var expected []string

	assert.Equal(t, expected, actual)
}

func Test_Sorted_givenUnsortedIntSlice(t *testing.T) {
	slice := []int{5, 3, 4, 1, 2}
	actual := Sorted(slice)
	expected := []int{1, 2, 3, 4, 5}

	assert.Equal(t, expected, actual)
}

func Test_Sorted_givenUnsortedStringSlice(t *testing.T) {
	slice := []string{"e", "c", "d", "a", "b"}
	actual := Sorted(slice)
	expected := []string{"a", "b", "c", "d", "e"}

	assert.Equal(t, expected, actual)
}

func Test_SortedBy(t *testing.T) {
	m1 := fake.New[map[string]string]()
	m2 := fake.New[map[string]string]()
	s1 := fake.New[[]string]()
	s2 := fake.New[[]string]()

	type obj struct {
		id string
		m  map[string]string
		s  []string
	}

	in := []obj{
		{
			id: "c",
			m:  m1,
			s:  s1,
		},
		{
			id: "a",
			m:  m2,
			s:  s2,
		},
	}

	want := []obj{
		{
			id: "a",
			m:  m2,
			s:  s2,
		},
		{
			id: "c",
			m:  m1,
			s:  s1,
		},
	}

	got := SortedBy(in, func(o obj) string { return o.id })
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("SortedBy() = %v, want %v", got, want)
	}
}
