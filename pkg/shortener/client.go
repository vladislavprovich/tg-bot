package shortener

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

const (
	shorten = "shorten"
	stats   = "stats"
)

type BasicClient struct {
	config     *Config
	httpClient *http.Client
	logger     *logrus.Logger
}

type Client interface {
	CreateShortURL(ctx context.Context, req *CreateShortURLRequest) (*CreateShortURLResponse, error)
	GetStatsURL(ctx context.Context, req *GetShortURLStatsRequest) (*GetShortURLStatsResponse, error)
}

func NewBasicClient(config *Config, httpClient *http.Client, logger *logrus.Logger) BasicClient {
	return BasicClient{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (c *BasicClient) CreateShortURL(ctx context.Context, req *CreateShortURLRequest) (*CreateShortURLResponse, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		c.logger.Errorf("error marshalling request: %s", err)
		return nil, errors.New("service error")
	}

	urlForCreateShortURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", c.config.BaseURL, c.config.Port),
		Path:   shorten,
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		urlForCreateShortURL.String(),
		bytes.NewBuffer(jsonReq))
	if err != nil {
		c.logger.Errorf("failed to create request: %v", err)
		return nil, errors.New("service error")
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Errorf("http request failed: %v", err)
		return nil, errors.New("service error")
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("bad status code: %d", resp.StatusCode)
		return nil, errors.New("service error")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorf("error reading body: %v", err)
		return nil, errors.New("service error")
	}

	var res CreateShortURLResponse
	if err = json.Unmarshal(body, &res); err != nil {
		c.logger.Errorf("bad response body: %s", string(body))
		return nil, errors.New("service error")
	}

	return &res, nil
}

func (c *BasicClient) GetStatsURL(ctx context.Context,
	req *GetShortURLStatsRequest) (*GetShortURLStatsResponse, error) {
	urlForGetStatus := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", c.config.BaseURL, c.config.Port),
		Path:   fmt.Sprintf("%s/%s", req.ShortURL, stats),
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, urlForGetStatus.String(), nil)
	if err != nil {
		c.logger.Errorf("failed to create request: %v", err)
		return nil, errors.New("service error")
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Errorf("http request failed: %v", err)
		return nil, errors.New("service error")
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("bad status code: %d", resp.StatusCode)
		return nil, errors.New("service error")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorf("error reading body: %v", err)
		return nil, errors.New("service error")
	}

	var res GetShortURLStatsResponse
	if err = json.Unmarshal(body, &res); err != nil {
		c.logger.Errorf("bad response body: %s", string(body))
		return nil, errors.New("service error")
	}

	return &res, nil
}
