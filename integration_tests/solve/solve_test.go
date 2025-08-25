package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/pldag"
	"github.com/ourstudio-se/puan-sdk-go/weights"
)

const url = "http://127.0.0.1:9000"

func Test_package_variant_will_change_when_selecting_another_xor_component(t *testing.T) {
	/*
		Given package (a) -> (and(x,y,z), xor(n,m)), reversed package rules
		(and(x,y,z,n) -> (a), and(x,y,z,m) -> (a)) and with preferred
		on variant (a,n), we test that if variant (a,x,y,z,m) is preselected
		and we select single variable (n), then we will change into the other
		package variant (a,x,y,z,n) (and not select single (n))
	*/

	// Different selections rule [[a,x], [b]] [a,x] means in the same module, [[a],[x]] means that x is chosen in a different module

}

func Test_select_exactly_one_constrainted_component_with_additional_requirements(t *testing.T) {
	/*
		Exactly one of (a), (b) or (c) must be select. (b) requires another
		variable x. Now, (a) is preselected and we select (b). We expect (b,x) as result
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "item1")

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	packageB, _ := model.SetImply("packageB", "item1") // TODO: Check this, with and the test fail.

	preferredA, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreference := []weights.XORWithPreference{
		{
			XORID:              exactlyOnePackage,
			PreferredVariantID: preferredA,
		},
	}

	root, _ := model.SetAnd(exactlyOnePackage, packageB)
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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
}

func Test_select_same_not_constrainted_selected_component(t *testing.T) {
	/*
		(a) requires (b). (b) has been preselected and we select (b)
		again. We now expect the empty set as the result.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB")

	packageARequiredPackageB, _ := model.SetImply("packageA", "packageB")

	_ = model.Assume(packageARequiredPackageB)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
}

func Test_select_same_selected_exactly_one_constrainted_component(t *testing.T) {
	/*
		Exactly one of (a), (b) or (c) must be select but (a) is preferred.
		(b) has been preselected but is selected again. We now expect (a) to be selected.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC")

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	preferredA, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreference := []weights.XORWithPreference{
		{
			XORID:              exactlyOnePackage,
			PreferredVariantID: preferredA,
		},
	}

	root, _ := model.SetAnd(exactlyOnePackage)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
}

func Test_select_package_when_xor_between_packages_and_larger_package_is_selected(t *testing.T) {
	/*
		Two packages (a) and (b) exists, with (b) being the larger one
		and exactly one of them has to be selected.
		(b) has been preselected and we select (a). We know expect
		(a) to be selected without nothing left from (b).
	*/
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "item1", "item2", "item3")

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB")
	preferredA, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreference := []weights.XORWithPreference{
		{
			XORID:              exactlyOnePackage,
			PreferredVariantID: preferredA,
		},
	}

	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	root, _ := model.SetAnd(exactlyOnePackage, packageARequiredItems, packageBRequiredItems, itemsInAllPackages, itemsInPackageB)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_package_when_xor_between_packages(t *testing.T) {
	/*
		Two packages, (a) and (b), exists with (b) being the larger
		one. They both share a subset of variables and exactly one
		of (a) and (b) must be selected, but with (a) as preferred.
		With nothing being preselected, we select (a) and expects to get (a).
	*/
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "item1", "item2", "item3")

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB")
	preferredA, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreference := []weights.XORWithPreference{
		{
			XORID:              exactlyOnePackage,
			PreferredVariantID: preferredA,
		},
	}

	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	root, _ := model.SetAnd(exactlyOnePackage, packageARequiredItems, packageBRequiredItems, itemsInAllPackages, itemsInPackageB)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case1(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	// Case 1: No preselected packages, select package A
	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case2(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case3(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case4(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item4"])
}

func change_package_when_xor_between_multiple_packages_setup() (*pldag.Model, []weights.XORWithPreference) {
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "item1", "item2", "item3", "item4")

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")
	includedItemsInC, _ := model.SetAnd("item1", "item2", "item3", "item4")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)
	packageCRequiredItems, _ := model.SetImply("packageC", includedItemsInC)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	preferredA, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreference := []weights.XORWithPreference{
		{
			XORID:              exactlyOnePackage,
			PreferredVariantID: preferredA,
		},
	}

	anyOfThePackages, _ := model.SetOr("packageA", "packageB", "packageC")
	packageBOrC, _ := model.SetOr("packageB", "packageC")

	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageBOrC, _ := model.SetImply(includedItemsInB, packageBOrC)
	itemsInPackageC, _ := model.SetImply(includedItemsInC, "packageC")

	root, _ := model.SetAnd(exactlyOnePackage, packageARequiredItems, packageBRequiredItems, packageCRequiredItems, itemsInAllPackages, itemsInPackageBOrC, itemsInPackageC)
	_ = model.Assume(root)

	return model, xorWithPreference
}

func Test_downgrade_package_when_xor_between_multiple_packages_case1(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case2(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case3(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_default_component_in_package_when_part_in_multiple_xors(t *testing.T) { // TODO: Check that the logic is correctly implemented.
	/*
		Package (a) has two variants: (a,x,y,n) and (a,x,y,m,o) with
		preferred on the former. Nothing is preselected and we expect
		(a,x,y,n) as our result configuration.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "item1", "item2", "item3", "item4", "item5")

	sharedItemsInPackage, _ := model.SetAnd("item1", "item2")
	includedItemsInVariantOne, _ := model.SetAnd("item1", "item2", "item3")
	includedItemsInVariantTwo, _ := model.SetAnd("item1", "item2", "item4", "item5")

	packageVariantOne, _ := model.SetAnd("packageA", includedItemsInVariantOne)
	packageVariantTwo, _ := model.SetAnd("packageA", includedItemsInVariantTwo)

	exactlyOnePackage, _ := model.SetXor(packageVariantOne, packageVariantTwo)
	preferredA, _ := model.SetAnd(packageVariantOne, exactlyOnePackage)
	xorWithPreference := []weights.XORWithPreference{
		{
			XORID:              exactlyOnePackage,
			PreferredVariantID: preferredA,
		},
	}

	anyOfTheVariants, _ := model.SetOr(packageVariantOne, packageVariantTwo)
	itemsInAllPackages, _ := model.SetImply(sharedItemsInPackage, anyOfTheVariants)
	itemsInPackageVariantOne, _ := model.SetImply(packageVariantOne, includedItemsInVariantOne)
	itemsInPackageVariantTwo, _ := model.SetImply(packageVariantTwo, includedItemsInVariantTwo)

	root, _ := model.SetAnd(exactlyOnePackage, itemsInAllPackages, itemsInPackageVariantOne, itemsInPackageVariantTwo)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item5"])
}

func Test_select_component_with_indirect_package_requirement(t *testing.T) {
	/*
		There exists a chain of requirements: (e) -> (f) -> (a) -> (x,y,z).
		We select (e) and expect our result configuration to (e,f,a,x,y,z)
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageE", "packageF", "item1", "item2", "item3")

	includedItemsInA, _ := model.SetAnd("item1", "item2", "item3")
	packageA, _ := model.SetAnd("packageA", includedItemsInA)

	packageERequiresF, _ := model.SetImply("packageE", "packageF")
	packageFRequiresA, _ := model.SetImply("packageF", packageA)

	root, _ := model.SetAnd(packageERequiresF, packageFRequiresA)
	_ = model.Assume(root)

	selections := weights.Selections{
		{
			ID:     "packageE",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageE"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageF"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func Test_select_single_xor_component_when_another_xor_pair_is_preferred(t *testing.T) {
	/*
		Package (a) has two variants: (a,x) and (a,y,z) with the latter
		being preferred. We select (a,x) and expect the result configuration
		(a,x)
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item1",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_xor_pair_when_xor_pair_is_preferred(t *testing.T) {
	/*
	   	Package (a) has two variants: (a,x) and (a,y,z) with the latter
	      being preferred. We select (a,y,z) and expect the result configuration
	      (a,y,z). This test is just to make sure that there is no weird behavior
	      such as an empty configuration as result.
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{ // TODO: How should packages be handled as selections?
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
		{
			ID: "item3",
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func Test_deselect_package_when_xor_pair_is_preferred_over_single_xor_component(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a(yz) is already selected, check that we will remove package when deselecting a
	*/
	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{ // TODO: How should packages be handled as selections?
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
		{
			ID:     "item3",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.REMOVE,
		},
		{
			ID:     "item2",
			Action: weights.REMOVE,
		},
		{
			ID:     "item3",
			Action: weights.REMOVE,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_single_xor_component_when_xor_pair_is_already_selected(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a(yz) is already selected, check that we will select a(x) variant when selecting x
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{ // TODO: How should packages be handled as selections?
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
		{
			ID:     "item3",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item1",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_one_component_in_xor_pair_when_single_xor_component_is_already_selected(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a(x) is already selected, check that we will get ayz config when selecting y (or z)
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item1",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, xorWithPreference)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func select_single_xor_component_when_another_xor_pair_is_preferred_setup() (*pldag.Model, []weights.XORWithPreference) {
	model := pldag.New()
	model.SetPrimitives("packageA", "item1", "item2", "item3")

	includedItemsInPackageVariantTwo, _ := model.SetAnd("item2", "item3")
	packageVariantOne, _ := model.SetAnd("packageA", "item1")
	packageVariantTwo, _ := model.SetAnd("packageA", includedItemsInPackageVariantTwo)

	exactlyOnePackage, _ := model.SetXor(packageVariantOne, packageVariantTwo)
	packageVariants, _ := model.SetImply("packageA", exactlyOnePackage)

	xorWithPreference := []weights.XORWithPreference{
		{
			XORID:              exactlyOnePackage,
			PreferredVariantID: packageVariantTwo,
		},
	}

	reversedVariantOne, _ := model.SetImply("item1", packageVariantOne)
	reversedVariantTwo, _ := model.SetImply(includedItemsInPackageVariantTwo, packageVariantTwo)

	exactlyItem1OrItem2, _ := model.SetXor("item1", "item2") // TODO: are x and y forbidden together?
	exactlyItem1OrItem3, _ := model.SetXor("item1", "item3")
	exactlyOneCombinationOfItems, _ := model.SetAnd(exactlyItem1OrItem2, exactlyItem1OrItem3)
	itemVariants, _ := model.SetImply("packageA", exactlyOneCombinationOfItems)

	root, _ := model.SetAnd(packageVariants, reversedVariantOne, reversedVariantTwo, itemVariants)
	_ = model.Assume(root)

	return model, xorWithPreference
}

func Test_will_change_package_variant_when_package_is_preselected_with_component_requiring_package(t *testing.T) { // TODO: Check that the logic is correctly implemented.
	/*
		# Following rules are applied (with preferreds on the left xor-component)
		# a -> x
		# a -> b
		# a -> ~c
		# a -> ~d
		# b -> xor(c, a)
		# c -> ~a
		# x -> xor(c, a)
		# x -> xor(d, b)
		# Our case is that a is already selected, which indirectly will add
		# package x with its preferred components c and d
		# Then we select xc and we expect a to be replaced
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "packageD", "packageX")

	notPackageA, _ := model.SetNot("packageA")
	notPackageC, _ := model.SetNot("packageC")
	notPackageD, _ := model.SetNot("packageD")

	packageARequiresX, _ := model.SetImply("packageA", "packageX")
	packageARequiresB, _ := model.SetImply("packageA", "packageB")
	packageAForbidsPackageC, _ := model.SetImply("packageA", notPackageC)
	packageAForbidsPackageD, _ := model.SetImply("packageA", notPackageD)
	packageCForbidsPackageA, _ := model.SetImply("packageC", notPackageA)

	exactlyOneOfPackageCOrA, _ := model.SetXor("packageC", "packageA")
	exactlyOneOfPackageDOrB, _ := model.SetXor("packageD", "packageB")

	packageBRequiresExactlyPackageCOrA, _ := model.SetImply("packageB", exactlyOneOfPackageCOrA)
	packageXRequiresExactlyPackageCOrA, _ := model.SetImply("packageX", exactlyOneOfPackageCOrA)
	packageXRequiresExactlyPackageDOrB, _ := model.SetImply("packageX", exactlyOneOfPackageDOrB)

	packageBVariantOne, _ := model.SetAnd("packageB", exactlyOneOfPackageCOrA)
	xorWithPreferencePackageB := weights.XORWithPreference{
		XORID:              exactlyOneOfPackageCOrA,
		PreferredVariantID: packageBVariantOne,
	}

	packageXVariantOne, _ := model.SetAnd("packageX", exactlyOneOfPackageCOrA)
	xorWithPreferencePackageXOne := weights.XORWithPreference{
		XORID:              exactlyOneOfPackageCOrA,
		PreferredVariantID: packageXVariantOne,
	}

	packageXVariantTwo, _ := model.SetAnd("packageX", exactlyOneOfPackageDOrB)
	xorWithPreferencePackageXTwo := weights.XORWithPreference{
		XORID:              exactlyOneOfPackageDOrB,
		PreferredVariantID: packageXVariantTwo,
	}

	root, _ := model.SetAnd(packageARequiresX, packageARequiresB, packageAForbidsPackageC, packageAForbidsPackageD, packageCForbidsPackageA, packageBRequiresExactlyPackageCOrA, packageXRequiresExactlyPackageCOrA, packageXRequiresExactlyPackageDOrB)
	_ = model.Assume(root)

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "packageX",
			Action: weights.ADD,
		},
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
		{
			ID:     "packageD",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreferencePackageB, xorWithPreferencePackageXOne, xorWithPreferencePackageXTwo})

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageD"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageX"])
}

/*
OLD TESTS BELOW
*/

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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	preferredA, _ := model.SetAnd("packageA", exactlyOneOfThePackages)

	xorWithPreference := weights.XORWithPreference{
		XORID:              exactlyOneOfThePackages,
		PreferredVariantID: preferredA,
	}

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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreference})

	client := glpk.NewClient(url)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
}

func Test_Select_Packages_When_XOR_Between_Packages(t *testing.T) { // Denna är ok
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

	packageA, _ := model.SetImply("packageA", includedItemsInA)
	packageB, _ := model.SetImply("packageB", includedItemsInB)

	exactlyOneOfThePackages, _ := model.SetXor("packageA", "packageB")
	preferredA, _ := model.SetAnd("packageA", exactlyOneOfThePackages)
	// This to ensure that if the items are selected, then the package is selected as well.
	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	reverseBothPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	reversePackageB, _ := model.SetImply(includedItemsInB, "packageB")

	xorWithPreference := weights.XORWithPreference{
		XORID:              exactlyOneOfThePackages,
		PreferredVariantID: preferredA,
	}

	root, _ := model.SetAnd(exactlyOneOfThePackages, reverseBothPackages, reversePackageB, packageA, packageB)
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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreference})

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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective = weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective = weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective = weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective = weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

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
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)

	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func Test_XOR_With_Preferred(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"item1", "item2"}...)

	exactlyOneItem, _ := model.SetXor([]string{"item1", "item2"}...)
	xorWithPreferred := weights.XORWithPreference{
		XORID:              exactlyOneItem,
		PreferredVariantID: "item2",
	}
	_ = model.Assume(exactlyOneItem)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreferred})

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
}

func Test_Select_Same_Selected_Exactly_One_Constrainted_Component(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"item1", "item2", "item3"}...)

	exactlyOneItem, _ := model.SetXor([]string{"item1", "item2", "item3"}...)

	xorWithPreferred := weights.XORWithPreference{
		XORID:              exactlyOneItem,
		PreferredVariantID: "item1",
	}

	_ = model.Assume(exactlyOneItem)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "item2",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.REMOVE,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreferred})

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_default_component_in_package_when_part_in_multiple_xors_old(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "item1", "item2", "item3", "item4", "item5"}...)

	packageAVariantOne, _ := model.SetAnd("packageA", "item1", "item2", "item3")
	packageAVariantTwo, _ := model.SetAnd("packageA", "item1", "item2", "item4", "item5")

	exactlyOneVariant, _ := model.SetXor(packageAVariantOne, packageAVariantTwo)

	xorWithPreferred := weights.XORWithPreference{
		XORID:              exactlyOneVariant,
		PreferredVariantID: packageAVariantOne,
	}

	model.Assume(exactlyOneVariant)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreferred})

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item5"])
}

