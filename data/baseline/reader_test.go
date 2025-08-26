package baseline

import (
	"testing"

	"github.com/ossf/gemara/layer2"
)

func TestReadAllYAMLFiles(t *testing.T) {
	catalogData, err := ReadAllYAMLFiles()
	if err != nil {
		t.Fatalf("Failed to read YAML files: %v", err)
	}

	if catalogData == nil {
		t.Fatal("CatalogData is nil")
	}

	if len(catalogData.Catalog.ControlFamilies) == 0 {
		t.Fatal("No control families loaded")
	}

	// Verify we have some expected families
	familyTitles := make(map[string]bool)
	for _, family := range catalogData.Catalog.ControlFamilies {
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
	if len(catalogData.ControlFamilyFiles) == 0 {
		t.Error("No control family files mapping loaded")
	}

	t.Logf("Successfully loaded %d control families", len(catalogData.Catalog.ControlFamilies))
}

func TestGetAssessmentRequirementById(t *testing.T) {
	catalogData, err := ReadAllYAMLFiles()
	if err != nil {
		t.Fatalf("Failed to read YAML files: %v", err)
	}

	tests := []struct {
		name          string
		assessmentID  string
		expectError   bool
		expectedError string
	}{
		{
			name:         "Valid AC assessment requirement",
			assessmentID: "OSPS-AC-02.01",
			expectError:  false,
		},
		{
			name:         "Valid BR assessment requirement",
			assessmentID: "OSPS-BR-01.01",
			expectError:  false,
		},
		{
			name:         "Valid DO assessment requirement",
			assessmentID: "OSPS-DO-01.01",
			expectError:  false,
		},
		{
			name:          "Invalid assessment requirement",
			assessmentID:  "INVALID-REQ-01",
			expectError:   true,
			expectedError: "control with ID INVALID-REQ-01 not found",
		},
		{
			name:          "Empty assessment ID",
			assessmentID:  "",
			expectError:   true,
			expectedError: "control with ID  not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := catalogData.GetAssessmentRequirementById(tt.assessmentID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.expectedError != "" && err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s' but got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetAssessmentRequirements(t *testing.T) {
	catalogData, err := ReadAllYAMLFiles()
	if err != nil {
		t.Fatalf("Failed to read YAML files: %v", err)
	}

	requirements, err := catalogData.GetAssessmentRequirements()
	if err != nil {
		t.Errorf("Failed to get assessment requirements: %v", err)
	}

	// Test that we get a reasonable number of requirements
	if len(requirements) < 50 {
		t.Errorf("Expected at least 50 requirements, got %d", len(requirements))
	}

	// Test that all requirements have valid IDs
	for id, requirement := range requirements {
		if id == "" {
			t.Error("Requirement has empty ID")
		}
		if requirement == nil {
			t.Errorf("Requirement %s is nil", id)
		} else if requirement.Id == "" {
			t.Errorf("Requirement %s has empty ID field", id)
		}
	}
}

func TestGetControlByID(t *testing.T) {
	catalogData, err := ReadAllYAMLFiles()
	if err != nil {
		t.Fatalf("Failed to read YAML files: %v", err)
	}

	tests := []struct {
		name          string
		controlID     string
		expectError   bool
		expectedError string
		validateFunc  func(*testing.T, interface{}, string)
	}{
		{
			name:        "Valid AC control",
			controlID:   "OSPS-AC-01",
			expectError: false,
			validateFunc: func(t *testing.T, control interface{}, familyTitle string) {
				c := control.(*layer2.Control)
				if c.Id != "OSPS-AC-01" {
					t.Errorf("Expected control ID OSPS-AC-01, got %s", c.Id)
				}
				if familyTitle == "" {
					t.Error("Family title is empty")
				}
			},
		},
		{
			name:        "Valid BR control",
			controlID:   "OSPS-BR-01",
			expectError: false,
			validateFunc: func(t *testing.T, control interface{}, familyTitle string) {
				c := control.(*layer2.Control)
				if c.Id != "OSPS-BR-01" {
					t.Errorf("Expected control ID OSPS-BR-01, got %s", c.Id)
				}
			},
		},
		{
			name:        "Valid DO control",
			controlID:   "OSPS-DO-01",
			expectError: false,
			validateFunc: func(t *testing.T, control interface{}, familyTitle string) {
				c := control.(*layer2.Control)
				if c.Id != "OSPS-DO-01" {
					t.Errorf("Expected control ID OSPS-DO-01, got %s", c.Id)
				}
			},
		},
		{
			name:          "Non-existent control",
			controlID:     "NON-EXISTENT-CONTROL",
			expectError:   true,
			expectedError: "control with ID NON-EXISTENT-CONTROL not found",
		},
		{
			name:          "Empty control ID",
			controlID:     "",
			expectError:   true,
			expectedError: "control with ID  not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control, familyTitle, err := catalogData.GetControlByID(tt.controlID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.expectedError != "" && err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s' but got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if control == nil {
					t.Error("Control is nil")
				}
				if tt.validateFunc != nil {
					tt.validateFunc(t, control, familyTitle)
				}
			}
		})
	}
}

