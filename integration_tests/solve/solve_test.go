package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/pldag"
	"github.com/ourstudio-se/puan-sdk-go/weights"
)

const url = "http://127.0.0.1:9000"

func Test_And(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"a", "b"}...)
	andID, _ := model.SetAnd("a", "b")
	_ = model.Assume(andID)

	polyhedron := model.GeneratePolyhedron()

	client := glpk.NewClient(url)
	solution, _ := client.Solve(polyhedron, model.Variables(), map[string]int{})
	assert.Equal(t, 1, len(solution.Solutions))
	assert.Equal(t, glpk.SolutionValues{"a": 1, "b": 1, andID: 1}, solution.Solutions[0].Solution)
}

func Test_Or(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"a", "b"}...)
	orID, _ := model.SetOr("a", "b")
	_ = model.Assume(orID)

	polyhedron := model.GeneratePolyhedron()

	client := glpk.NewClient(url)
	solution, _ := client.Solve(polyhedron, model.Variables(), map[string]int{"a": 1, "b": -1})
	assert.Equal(t, 1, len(solution.Solutions))
	assert.Equal(t, glpk.SolutionValues{"a": 1, "b": 0, orID: 1}, solution.Solutions[0].Solution)
}

func Test_Add_Remove_Selection(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"a", "b"}...)
	implyID, _ := model.SetImply("a", "b")
	_ = model.Assume(implyID)

	polyhedron := model.GeneratePolyhedron()
	selections := weights.Selections{
		{
			ID:     "b",
			Action: weights.ADD,
		},
		{
			ID:     "b",
			Action: weights.REMOVE,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["a"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["b"])
}

func Test_Add_Condition_Selection(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"a", "b"}...)
	implyID, _ := model.SetImply("a", "b")
	_ = model.Assume(implyID)

	polyhedron := model.GeneratePolyhedron()
	selections := weights.Selections{
		{
			ID:     "a",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["a"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["b"])
}

func Test_Add_Consequence(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"a", "b"}...)
	implyID, _ := model.SetImply("a", "b")
	_ = model.Assume(implyID)

	polyhedron := model.GeneratePolyhedron()
	selections := weights.Selections{
		{
			ID:     "b",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["a"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["b"])
}
