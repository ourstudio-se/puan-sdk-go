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

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		variables   []string
		selectedIDs []string
		want        Weights
	}{
		{
			"no selected IDs",
			[]string{"a", "b", "c"},
			[]string{},
			Weights{
				"a": -1,
				"b": -1,
				"c": -1,
			},
		},
		{
			"a selected",
			[]string{"a", "b", "c"},
			[]string{"a"},
			Weights{
				"a": 3,
				"b": -1,
				"c": -1,
			},
		},
		{
			"a, b selected",
			[]string{"a", "b", "c"},
			[]string{"a", "b"},
			Weights{
				"a": 2,
				"b": 4,
				"c": -1,
			},
		},
		{
			"a, b, c selected",
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
			Weights{
				"a": 1,
				"b": 2,
				"c": 4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Create(tt.variables, tt.selectedIDs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
