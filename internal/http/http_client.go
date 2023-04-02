package http

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"math"
	netHttp "net/http"
	"time"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
}

//go:generate mockgen -destination=mocks/mock_http_client.go -source=http_client.go -package=mocks Client
type Client interface {
	Do(req *netHttp.Request) (*netHttp.Response, error)
}

type RetryableClient struct {
	HttpClient Client
	Retries    int
	Backoff    float64
}

type RequestOption func(*netHttp.Request) error

// Post - http client abstraction for making http requests. includes retries and timeouts with exponential backoff
func (client *RetryableClient) Post(url string, options ...RequestOption) error {
	var err error
	req, err := netHttp.NewRequest(netHttp.MethodPost, url, nil)
	if err != nil {
		return err
	}

	for _, option := range options {
		err = option(req)
		if err != nil {
			return err
		}
	}

	// +1 for the first try
	maxTries := client.Retries + 1
	var resp *netHttp.Response
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	for i := 0; i < maxTries; i++ {
		resp, err = client.HttpClient.Do(req)
		if err == nil {
			break
		}

		sleepTime := client.Backoff * math.Pow(2, float64(i))
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("failed to send message after %v attempts", maxTries))
		return err
	}

	if resp.StatusCode >= 400 {
		logger.Error(fmt.Sprintf("request failed with status code %v", resp.StatusCode))
		return errors.New("error occurred sending email")
	}

	logger.Debug(fmt.Sprintf("request succeeded: %v", resp.StatusCode))
	return nil
}
