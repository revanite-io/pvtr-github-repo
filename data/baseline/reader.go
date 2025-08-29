package baseline

import (
	"embed"
	"fmt"
	"path"

	"github.com/goccy/go-yaml"
	"github.com/ossf/gemara/layer2"
)

// We have tight control over the catalog right now while it is local, but
// if/when we have dynamic retrieval, we should retain the local files for testing
const dataDir string = "catalog"

//go:embed catalog
var files embed.FS

func GetAssessmentRequirements() (map[string]*layer2.AssessmentRequirement, error) {
	requirements := make(map[string]*layer2.AssessmentRequirement)
	catalog, err := loadCatalog()
	if err != nil {
		return nil, err
	}
	for _, family := range catalog.ControlFamilies {
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

// ReadAllYAMLFiles reads all YAML files in the data directory and returns the complete catalog data
func loadCatalog() (catalog layer2.Catalog, err error) {
	dir, err := files.ReadDir(dataDir)
	// Check if files are in the right place
	if err != nil {
		return catalog, fmt.Errorf("data directory does not exist: %s", dataDir)
	}

	catalog = layer2.Catalog{
		ControlFamilies: []layer2.ControlFamily{},
	}

	// Process each YAML file
	for _, file := range dir {
		filePath := path.Join(dataDir, file.Name())
		controlFamily, err := readYAMLFile(filePath)
		if err != nil {
			return catalog, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		catalog.ControlFamilies = append(catalog.ControlFamilies, *controlFamily)
	}

	return catalog, nil
}

// ReadYAMLFile reads a single YAML file and returns the control family data
func readYAMLFile(filePath string) (*layer2.ControlFamily, error) {
	data, err := files.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var yamlData layer2.Catalog
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	if len(yamlData.ControlFamilies) == 0 {
		return nil, fmt.Errorf("no control families found in file: %s", filePath)
	}

	// Assuming one control family per file as per the current structure
	familyData := yamlData.ControlFamilies[0]

	controlFamily := &layer2.ControlFamily{
		Id:          familyData.Id, // Use the ID from the YAML data
		Title:       familyData.Title,
		Description: familyData.Description,
		Controls:    familyData.Controls,
	}

	return controlFamily, nil
}
