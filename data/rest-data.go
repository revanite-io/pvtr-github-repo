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
		TopLevel []DirContents
		ForgeDir []DirContents
		WorkFlows []DirFile
	}
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
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
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

	r.getMetadata()
	r.loadSecurityInsights()
	r.getWorkflow()
	r.getReleases()
	r.loadOrgData()
	r.getWorkflowFiles()
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
		if r.foundSecurityInsights(dirContents) {
			insights, err := si.Read(r.owner, r.repo, "security-insights.yml")
			r.Insights = insights
			if err != nil {
				r.Config.Logger.Error(fmt.Sprintf("error reading security insights file: %s", err.Error()))
			}
			return
		}
	}
	r.getForgeDirContents()
	for _, dirContents := range r.Contents.ForgeDir {
		if r.foundSecurityInsights(dirContents) {
			insights, err := si.Read(r.owner, r.repo, ".github/security-insights.yml")
			r.Insights = insights
			if err != nil {
				r.Config.Logger.Error(fmt.Sprintf("error reading security insights file: %s", err.Error()))
			}
			return
		}
	}
}

func (r *RestData) getWorkflowFiles() {
	//Only get top level for now, will need to recurse later
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/.github/workflows", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting workflows: %s", err.Error()))
		return
	}
	
	var workflowFileList []DirContents
	json.Unmarshal(responseData, &workflowFileList)


	//For each file, listed we need to get it and put it in a format the action parser can use
	var dirFiles = make([]DirFile, len(workflowFileList))
	for i, workflowFile := range workflowFileList {

		//TODO: log error
		response , _:= r.MakeApiCall(workflowFile.URL, true)

		var dirFile DirFile
		err := json.Unmarshal(response, &dirFile)
		if( err != nil ){
			//TODO do some logging
		}
		
		dirFiles[i] = dirFile;
	}

	r.Contents.WorkFlows = dirFiles

	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error unmarshalling API response for workflows: %s", err.Error()))
		return
	}
	// r.Contents.WorkFlows = workflows

}

func (r *RestData) foundSecurityInsights(content DirContents) bool {
	if strings.Contains(strings.ToLower(content.Name), "security-insights.") {
		response, err := r.getSourceFile(r.owner, r.repo, content.Path)
		if err != nil {
			r.Config.Logger.Error(fmt.Sprintf("error unmarshalling API response for security insights file: %s", err.Error()))
			return false
		}
		r.Config.Logger.Trace(fmt.Sprintf("Security Insights Exists - SHA: %v", response.SHA))
		return true
	}
	return false
}

func (r *RestData) getTopDirContents() {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting top level contents: %s", err.Error()))
		return
	}
	json.Unmarshal(responseData, &r.Contents.TopLevel)
}

func (r *RestData) getForgeDirContents() {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/.github", APIBase, r.owner, r.repo)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting forge contents: %s", err.Error()))
		return
	}
	json.Unmarshal(responseData, &r.Contents.ForgeDir)
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
	if err := json.Unmarshal(responseData, &r.Workflow); err != nil {
		return fmt.Errorf("failed to parse permissions: %v", err)
	}
	return nil
}


func (r *RestData) loadOrgData() {
	endpoint := fmt.Sprintf("%s/orgs/%s", APIBase, r.owner)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting org data: %s (%s)", err.Error(), endpoint))
		return
	}
	json.Unmarshal(responseData, &r.Organization)

	return
}
