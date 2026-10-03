package ri

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestGoGraphEmptyModuleInventorySurvivesCloneAndReplay(t *testing.T) {
	identity := goModuleFixture(t, map[string]string{"file.go": "package fixture\n"})
	inventory, err := CollectGoModuleInventory(context.Background(), identity)
	if err != nil {
		t.Fatal(err)
	}
	sourceID, err := identity.ID()
	if err != nil {
		t.Fatal(err)
	}
	binding, err := DeclaredGoPackageBinding(inventory, "file.go", "fixture")
	if err != nil {
		t.Fatal(err)
	}
	producer := strings.Repeat("b", 64)
	input := graphInput("file.go", "package fixture\n", binding, nil, nil, nil, nil, producer)
	graph, err := BuildGoEngineeringGraph(GoGraphSnapshotInput{SourceID: sourceID, ProducerSHA256: producer, Files: []GoGraphFileInput{input}, ModuleInventory: &inventory})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateGoEngineeringGraph(graph); err != nil {
		t.Fatal(err)
	}
	if graph.ModuleInventory == nil || !reflect.DeepEqual(*graph.ModuleInventory, inventory) {
		t.Fatal("graph clone changed empty inventory representation")
	}
	encoded, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	var replay GoEngineeringGraph
	if err := json.Unmarshal(encoded, &replay); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGoEngineeringGraph(replay); err != nil {
		t.Fatal(err)
	}
}

func TestGoModuleInventoryClonePreservesSliceShapeAndOwnership(t *testing.T) {
	for _, inventory := range []GoModuleInventory{
		{},
		{Files: []GoManifestObservation{}, Omissions: []GoManifestOmission{}},
		{Files: []GoManifestObservation{{Requires: []GoModuleRequirement{}, Replaces: []GoModuleReplacement{}, Uses: []GoWorkspaceUse{}}}},
		{Files: []GoManifestObservation{{Requires: []GoModuleRequirement{{Path: "original"}}}}, Omissions: []GoManifestOmission{{Path: "original"}}},
	} {
		clone := cloneGoModuleInventory(inventory)
		if !reflect.DeepEqual(clone, inventory) {
			t.Fatal("clone changed nil/empty shape or declarations")
		}
		if len(clone.Omissions) > 0 {
			clone.Omissions[0].Path = "changed"
			if inventory.Omissions[0].Path != "original" {
				t.Fatal("clone aliases omission storage")
			}
		}
		if len(clone.Files) > 0 && len(clone.Files[0].Requires) > 0 {
			clone.Files[0].Requires[0].Path = "changed"
			if inventory.Files[0].Requires[0].Path != "original" {
				t.Fatal("clone aliases nested declaration storage")
			}
		}
	}
}
