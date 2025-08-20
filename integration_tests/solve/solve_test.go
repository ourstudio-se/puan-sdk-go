package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/pldag"
	"github.com/ourstudio-se/puan-sdk-go/weights"
)

const url = "http://127.0.0.1:9000"

func Test_Remove_Selection(t *testing.T) {
	/*
		(packageA) requires (packageB). (packageB) has been preselected, and we remove (packageB)
		again. We now expect that (packageA) and (packageB) to be zero.
	*/

	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB"}...)
	packageARequiresB, _ := model.SetImply("packageA", "packageB")
	_ = model.Assume(packageARequiresB)

	polyhedron := model.GeneratePolyhedron()
	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.REMOVE,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
}

func Test_Add_Consequence(t *testing.T) {
	/*
		(packageA) requires (packageB) only (packageB) has been selected.
		We now expect that (packageA) is false and (packageB) is true.
	*/
	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB"}...)
	packageARequiresB, _ := model.SetImply("packageA", "packageB")
	_ = model.Assume(packageARequiresB)

	polyhedron := model.GeneratePolyhedron()
	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
}

func Test_Select_Exactly_One_With_Additional_Requirements(t *testing.T) {
	/*
		Exactly one of (packageA), (packageB) or (packageC) must be selected. (packageB) requires another
		variable item1. Now, (packageA) is preselected, and we select (packageB). We expect (packageB, item1) to be true.
	*/

	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB", "packageC", "item1"}...)
	exactlyOneOfThePackages, _ := model.SetXor("packageA", "packageB", "packageC")
	packageBRequiresItem1, _ := model.SetImply("packageB", "item1")

	root, _ := model.SetAnd(exactlyOneOfThePackages, packageBRequiresItem1)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
}

