package data

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPrivateVulnReporting(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		statusCode  int
		httpErr     error
		wantEnabled bool
		wantKnown   bool
	}{
		{
			name:        "enabled",
			body:        `{"enabled": true}`,
			wantEnabled: true,
			wantKnown:   true,
		},
		{
			name:        "disabled",
			body:        `{"enabled": false}`,
			wantEnabled: false,
			wantKnown:   true,
		},
		{
			name:        "not found leaves status unknown",
			body:        `{"message": "Not Found"}`,
			statusCode:  http.StatusNotFound,
			wantEnabled: false,
			wantKnown:   false,
		},
		{
			name:        "malformed body leaves status unknown",
			body:        `not json`,
			wantEnabled: false,
			wantKnown:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := NewPayloadWithHTTPMock(Payload{}, []byte(test.body), test.statusCode, test.httpErr)
			payload.getPrivateVulnReporting()

			assert.Equal(t, test.wantEnabled, payload.PrivateVulnReporting.Enabled)
			assert.Equal(t, test.wantKnown, payload.PrivateVulnReporting.Known)
		})
	}
}

func TestGetSecurityAdvisories(t *testing.T) {
	fullPage := "[" + strings.TrimSuffix(strings.Repeat(`{"state":"published"},`, 100), ",") + "]"
	tests := []struct {
		name         string
		body         string
		statusCode   int
		httpErr      error
		wantCount    int
		wantKnown    bool
		wantLowerBnd bool
	}{
		{
			name:      "single published advisory",
			body:      `[{"state":"published"}]`,
			wantCount: 1,
			wantKnown: true,
		},
		{
			name:      "only published advisories are counted",
			body:      `[{"state":"published"},{"state":"draft"},{"state":"published"}]`,
			wantCount: 2,
			wantKnown: true,
		},
		{
			name:      "no advisories is a known zero",
			body:      `[]`,
			wantCount: 0,
			wantKnown: true,
		},
		{
			name:         "full page marks count as lower bound",
			body:         fullPage,
			wantCount:    100,
			wantKnown:    true,
			wantLowerBnd: true,
		},
		{
			name:       "forbidden leaves status unknown",
			body:       `{"message": "Forbidden"}`,
			statusCode: http.StatusForbidden,
			wantCount:  0,
			wantKnown:  false,
		},
		{
			name:      "malformed body leaves status unknown",
			body:      `not json`,
			wantCount: 0,
			wantKnown: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := NewPayloadWithHTTPMock(Payload{}, []byte(test.body), test.statusCode, test.httpErr)
			payload.getSecurityAdvisories()

			assert.Equal(t, test.wantCount, payload.SecurityAdvisories.Count)
			assert.Equal(t, test.wantKnown, payload.SecurityAdvisories.Known)
			assert.Equal(t, test.wantLowerBnd, payload.SecurityAdvisories.CountIsLowerBound)
		})
	}
}
