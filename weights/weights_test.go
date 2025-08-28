package weights

import (
	"reflect"
	"testing"
)

func TestSelections_extractActiveSelections(t *testing.T) {
	tests := []struct {
		name       string
		selections Selections
		want       Selections
	}{
		{
			"with select and unselected selections",
			Selections{
				{
					ID:     "x",
					Action: ADD,
				},
				{
					ID:     "y",
					Action: ADD,
				},
				{
					ID:     "x",
					Action: REMOVE,
				},
			},
			Selections{
				{
					ID:     "y",
					Action: ADD,
				},
			},
		},
		{
			"no redundant selections",
			Selections{
				{
					ID:     "x",
					Action: ADD,
				},
				{
					ID:     "y",
					Action: ADD,
				},
				{
					ID:     "z",
					Action: ADD,
				},
			},
			Selections{
				{
					ID:     "x",
					Action: ADD,
				},
				{
					ID:     "y",
					Action: ADD,
				},
				{
					ID:     "z",
					Action: ADD,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.selections.extractActiveSelections(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractActiveSelections() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelections_extractSelectionsIDs(t *testing.T) {
	tests := []struct {
		name string
		s    Selections
		want []string
	}{
		{
			"empty selections",
			Selections{},
			[]string{},
		},
		{
			"one selection",
			Selections{{
				ID:     "x",
				Action: ADD,
			}},
			[]string{"x"},
		},
		{
			"two selections",
			Selections{
				{
					ID:     "x",
					Action: ADD,
				},
				{
					ID:     "y",
					Action: ADD,
				},
			},
			[]string{"x", "y"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.extractSelectionsIDs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractSelectionsIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name         string
		variables    []string
		selectedIDs  []string
		preferredIDs []string
		want         Weights
	}{
		{
			name:         "with preferred and selected IDs",
			variables:    []string{"a", "b", "c"},
			selectedIDs:  []string{"a"},
			preferredIDs: []string{"c"},
			want: Weights{
				"a": 5,
				"b": -1,
				"c": 3,
			},
		},
		{
			name:         "with preferred and no selected IDs",
			variables:    []string{"a", "b", "c"},
			selectedIDs:  []string{},
			preferredIDs: []string{"c"},
			want: Weights{
				"a": -1,
				"b": -1,
				"c": 3,
			},
		},
		{
			name:         "with preferred IDs and no selected IDs",
			variables:    []string{"a", "b", "c"},
			selectedIDs:  []string{},
			preferredIDs: []string{"a", "c"},
			want: Weights{
				"a": 2,
				"b": -1,
				"c": 2,
			},
		},
		{
			name:         "with preferred IDs and selected IDs",
			variables:    []string{"a", "b", "c"},
			selectedIDs:  []string{"b"},
			preferredIDs: []string{"a", "c"},
			want: Weights{
				"a": 2,
				"b": 5,
				"c": 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Create(
				tt.variables,
				tt.selectedIDs,
				tt.preferredIDs,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
