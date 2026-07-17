package httpclient

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"otester/internal/logbus"
	"otester/internal/model"
	"otester/internal/security"
)

type Client struct {
	client           *http.Client
	cancelFuncs      map[string]context.CancelFunc
	mu               sync.RWMutex
	allowInsecureTLS bool
	logger           logbus.Logger
}

func NewClient(logger logbus.Logger) *Client {
	if logger == nil {
		logger = logbus.Nop()
	}
	return &Client{
		client:      &http.Client{Timeout: 30 * time.Second},
		cancelFuncs: make(map[string]context.CancelFunc),
		logger:      logger,
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
	timeout := time.Duration(input.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
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

	c.logger.Infof("→ %s %s", req.Method, security.RedactURL(req.URL.String()))
	for k, v := range security.RedactHeaders(req.Header) {
		c.logger.Debugf("  header %s: %s", k, strings.Join(v, ", "))
	}
	// Log body details so POST/PUT/PATCH payload problems are visible in
	// Open Logs. Body content is redacted for secrets; empty bodies are
	// noted explicitly so a missing body is not confused with a quiet log.
	if input.BodyType != "" && input.BodyType != "none" {
		c.logger.Infof("  body type=%s content-length=%d", input.BodyType, req.ContentLength)
		if input.Body != "" {
			c.logger.Debugf("  body: %s", security.RedactJSON(input.Body))
		} else {
			c.logger.Warnf("  body is empty (type=%s)", input.BodyType)
		}
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err == nil {
		defer resp.Body.Close()
		c.logger.Infof("← %d %s (%d ms)", resp.StatusCode, resp.Status, duration.Milliseconds())
		return ParseResponse(resp, req, input.RequestID, duration)
	}

	// First attempt failed. Classify: is this a TLS error and is the insecure flag on?
	if c.getAllowInsecureTLS() && isTLSError(err) {
		c.logger.Warnf("TLS validation failed, retrying with insecure TLS: %s", err.Error())
		retryClient := InsecureTLSClient(timeout)
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
			c.logger.Infof("← %d %s (%d ms, insecure TLS)", retryResp.StatusCode, retryResp.Status, totalDuration.Milliseconds())
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
		c.logger.Errorf("request failed: %v", retryErr)
		return handleRequestError(input.RequestID, reqCtx.Err(), retryErr, duration+retryDuration, req)
	}

	c.logger.Errorf("request failed: %v", err)
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
