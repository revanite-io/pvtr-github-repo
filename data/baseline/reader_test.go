package baseline

import (
	"testing"

	"github.com/ossf/gemara/layer2"
)

func TestReader_ReadAllYAMLFiles(t *testing.T) {
	reader := NewReader()

	baselineData, err := reader.ReadAllYAMLFiles()
	if err != nil {
		t.Fatalf("Failed to read YAML files: %v", err)
	}

	if baselineData == nil {
		t.Fatal("BaselineData is nil")
	}

	if len(baselineData.Catalog.ControlFamilies) == 0 {
		t.Fatal("No control families loaded")
	}

	// Verify we have some expected families
	familyTitles := make(map[string]bool)
	for _, family := range baselineData.Catalog.ControlFamilies {
		familyTitles[family.Title] = true

		// Each family should have a title and description
		if family.Title == "" {
			t.Error("Control family missing title")
		}
		if family.Description == "" {
			t.Error("Control family missing description")
		}

		// Each family should have controls
		if len(family.Controls) == 0 {
			t.Errorf("Control family %s has no controls", family.Title)
		}

		// Verify controls have required fields
		for _, control := range family.Controls {
			if control.Id == "" {
				t.Error("Control missing ID")
			}
			if control.Title == "" {
				t.Errorf("Control %s missing title", control.Id)
			}
			if control.Objective == "" {
				t.Errorf("Control %s missing objective", control.Id)
			}
		}
	}

	// Verify we loaded the control family files mapping
	if len(baselineData.ControlFamilyFiles) == 0 {
		t.Error("No control family files mapping loaded")
	}

	t.Logf("Successfully loaded %d control families", len(baselineData.Catalog.ControlFamilies))
}

func TestReader_GetAssessmentById(t *testing.T) {

	reader := NewReader()

	req, err := reader.GetAssessmentRequirementById("OSPS-AC-02.01")
	if err != nil {
		t.Errorf("Failed to get assessment requirement: %v", err)
	}
	if req == nil {
		t.Error("Assessment requirement is nil")
	}

}

func TestBuildAssessmentRequirementMap(t *testing.T) {

	br := NewReader()

	data, err := br.ReadAllYAMLFiles()
	if err != nil {
		t.Errorf("TestBuildAssessmentRequirementMap Could not read Yaml files: %s", err)
	}

	requirements := make(map[string]*layer2.AssessmentRequirement)

	for _, family := range data.Catalog.ControlFamilies {
		for _, control := range family.Controls {
			for _, requirement := range control.AssessmentRequirements {
				requirements[requirement.Id] = &requirement
			}
		}
	}

	expected := 60

	if len(requirements) != expected {
		t.Errorf("List of requirements - wanted: %v  got: %v \n", expected, len(requirements))
	}

}

func TestReader_GetControlByID(t *testing.T) {

	reader := NewReader()

	// Test getting a known control (assuming OSPS-AC-01 exists)
	control, familyTitle, err := reader.GetControlByID("OSPS-AC-01")
	if err != nil {
		t.Fatalf("Failed to get control OSPS-AC-01: %v", err)
	}

	if control == nil {
		t.Fatal("Control is nil")
	}

	if control.Id != "OSPS-AC-01" {
		t.Errorf("Expected control ID OSPS-AC-01, got %s", control.Id)
	}

	if familyTitle == "" {
		t.Error("Family title is empty")
	}

	// Test getting a non-existent control
	_, _, err = reader.GetControlByID("NON-EXISTENT-CONTROL")
	if err == nil {
		t.Error("Expected error for non-existent control")
	}

	t.Logf("Successfully retrieved control %s from family %s", control.Id, familyTitle)
}

func TestReader_GetControlsByFamily(t *testing.T) {
	reader := NewReader()

	// Test getting controls for AC family (assuming OSPS-AC.yaml exists)
	controls, err := reader.GetControlsByFamily("AC")
	if err != nil {
		t.Fatalf("Failed to get controls for family AC: %v", err)
	}

	if len(controls) == 0 {
		t.Fatal("No controls found for AC family")
	}

	// Verify all controls have AC prefix
	for _, control := range controls {
		if !containsString(control.Id, "AC") {
			t.Errorf("Control %s does not contain AC family identifier", control.Id)
		}
	}

	// Test getting controls for non-existent family
	_, err = reader.GetControlsByFamily("NON-EXISTENT")
	if err == nil {
		t.Error("Expected error for non-existent family")
	}

	t.Logf("Successfully retrieved %d controls for AC family", len(controls))
}

func TestReader_GetReccomendationForEval(t *testing.T) {
	reader := NewReader()

	control, _, _ := reader.GetControlByID("OSPS-VM-04")

	if len(control.AssessmentRequirements) == 0 {
		t.Error("Control has no assessment requirements")
	} else {
		// Then check if Recommendation is empty
		if control.AssessmentRequirements[0].Recommendation == "" {
			t.Error("Assessment requirement recommendation is empty")
		}
	}

}

func TestReader_ExtractFamilyID(t *testing.T) {
	reader := NewReader()

	tests := []struct {
		filename   string
		expectedID string
	}{
		{"OSPS-AC.yaml", "AC"},
		{"OSPS-BR.yaml", "BR"},
		{"OSPS-DO.yaml", "DO"},
		{"other-file.yaml", "other-file"},
		{"file.yml", "file"},
	}

	for _, test := range tests {
		result := reader.extractFamilyID(test.filename)
		if result != test.expectedID {
			t.Errorf("extractFamilyID(%s) = %s; expected %s", test.filename, result, test.expectedID)
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
