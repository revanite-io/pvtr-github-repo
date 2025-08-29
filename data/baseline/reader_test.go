package baseline

import (
	"fmt"
	"path"
	"testing"

	"github.com/ossf/gemara/layer2"
)

const testDataDir = "test_data"

func TestGetAssessmentRequirements(t *testing.T) {
	reqs, err := GetAssessmentRequirements()
	if err != nil {
		t.Error(err)
	}
	if len(reqs) == 0 {
		t.Errorf("expected  one or more requirements but got %v", len(reqs))
	}
}

func TestLoadCatalog(t *testing.T) {
	catalog, err := loadCatalog()
	if err != nil {
		t.Error(err)
	}
	if len(catalog.ControlFamilies) == 0 {
		t.Errorf("expected one or more control families but got %v", len(catalog.ControlFamilies))
	}
	for i, family := range catalog.ControlFamilies {
		t.Run(fmt.Sprintf("Test Control family %v", i), func(t *testing.T) {
			testFamily(t, family)
		})
	}
}

func TestReadYAMLFile(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		expectedError bool
	}{
		{
			name:          "ensure dataDir succeeds",
			file:          path.Join(dataDir, "OSPS-AC.yaml"),
			expectedError: false,
		},
		{
			name:          "ensure nonsense file fails",
			file:          path.Join(testDataDir, "malformed_catalog.yaml"),
			expectedError: true,
		},
		{
			name:          "ensure bad file fails",
			file:          path.Join(testDataDir, "x.yaml"),
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			family, err := readYAMLFile(tt.file)
			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
				return
			} else if tt.expectedError && (err != nil) {
				return
			} else if !tt.expectedError && (err != nil) {
				t.Errorf("Did not expect error that was found: %s", err.Error())
				return
			}
			if family == nil {
				t.Error(err)
			}
			if family == nil {
				t.Errorf("readYAMLFile returned nil instead of control family for test %s", tt.name)
			} else {
				testFamily(t, *family)
			}
		})
	}
}

func testFamily(t *testing.T, family layer2.ControlFamily) {
	if family.Id == "" {
		t.Error("expected control family to have id, but got none in family")
	}
	if family.Title == "" {
		t.Error("expected control family to have title, but got none in family")
	}
	if family.Description == "" {
		t.Error("expected control family to have description, but got none in family")
	}
	if len(family.Controls) == 0 {
		t.Error("expected control family to have controls, but got none in family")
	}
}
