package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/google/go-github/v71/github"
	"github.com/ossf/si-tooling/v2/si"
	"github.com/privateerproj/privateer-sdk/config"
)

type RestData struct {
	owner    string
	repo     string
	token    string
	Config   *config.Config
	Workflow Workflow
	Insights si.SecurityInsights
	Releases []ReleaseData
	Contents Contents
	Rulesets []Ruleset
	ghClient *github.Client
}

type Contents struct {
	TopLevel  []*github.RepositoryContent
	ForgeDir  []*github.RepositoryContent
	WorkFlows []DirFile
}

type Ruleset struct {
	Type       string `json:"type"`
	Parameters struct {
		RequiredChecks []struct {
			Context string `json:"context"`
		} `json:"required_status_checks"`
	} `json:"parameters"`
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

type Workflow struct {
	DefaultPermissions    string `json:"default_workflow_permissions"`
	CanApprovePullRequest bool   `json:"can_approve_pull_request_reviews"`
}

var APIBase = "https://api.github.com"

func (r *RestData) Setup() error {
	r.owner = r.Config.GetString("owner")
	r.repo = r.Config.GetString("repo")
	r.token = r.Config.GetString("token")

	r.getTopDirContents()
	r.getForgeDirContents()
	r.loadSecurityInsights()
	_ = r.getWorkflow()
	_ = r.getReleases()
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

func (r *RestData) getSourceFile(owner, repo, path string) (content *github.RepositoryContent, err error) {
	content, _, _, err = r.ghClient.Repositories.GetContents(context.Background(), owner, repo, path, nil)
	if err != nil {
		return
	}
	return content, nil
}

// checkFile accepts a filename like security-insights.yml or security.md and returns the path to that file
// if it exists in the root directory or forge directory of the repository or returns "" when the file is not found
func (r *RestData) checkFile(filename string) (filepath string) {
	filepath = ""
	for _, dirContents := range r.Contents.TopLevel {
		if strings.EqualFold(*dirContents.Name, filename) {
			filepath = *dirContents.Path
			break
		}
	}
	// prefer files found in the root directory
	if filepath != "" {
		return filepath
	}
	for _, dirContents := range r.Contents.ForgeDir {
		if strings.EqualFold(*dirContents.Name, filename) {
			filepath = *dirContents.Path
			break
		}
	}
	return filepath
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

// returns true when a file with case insensitive name matching support.md is found in the root or forge directories or when the readme.md contains a heading named "Support"
func (r *RestData) HasSupportMarkdown() bool {
	if r.checkFile("support.md") != "" {
		return true
	}
	readmePath := r.checkFile("readme.md")
	if readmePath != "" {
		contents, err := r.getSourceFile(r.owner, r.repo, readmePath)
		if err != nil {
			r.Config.Logger.Error(fmt.Sprintf("error getting readme contents: %s", err.Error()))
			return false
		}
		content, err := contents.GetContent()
		if err != nil {
			r.Config.Logger.Error(fmt.Sprintf("error getting readme contents: %s", err.Error()))
			return false
		}
		headings := parseMarkdownHeadings([]byte(content))
		for _, heading := range headings {
			if heading == "Support" {
				return true
			}
		}
	}
	return false
}

func parseMarkdownHeadings(content []byte) []string {
	var headings []string

	// Parse markdown into AST
	md := markdown.Parse(content, nil)

	// Walk the AST and collect headings
	ast.WalkFunc(md, func(node ast.Node, entering bool) ast.WalkStatus {
		if heading, ok := node.(*ast.Heading); ok && entering {
			// Get the text content of the heading
			if len(heading.Children) > 0 {
				if text, ok := heading.Children[0].(*ast.Text); ok {
					headings = append(headings, string(text.Literal))
				}
			}
		}
		return ast.GoToNext
	})

	return headings
}

func (r *RestData) loadSecurityInsights() {
	filepath := r.checkFile(si.SecurityInsightsFilename)
	if filepath != "" {
		insights, err := si.Read(r.owner, r.repo, filepath)
		r.Insights = insights
		if err != nil {
			r.Config.Logger.Error(fmt.Sprintf("error reading security insights file: %s", err.Error()))
		}
		return
	}
}

func (r *RestData) getTopDirContents() {
	_, contents, _, err := r.ghClient.Repositories.GetContents(context.Background(), r.owner, r.repo, "", nil)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting top level contents: %s", err.Error()))
		return
	}
	r.Contents.TopLevel = contents
	if len(r.Contents.TopLevel) == 0 {
		r.Config.Logger.Error("no contents retrieved from the top level of the repository")
		return
	}
}

func (r *RestData) getForgeDirContents() {
	_, contents, _, err := r.ghClient.Repositories.GetContents(context.Background(), r.owner, r.repo, ".github", nil)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting forge contents: %s", err.Error()))
		return
	}
	r.Contents.ForgeDir = contents
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

func (r *RestData) GetRulesets(branchName string) []Ruleset {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/rules/branches/%s", APIBase, r.owner, r.repo, branchName)
	responseData, err := r.MakeApiCall(endpoint, true)
	if err != nil {
		r.Config.Logger.Error(fmt.Sprintf("error getting rulesets: %s", err.Error()))
	}

	_ = json.Unmarshal(responseData, &r.Rulesets)
	return r.Rulesets
}
