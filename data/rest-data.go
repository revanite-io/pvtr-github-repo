package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ossf/si-tooling/v2/si"
	"github.com/privateerproj/privateer-sdk/config"
)

type RestData struct {
	owner        string
	repo         string
	token        string
	Config       *config.Config
	Organization OrgData
	Workflow     Workflow
	Insights     si.SecurityInsights
	Name         string `json:"name"`
	Private      bool   `json:"private"`
	WebsiteURL   string `json:"websiteUrl"`
	Releases     []ReleaseData
	Contents     struct {
		TopLevel  []DirContents
		ForgeDir  []DirContents
		WorkFlows []DirFile
	}
	Rulesets []Ruleset
}

type Ruleset struct {
	Type       string `json:"type"`
	Parameters struct {
		RequiredChecks []struct {
			Context string `json:"context"`
		} `json:"required_status_checks"`
	} `json:"parameters"`
}

type OrgData struct {
	Name               string        `json:"name"`
	Blog               string        `json:"blog"`
	WebSignoffRequired bool          `json:"web_commit_signoff_required"`
	TwoFactorRequired  *nullableBool `json:"two_factor_requirement_enabled"`
}

type ReleaseData struct {
	Id      int            `json:"id"`
	Name    string         `json:"name"`
	TagName string         `json:"tag_name"`
	URL     string         `json:"url"`
	Assets  []ReleaseAsset `json:"assets"`
}

type ReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

type DirContents struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
}

type DirFile struct {
	DirContents
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type FileAPIResponse struct {
	ByteContent []byte `json:"content"`
	SHA         string `json:"sha"`
}

type Workflow struct {
	DefaultPermissions    string `json:"default_workflow_permissions"`
	CanApprovePullRequest bool   `json:"can_approve_pull_request_reviews"`
}

// Golang bools are binary, but JSON bools can also be null.
// If null is found, the value of a golang bool is set to false, but
// the GitHub API sometimes uses the third value when the call is unauthenticated.
type nullableBool bool

var APIBase = "https://api.github.com"

func (n *nullableBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	*n = nullableBool(b)
	return nil
}

func (r *RestData) Setup() error {
	r.owner = r.Config.GetString("owner")
	r.repo = r.Config.GetString("repo")
	r.token = r.Config.GetString("token")

	_ = r.getMetadata()
	r.loadSecurityInsights()
	_ = r.getWorkflow()
	_ = r.getReleases()
	r.loadOrgData()
	_ = r.getWorkflowFiles()
	return nil
}

func (r *RestData) MakeApiCall(endpoint string, isGithub bool) (body []byte, err error) {
	r.Config.Logger.Trace(fmt.Sprintf("GET %s", endpoint))
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	if isGithub {
		request.Header.Set("Authorization", "Bearer "+r.token)
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		err = fmt.Errorf("error making http call: %s", err.Error())
		return nil, err
	}
	if response.StatusCode != 200 {
		err = fmt.Errorf("unexpected response: %s", response.Status)
		return nil, err
	}
	return io.ReadAll(response.Body)
}

func (r *RestData) getSourceFile(owner, repo, path string) (response FileAPIResponse, err error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s", APIBase, owner, repo, path)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		return
	}
	err = json.Unmarshal(responseData, &response)
	return
}

func (r *RestData) loadSecurityInsights() {
	r.getTopDirContents()
	if len(r.Contents.TopLevel) == 0 {
		r.Config.Logger.Error("no contents retrieved from the top level of the repository")
		return
	}
	for _, dirContents := range r.Contents.TopLevel {
		insightsFileName := r.foundSecurityInsights(dirContents)
		if insightsFileName != "" {
			insights, err := si.Read(r.owner, r.repo, insightsFileName)
			r.Insights = insights
			if err != nil {
				r.Config.Logger.Error(fmt.Sprintf("error reading security insights file: %s", err.Error()))
			}
			return
		}
	}
	r.getForgeDirContents()
	for _, dirContents := range r.Contents.ForgeDir {
		insightsFileName := r.foundSecurityInsights(dirContents)
		if insightsFileName != "" {
			insights, err := si.Read(r.owner, r.repo, fmt.Sprintf(".github/%s", insightsFileName))
			r.Insights = insights
			if err != nil {
				r.Config.Logger.Error(fmt.Sprintf("error reading security insights file: %s", err.Error()))
			}
			return
		}
	}
}

