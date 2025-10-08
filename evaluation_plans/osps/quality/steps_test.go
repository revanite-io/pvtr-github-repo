package quality

import (
	"testing"

	"github.com/ossf/gemara/layer4"
	"github.com/ossf/si-tooling/v2/si"
	"github.com/revanite-io/pvtr-github-repo/data"
)

func Test_InsightsListsRepositories(t *testing.T) {
	tests := []struct {
		name       string
		payload    data.Payload
		wantResult layer4.Result
		wantMsg    string
	}{
		{
			name: "insights contains repositories",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: si.Project{
							Repositories: []si.Repo{
								si.Repo{
									URL: "https://github.com/org/repo",
								},
							},
						},
					},
				},
			},
			wantResult: layer4.Passed,
			wantMsg:    "Insights contains a list of repositories",
		},
		{
			name: "insights does not contain repositories",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{
						Project: si.Project{
							Repositories: []si.Repo{},
						},
					},
				},
			},
			wantResult: layer4.Failed,
			wantMsg:    "Insights does not contain a list of repositories",
		},
		{
			name: "insights is nil",
			payload: data.Payload{
				RestData: &data.RestData{
					Insights: si.SecurityInsights{},
				},
			},
			wantResult: layer4.Failed,
			wantMsg:    "Insights does not contain a list of repositories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotMsg := InsightsListsRepositories(tt.payload)
			if gotResult != tt.wantResult {
				t.Errorf("result = %v, want %v", gotResult, tt.wantResult)
			}
			if gotMsg != tt.wantMsg {
				t.Errorf("message = %q, want %q", gotMsg, tt.wantMsg)
			}
		})
	}
}
