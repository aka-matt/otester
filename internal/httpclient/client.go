package httpclient

import (
	"context"
	"net/http"
	"sync"
	"time"

	"otester/internal/model"
)

type Client struct {
	client      *http.Client
	cancelFuncs map[string]context.CancelFunc
	mu          sync.RWMutex
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cancelFuncs: make(map[string]context.CancelFunc),
	}
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

	if err != nil {
		return handleRequestError(input.RequestID, reqCtx.Err(), err, duration, req)
	}
	defer resp.Body.Close()

	return ParseResponse(resp, req, input.RequestID, duration)
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