func (r *RestData) getWorkflowFiles() error {

	//Only subdirectories are not allowed in the .github/workflows directory, so no need to recurse
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/.github/workflows", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("Error calling github to retrive workflow files list: %s", err.Error()))
		return err
	}

	var workflowFileList []DirContents
	err = json.Unmarshal(responseData, &workflowFileList)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("Error unmarshalling json response for workflow files list: %s", err.Error()))
		return err
	}

	//For each file, listed we need to get it and put it in a format the action parser can use
	var dirFiles = make([]DirFile, len(workflowFileList))
	for i, workflowFile := range workflowFileList {

		response, err := r.MakeApiCall(workflowFile.URL, true)
		if err != nil {
			r.Config.Logger.Error(fmt.Sprintf("Could not get workflow file data from github, error: %s", err.Error()))
			return err
		}

		var dirFile DirFile
		err = json.Unmarshal(response, &dirFile)
		if err != nil {
			r.Config.Logger.Error(fmt.Sprintf("Could not Unmarshal json response for file data, error: %s", err.Error()))
			return err
		}

		dirFiles[i] = dirFile
	}

	r.Contents.WorkFlows = dirFiles

	return err
}

func (r *RestData) foundSecurityInsights(content DirContents) string {
	if strings.Contains(strings.ToLower(content.Name), "security-insights.") {
		response, err := r.getSourceFile(r.owner, r.repo, content.Path)
		if err != nil {
			r.Config.Logger.Error(fmt.Sprintf("error unmarshalling API response for security insights file: %s", err.Error()))
			return ""
		}
		r.Config.Logger.Trace(fmt.Sprintf("Security Insights Exists - SHA: %v", response.SHA))
		return content.Name
	}
	return ""
}

func (r *RestData) getTopDirContents() {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting top level contents: %s", err.Error()))
		return
	}
	_ = json.Unmarshal(responseData, &r.Contents.TopLevel)
}

func (r *RestData) getForgeDirContents() {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/.github", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting forge contents: %s", err.Error()))
		return
	}
	_ = json.Unmarshal(responseData, &r.Contents.ForgeDir)
}

func (r *RestData) getMetadata() error {
	endpoint := fmt.Sprintf("%s/repos/%s/%s", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		return err
	}
	return json.Unmarshal(responseData, &r)
}

func (r *RestData) getReleases() error {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/releases", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		return err
	}
	return json.Unmarshal(responseData, &r.Releases)
}

func (r *RestData) getWorkflow() error {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/actions/permissions/workflow", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		return err
	}
	//This is where we set the data in the restdata struct r.Workflow
	if err := json.Unmarshal(responseData, &r.Workflow); err != nil {
		return fmt.Errorf("failed to parse permissions: %v", err)
	}
	return err
}

func (r *RestData) loadOrgData() {
	endpoint := fmt.Sprintf("%s/orgs/%s", APIBase, r.owner)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting org data: %s (%s)", err.Error(), endpoint))
		return
	}
	_ = json.Unmarshal(responseData, &r.Organization)
}

func (r *RestData) GetRulesets(branchName string) []Ruleset {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/rules/branches/%s", APIBase, r.owner, r.repo, branchName)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting rulesets: %s", err.Error()))
	}

	_ = json.Unmarshal(responseData, &r.Rulesets)
	_ = json.Unmarshal(responseData, &r.Organization)
	return r.Rulesets
}