func Test_default_component_in_package_when_part_in_multiple_xors_heavy_variant_preferred(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "item1", "item2", "item3", "item4", "item5"}...)

	packageAVariantOne, _ := model.SetAnd("packageA", "item1", "item2", "item3")
	packageAVariantTwo, _ := model.SetAnd("packageA", "item1", "item2", "item4", "item5")

	exactlyOneVariant, _ := model.SetXor(packageAVariantOne, packageAVariantTwo)

	xorWithPreferred := weights.XORWithPreference{
		XORID:              exactlyOneVariant,
		PreferredVariantID: packageAVariantTwo,
	}
	model.Assume(exactlyOneVariant)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreferred})

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item4"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item5"])
}

func Test_select_single_xor_component_when_another_xor_pair_is_preferred1(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "item1", "item2", "item3"}...)

	packageAVariantOne, _ := model.SetAnd("packageA", "item1")
	packageAVariantTwo, _ := model.SetAnd("packageA", "item2", "item3")

	exactlyOneVariant, _ := model.SetXor(packageAVariantOne, packageAVariantTwo)
	packageAVariants, _ := model.SetImply("packageA", exactlyOneVariant)
	xorWithPreferred := weights.XORWithPreference{
		XORID:              exactlyOneVariant,
		PreferredVariantID: packageAVariantTwo,
	}

	model.Assume(packageAVariants)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item1",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreferred})

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_single_xor_component_when_another_xor_pair_is_preferred_no_selection(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "item1", "item2", "item3"}...)

	packageAVariantOne, _ := model.SetAnd("packageA", "item1")
	packageAVariantTwo, _ := model.SetAnd("packageA", "item2", "item3")

	exactlyOneVariant, _ := model.SetXor(packageAVariantOne, packageAVariantTwo)
	packageAVariants, _ := model.SetImply("packageA", exactlyOneVariant)

	xorWithPreferred := weights.XORWithPreference{
		XORID:              exactlyOneVariant,
		PreferredVariantID: packageAVariantTwo,
	}

	model.Assume(packageAVariants)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreferred})

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_single_xor_component_when_another_xor_pair_is_preferred_with_selection(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives([]string{"packageA", "item1", "item2", "item3", "item4", "item5"}...)

	packageAVariantTwoItems, _ := model.SetAnd("item2", "item3", "item4", "item5")
	packageAVariantOne, _ := model.SetAnd("packageA", "item1")
	packageAVariantTwo, _ := model.SetAnd("packageA", packageAVariantTwoItems)

	exactlyOneVariant, _ := model.SetXor(packageAVariantOne, packageAVariantTwo)

	packageAVariants, _ := model.SetImply("packageA", exactlyOneVariant)
	reversePackageAVariantOne, _ := model.SetImply("item1", "packageA")
	reversePackageAVariantTwo, _ := model.SetImply(packageAVariantTwoItems, "packageA")

	xorWithPreferred := weights.XORWithPreference{
		XORID:              exactlyOneVariant,
		PreferredVariantID: packageAVariantOne,
	}

	root, _ := model.SetAnd(packageAVariants, reversePackageAVariantOne, reversePackageAVariantTwo)
	model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		//{
		//	ID:     "packageA",
		//	Action: weights.ADD,
		//},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
		{
			ID:     "item3",
			Action: weights.ADD,
		},
		{
			ID:     "item4",
			Action: weights.ADD,
		},
		{
			ID:     "item5",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, []weights.XORWithPreference{xorWithPreferred})

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
}
