package httpclient

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"otester/internal/model"
)

type Client struct {
	client           *http.Client
	cancelFuncs      map[string]context.CancelFunc
	mu               sync.RWMutex
	allowInsecureTLS bool
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cancelFuncs: make(map[string]context.CancelFunc),
	}
}

// SetAllowInsecureTLS toggles the fallback behaviour: when a TLS handshake
// fails because of certificate validation, the next request will be retried
// once with InsecureSkipVerify=true. The flag is read on every DoRequest call,
// so callers can flip it at runtime without restarting the client.
func (c *Client) SetAllowInsecureTLS(enabled bool) {
	c.mu.Lock()
	c.allowInsecureTLS = enabled
	c.mu.Unlock()
}

func (c *Client) getAllowInsecureTLS() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allowInsecureTLS
}

func (c *Client) DoRequest(ctx context.Context, input *model.RequestInput) (*model.ResponseOutput, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(input.TimeoutSeconds)*time.Second)
	defer cancel()

	c.mu.Lock()
	c.cancelFuncs[input.RequestID] = cancel
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.cancelFuncs, input.RequestID)
		c.mu.Unlock()
	}()

	req, err := BuildRequest(reqCtx, input)
	if err != nil {
		return &model.ResponseOutput{
			RequestID:    input.RequestID,
			ErrorCode:    string(model.ErrInvalidURL),
			ErrorMessage: err.Error(),
		}, nil
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err == nil {
		defer resp.Body.Close()
		return ParseResponse(resp, req, input.RequestID, duration)
	}

	// First attempt failed. Classify: is this a TLS error and is the insecure flag on?
	if c.getAllowInsecureTLS() && isTLSError(err) {
		retryClient := InsecureTLSClient(time.Duration(input.TimeoutSeconds) * time.Second)
		retryStart := time.Now()
		// Clone makes a shallow copy of Body. The first attempt's Do has already
		// consumed the underlying reader, so re-using it would silently send an
		// empty body for non-GET methods. Re-populate Body from GetBody() (set by
		// http.NewRequestWithContext for *strings.Reader / *bytes.Reader /
		// *bytes.Buffer) so the retry transmits the original payload.
		retryReq := req.Clone(reqCtx)
		if retryReq.GetBody != nil {
			if body, bodyErr := retryReq.GetBody(); bodyErr == nil {
				retryReq.Body = body
			}
		}
		retryResp, retryErr := retryClient.Do(retryReq)
		retryDuration := time.Since(retryStart)

		if retryErr == nil {
			defer retryResp.Body.Close()
			totalDuration := duration + retryDuration
			output, parseErr := ParseResponse(retryResp, req, input.RequestID, totalDuration)
			if parseErr != nil {
				return nil, parseErr
			}
			if output.TLS != nil {
				output.TLS.ValidationSkipped = true
				output.TLS.OriginalError = err.Error()
			}
			return output, nil
		}

		// Retry also failed — fall through to handleRequestError with the second error.
		return handleRequestError(input.RequestID, reqCtx.Err(), retryErr, duration+retryDuration, req)
	}

	return handleRequestError(input.RequestID, reqCtx.Err(), err, duration, req)
}

// isTLSError returns true when err's message contains "tls:" or "x509:",
// matching the same heuristic as BuildTLSInfoFromError. Context timeouts and
// cancellations are never TLS handshake failures and are excluded here so the
// insecure-TLS retry is not triggered for them.
func isTLSError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "tls:") || strings.Contains(msg, "x509:")
}

func (c *Client) CancelRequest(requestID string) error {
	c.mu.RLock()
	cancel, ok := c.cancelFuncs[requestID]
	c.mu.RUnlock()

	if ok {
		cancel()
		return nil
	}
	return nil // already completed or not found
}