func Test_Select_Packages_When_XOR_Between_Packages(t *testing.T) {
	/*
		Two packages (A) and (B) exist, with (B) being the larger one
		and exactly one of them has to be selected.
		(B) has been preselected, and we select (A). We know expect
		(A) to be selected without nothing left from (B).
	*/

	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB", "item1", "item2", "item3"}...)

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)

	exactlyOneOfThePackages, _ := model.SetXor("packageA", "packageB")

	// This to ensure that if the items are selected, then the package is selected as well.
	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAIsInBothPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInBIsPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	root, _ := model.SetAnd(packageARequiredItems, packageBRequiredItems, exactlyOneOfThePackages, itemsInAIsInBothPackages, itemsInBIsPackageB)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()

	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_Cheapest_Package_When_XOR_Between(t *testing.T) {
	/*
		Two packages, (A) and (B), exists with (B) being the larger
		one. They both share a subset of variables and exactly one
		of (A) and (B) must be selected, but with (A) as preferred.
		With nothing being preselected, we select (A) and expects to
		get (A).
	*/

	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB", "item1", "item2", "item3"}...)

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)

	exactlyOneOfThePackages, _ := model.SetXor("packageA", "packageB")

	// This to ensure that if the items are selected, then the corresponding package is selected as well.
	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	root, _ := model.SetAnd(packageARequiredItems, packageBRequiredItems, exactlyOneOfThePackages, itemsInAllPackages, itemsInPackageB)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_Upgrade_Package_When_XOR_Between(t *testing.T) {
	/*
		Here are three packages, (A), (B) and (C), exists. (C) is larger
		than (B) and (B) is larger than (A). We will do several test going from
		nothing preselected to from (B) preselected while selecting (C).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB", "packageC", "item1", "item2", "item3", "item4"}...)

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")
	includedItemsInC, _ := model.SetAnd("item1", "item2", "item3", "item4")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)
	packageCRequiredItems, _ := model.SetImply("packageC", includedItemsInC)

	exactlyOneOfThePackages, _ := model.SetXor("packageA", "packageB", "packageC")

	// This to ensure that if the items are selected, then the corresponding package is selected as well.
	anyOfThePackages, _ := model.SetOr("packageA", "packageB", "packageC")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)

	anyOfPackagesBC, _ := model.SetOr("packageB", "packageC")
	itemsInPackagesBC, _ := model.SetImply(includedItemsInB, anyOfPackagesBC)

	itemsInPackageC, _ := model.SetImply(includedItemsInC, "packageC")

	root, _ := model.SetAnd(packageARequiredItems, packageBRequiredItems, packageCRequiredItems, exactlyOneOfThePackages, itemsInAllPackages, itemsInPackagesBC, itemsInPackageC)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])

	selections = weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}
	selectionsIDs = selections.ExtractActiveSelectionIDS()
	objective = weights.Create(model.Variables(), selectionsIDs)

	resp, _ = client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])

	selections = weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
	}
	selectionsIDs = selections.ExtractActiveSelectionIDS()
	objective = weights.Create(model.Variables(), selectionsIDs)

	resp, _ = client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item4"])
}

func Test_Downgrade_Package_When_XOR_Between(t *testing.T) {
	/*
		Here we have three packages, (A), (B) and (c), with (C) being largest
		and (A) being smallest. We will try and select (B) when (C) is preselected,
		try select (A) and (B) is selected and try (A) when (C) is selected. All
		tests should result in the selected package.
	*/

	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB", "packageC", "item1", "item2", "item3", "item4"}...)

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")
	includedItemsInC, _ := model.SetAnd("item1", "item2", "item3", "item4")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)
	packageCRequiredItems, _ := model.SetImply("packageC", includedItemsInC)

	exactlyOneOfThePackages, _ := model.SetXor("packageA", "packageB", "packageC")

	// This to ensure that if the items are selected, then the corresponding package is selected as well.
	anyOfThePackages, _ := model.SetOr("packageA", "packageB", "packageC")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)

	anyOfPackagesBC, _ := model.SetOr("packageB", "packageC")
	itemsInPackagesBC, _ := model.SetImply(includedItemsInB, anyOfPackagesBC)

	itemsInPackageC, _ := model.SetImply(includedItemsInC, "packageC")

	root, _ := model.SetAnd(packageARequiredItems, packageBRequiredItems, packageCRequiredItems, exactlyOneOfThePackages, itemsInAllPackages, itemsInPackagesBC, itemsInPackageC)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])

	selections = weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}
	selectionsIDs = selections.ExtractActiveSelectionIDS()
	objective = weights.Create(model.Variables(), selectionsIDs)

	resp, _ = client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])

	selections = weights.Selections{
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}
	selectionsIDs = selections.ExtractActiveSelectionIDS()
	objective = weights.Create(model.Variables(), selectionsIDs)

	resp, _ = client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_Select_Component_With_Indirect_Package_Requirement(t *testing.T) {
	/*
		There exists a chain of requirements: (packageE) -> (packageF) -> (packageA) -> (item1, item2, item3).
		We select (packageE) and expect our result configuration to (packageE, packageF, packageA, item1, item2, item3)
	*/
	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB", "packageC", "packageD", "packageE", "packageF", "item1", "item2", "item3"}...)

	includedItemsInA, _ := model.SetAnd("item1", "item2", "item3")
	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageERequiredPackageF, _ := model.SetImply("packageE", "packageF")
	packageFRequiredPackageA, _ := model.SetImply("packageF", "packageA")

	root, _ := model.SetAnd(packageERequiredPackageF, packageFRequiredPackageA, packageARequiredItems)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()

	selections := weights.Selections{
		{
			ID:     "packageE",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageE"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageF"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func Test_Will_Delete_Package_From_Selected_Actions_When_Adding_Upgrading_Package(t *testing.T) {
	/*
	   Following rules are applied (with preferreds on the left xor-component)
	   packageA -> (item1, item2)
	   packageB -> (item1, item2, item3)
	   packageA -> -packageB
	   packageB -> -packageA
	   (item1, item2) -> or(packageA, packageB)
	   (item1, item2, item3) -> packageB

	   We have already selected package A and now we select package B. We expect B to be selected
	   and A deleted from pre selected actions
	*/

	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "packageB", "item1", "item2", "item3"}...)

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)

	onlyOnePackage, _ := model.SetOneOrNone("packageA", "packageB")

	// This to ensure that if the items are selected, then the corresponding package is selected as well.
	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	root, err := model.SetAnd(packageARequiredItems, packageBRequiredItems, itemsInAllPackages, itemsInPackageB, onlyOnePackage)
	if err != nil {
		panic(err)
	}

	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.Variables(), selectionsIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}
