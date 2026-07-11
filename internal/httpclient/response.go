package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"

	"otester/internal/model"
)

const maxBodySize = 10 * 1024 * 1024 // 10MB

func ParseResponse(resp *http.Response, requestID string, duration time.Duration) (*model.ResponseOutput, error) {
	bodyBytes, truncated, err := readBody(resp.Body)
	if err != nil {
		return &model.ResponseOutput{
			RequestID:    requestID,
			StatusCode:   resp.StatusCode,
			Status:       resp.Status,
			ErrorCode:    string(model.ErrAPIRequestFailed),
			ErrorMessage: err.Error(),
			DurationMs:   duration.Milliseconds(),
		}, nil
	}

	contentType := resp.Header.Get("Content-Type")

	return &model.ResponseOutput{
		RequestID:     requestID,
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       resp.Header,
		Body:          string(bodyBytes),
		BodyTruncated: truncated,
		DurationMs:    duration.Milliseconds(),
		SizeBytes:     int64(len(bodyBytes)),
		ContentType:   contentType,
	}, nil
}

func readBody(body io.Reader) ([]byte, bool, error) {
	reader := io.LimitReader(body, maxBodySize+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, false, err
	}

	truncated := len(data) > maxBodySize
	if truncated {
		data = data[:maxBodySize]
	}

	return data, truncated, nil
}

func handleRequestError(requestID string, ctxErr error, err error, duration time.Duration) (*model.ResponseOutput, error) {
	output := &model.ResponseOutput{
		RequestID:    requestID,
		DurationMs:   duration.Milliseconds(),
	}

	switch ctxErr {
	case context.DeadlineExceeded:
		output.ErrorCode = string(model.ErrRequestTimeout)
		output.ErrorMessage = "request timed out"
	case context.Canceled:
		output.ErrorCode = string(model.ErrRequestCancelled)
		output.ErrorMessage = "request was cancelled"
	default:
		output.ErrorCode = string(model.ErrConnectionFailed)
		output.ErrorMessage = err.Error()
	}

	return output, nil
}