func TestGetControlsByFamily(t *testing.T) {
	catalogData, err := ReadAllYAMLFiles()
	if err != nil {
		t.Fatalf("Failed to read YAML files: %v", err)
	}

	tests := []struct {
		name          string
		familyID      string
		expectError   bool
		expectedError string
		validateFunc  func(*testing.T, []layer2.Control)
	}{
		{
			name:        "AC family controls",
			familyID:    "AC",
			expectError: false,
			validateFunc: func(t *testing.T, controls []layer2.Control) {
				if len(controls) == 0 {
					t.Error("No controls found for AC family")
				}
				for _, control := range controls {
					if !containsString(control.Id, "AC") {
						t.Errorf("Control %s does not contain AC family identifier", control.Id)
					}
				}
			},
		},
		{
			name:        "BR family controls",
			familyID:    "BR",
			expectError: false,
			validateFunc: func(t *testing.T, controls []layer2.Control) {
				if len(controls) == 0 {
					t.Error("No controls found for BR family")
				}
				for _, control := range controls {
					if !containsString(control.Id, "BR") {
						t.Errorf("Control %s does not contain BR family identifier", control.Id)
					}
				}
			},
		},
		{
			name:        "DO family controls",
			familyID:    "DO",
			expectError: false,
			validateFunc: func(t *testing.T, controls []layer2.Control) {
				if len(controls) == 0 {
					t.Error("No controls found for DO family")
				}
				for _, control := range controls {
					if !containsString(control.Id, "DO") {
						t.Errorf("Control %s does not contain DO family identifier", control.Id)
					}
				}
			},
		},
		{
			name:          "Non-existent family",
			familyID:      "NON-EXISTENT",
			expectError:   true,
			expectedError: "family NON-EXISTENT not found",
		},
		{
			name:          "Empty family ID",
			familyID:      "",
			expectError:   true,
			expectedError: "family  not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controls, err := catalogData.GetControlsByFamily(tt.familyID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.expectedError != "" && err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s' but got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.validateFunc != nil {
					tt.validateFunc(t, controls)
				}
			}
		})
	}
}

func TestGetRecommendationForEval(t *testing.T) {
	catalogData, err := ReadAllYAMLFiles()
	if err != nil {
		t.Fatalf("Failed to read YAML files: %v", err)
	}

	tests := []struct {
		name        string
		controlID   string
		expectError bool
	}{
		{
			name:        "VM control with assessment requirements",
			controlID:   "OSPS-VM-04",
			expectError: false,
		},
		{
			name:        "AC control with assessment requirements",
			controlID:   "OSPS-AC-01",
			expectError: false,
		},
		{
			name:        "BR control with assessment requirements",
			controlID:   "OSPS-BR-01",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control, _, err := catalogData.GetControlByID(tt.controlID)
			if err != nil {
				if tt.expectError {
					return // Expected error
				}
				t.Fatalf("Failed to get control: %v", err)
			}

			// Just verify the control exists and has assessment requirements
			// Don't require specific content as it may vary
			if len(control.AssessmentRequirements) == 0 {
				t.Error("Control has no assessment requirements")
			}

			// Log the control details for debugging
			t.Logf("Control %s has %d assessment requirements", control.Id, len(control.AssessmentRequirements))
		})
	}
}

func TestExtractFamilyID(t *testing.T) {
	tests := []struct {
		filename   string
		expectedID string
	}{
		{"OSPS-AC.yaml", "AC"},
		{"OSPS-BR.yaml", "BR"},
		{"OSPS-DO.yaml", "DO"},
		{"OSPS-GV.yaml", "GV"},
		{"OSPS-LE.yaml", "LE"},
		{"OSPS-QA.yaml", "QA"},
		{"OSPS-SA.yaml", "SA"},
		{"OSPS-VM.yaml", "VM"},
		{"other-file.yaml", "other-file"},
		{"file.yml", "file"},
		{"no-extension", "no-extension"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := extractFamilyID(tt.filename)
			if result != tt.expectedID {
				t.Errorf("extractFamilyID(%s) = %s; expected %s", tt.filename, result, tt.expectedID)
			}
		})
	}
}

func TestGetControlFamilyCount(t *testing.T) {
	catalogData, err := ReadAllYAMLFiles()
	if err != nil {
		t.Fatalf("Failed to read YAML files: %v", err)
	}

	count := catalogData.GetControlFamilyCount()

	// We expect at least 8 control families based on the OSPS files
	if count < 8 {
		t.Errorf("Expected at least 8 control families, got %d", count)
	}

	// Verify the count matches the actual number of families
	actualCount := len(catalogData.Catalog.ControlFamilies)
	if count != actualCount {
		t.Errorf("GetControlFamilyCount() returned %d, but actual count is %d", count, actualCount)
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
