package build_release

import (
	"fmt"
	"regexp"
	"slices"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/rhysd/actionlint"
)


var goodWorkflowFile = 
`name: OSPS Baseline Scan

on: [workflow_dispatch]

jobs:
  scan:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Pull the pvtr-github-repo image
        run: docker pull eddieknight/pvtr-github-repo:latest

      - name: Add GitHub Secret to config file so it is protected in outputs
        run: |
          sed -i 's/{{ TOKEN }}/${{ secrets.TOKEN }}/g' ${{ github.workspace }}/.github/pvtr-config.yml

      - name: Scan all repos specified in .github/pvtr-config.yml
        run: |
          docker run --rm \
            -v ${{ github.workspace }}/.github/pvtr-config.yml:/.privateer/config.yml \
            -v ${{ github.workspace }}/docker_output:/evaluation_results \
            eddieknight/pvtr-github-repo:latest`


var badWorkflowFile =
`name: OSPS Baseline Scan

on: [workflow_dispatch]

jobs:
  scan:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Pull the pvtr-github-repo image
        run: docker pull eddieknight/pvtr-github-repo:latest

      - name: Add GitHub Secret to config file so it is protected in outputs
        run: |
          sed -i 's/{{ TOKEN }}/${{ secrets.TOKEN }}/g' ${{ github.event.review.body }}/.github/pvtr-config.yml

      - name: Scan all repos specified in .github/pvtr-config.yml
        run: |
          docker run --rm \
            -v ${{ github.event.issue.title }}/.github/pvtr-config.yml:/.privateer/config.yml \
            -v ${{ github.workspace }}/docker_output:/evaluation_results \
            eddieknight/pvtr-github-repo:latest`


type testingData struct {
	expectedResult bool
	workflowFile string
	assertionMessage string
}


func TestCicdSanitizedInputParameters (t * testing.T) {

	testData := []testingData {
		{
			expectedResult: false,
			workflowFile: badWorkflowFile,
			assertionMessage: "Untrusted input not detected",
		},
		{
			expectedResult: true,
			workflowFile: goodWorkflowFile,
			assertionMessage: "Untrusted input detected where it should not have been",
		},
	}

	for _, data := range testData {

		workflow, _ := actionlint.Parse([]byte(data.workflowFile))

		result, message := checkWorkflowFileForUntrustedInputs(workflow)

		fmt.Println(message)
		assert.Equal(t, result, data.expectedResult, data.assertionMessage)
	}
}


func TestVariableExtraction(t *testing.T) {

	var testScript = 
		`echo ${{github.event.issue.title }}
		if ${{ github.event.commits.arbitrary.data.message}} -ne 0
		then
			echo "Checkout report image" ${{ githubnodotevent.commits.arbitrary.data.message}}
			run: docker pull the pvt-r-github-repo image
		fi`

	varNames := pullVariablesFromScript(testScript)

	assert.Equal(t, slices.Contains(varNames, "github.event.issue.title"), true, "Variable extraction failed")
	assert.Equal(t, slices.Contains(varNames, "github.event.commits.arbitrary.data.message"), true, "Variable extraction failed")
	
}


func TestMultipleVariables(t *testing.T) {

	var testScript = `sed -i 's/{{ TOKEN }}/${{ secrets.TOKEN }}/g' ${{ github.event.review.body }}/.github/pvtr-config.yml`

	varNames := pullVariablesFromScript(testScript)
	assert.Equal(t, varNames[0], "secrets.TOKEN", "Variable extraction failed")
	assert.Equal(t, varNames[1], "github.event.review.body", "Variable extraction failed")

}


func TestRegex ( t * testing.T ) {

	expression, err := regexp.Compile(regex)
	if err != nil {
		t.Errorf("Error compiling regex: %v", err)
		return
	}

	assert.Equal(t, expression.Match([]byte("github.event.issue.title")), true, "regex match failed" )
	assert.Equal(t, expression.Match([]byte("github.event.commits.arbitrary.data.message")), true, "regex match failed" )
}


