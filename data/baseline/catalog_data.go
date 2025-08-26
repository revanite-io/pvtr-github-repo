package baseline

import (
	"fmt"
	"strings"

	"github.com/ossf/gemara/layer2"
)

// CatalogData represents the complete baseline data structure
type CatalogData struct {
	// ControlFamilyFiles maps family ID to the file path it was loaded from
	ControlFamilyFiles map[string]string `json:"control_family_files"`
	// Catalog contains all the control families and their controls
	Catalog layer2.Catalog `json:"catalog"`
}

func NewCatalogData() (*CatalogData, error) {
	cd, err := ReadAllYAMLFiles()
	if err != nil {
		return nil, err
	}

	return cd, nil
}

// GetAssessmentRequirements returns all assessment requirements from the catalog
func (cd *CatalogData) GetAssessmentRequirements() (map[string]*layer2.AssessmentRequirement, error) {
	requirements := make(map[string]*layer2.AssessmentRequirement)

	for _, family := range cd.Catalog.ControlFamilies {
		for _, control := range family.Controls {
			for _, requirement := range control.AssessmentRequirements {
				requirements[requirement.Id] = &requirement
			}
		}
	}

	if len(requirements) == 0 {
		return nil, fmt.Errorf("GetAssessmentRequirements: 0 requirements found")
	}

	return requirements, nil
}

// GetControlByID searches for a control by its ID across all control families
func (cd *CatalogData) GetControlByID(controlID string) (*layer2.Control, string, error) {
	for _, family := range cd.Catalog.ControlFamilies {
		for _, control := range family.Controls {
			if control.Id == controlID {
				return &control, family.Title, nil
			}
		}
	}

	return nil, "", fmt.Errorf("control with ID %s not found", controlID)
}

// GetAssessmentRequirementById retrieves an assessment requirement by its ID
func (cd *CatalogData) GetAssessmentRequirementById(assessmentID string) (*layer2.AssessmentRequirement, error) {
	// extract the control id
	controlID := strings.Split(assessmentID, ".")[0]
	control, _, err := cd.GetControlByID(controlID)

	if err != nil {
		return nil, err
	}

	for _, assessment := range control.AssessmentRequirements {
		if assessment.Id == assessmentID {
			return &assessment, nil
		}
	}

	return nil, fmt.Errorf("control with ID %s not found", controlID)
}

// GetControlsByFamily returns all controls for a specific family
func (cd *CatalogData) GetControlsByFamily(familyID string) ([]layer2.Control, error) {
	// Look for the family in the catalog by ID
	for _, family := range cd.Catalog.ControlFamilies {
		if family.Id == familyID {
			return family.Controls, nil
		}
	}

	return nil, fmt.Errorf("family %s not found", familyID)
}

// GetControlFamilyCount returns the number of control families available
func (cd *CatalogData) GetControlFamilyCount() int {
	return len(cd.Catalog.ControlFamilies)
}
