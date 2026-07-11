package httpclient

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"otester/internal/model"
)

func BuildRequest(ctx context.Context, input *model.RequestInput) (*http.Request, error) {
	// Build URL with query params
	baseURL, err := url.Parse(input.URL)
	if err != nil {
		return nil, err
	}

	query := baseURL.Query()
	for _, qp := range input.QueryParams {
		if qp.Enabled {
			query.Add(qp.Key, qp.Value)
		}
	}
	baseURL.RawQuery = query.Encode()

	// Build body
	var bodyReader *strings.Reader
	if input.BodyType != "none" && input.Body != "" {
		bodyReader = strings.NewReader(input.Body)
	} else {
		bodyReader = strings.NewReader("")
	}

	req, err := http.NewRequestWithContext(ctx, input.Method, baseURL.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for _, h := range input.Headers {
		if h.Enabled {
			req.Header.Add(h.Key, h.Value)
		}
	}

	// Set content type for body types
	switch input.BodyType {
	case "json":
		req.Header.Set("Content-Type", "application/json")
	case "text":
		req.Header.Set("Content-Type", "text/plain")
	case "x-www-form-urlencoded":
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}
