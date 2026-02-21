package data

import (
	"bytes"
	"io"
	"net/http"
)

type ClientMock struct {
	Response *http.Response
	Err      error
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
    return c.Response, c.Err
}

func NewPayloadWithHTTPMock(base Payload, body []byte, statusCode int, httpErr error) Payload {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	mock := &ClientMock{
		Response: &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewReader(body)),
		},
		Err: httpErr,
	}
	if base.RestData == nil {
		base.RestData = &RestData{}
	}
	base.ensureInsightsInitialized()
	base.HttpClient = mock
	return base
}
