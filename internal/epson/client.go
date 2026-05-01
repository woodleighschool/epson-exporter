package epson

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"

	promconfig "github.com/prometheus/common/config"
)

const userAgent = "epson_exporter"

type Client struct {
	baseURL    *url.URL
	paths      Paths
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(target string, paths Paths, httpClientConfig promconfig.HTTPClientConfig, logger *slog.Logger) (*Client, error) {
	baseURL, err := url.Parse(strings.TrimRight(target, "/"))
	if err != nil {
		return nil, err
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("target must include scheme and host")
	}

	httpClient, err := promconfig.NewClientFromConfig(httpClientConfig, userAgent, promconfig.WithUserAgent(userAgent))
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Client{
		baseURL:    baseURL,
		paths:      paths,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

func (c *Client) Scrape(ctx context.Context) (Snapshot, error) {
	productStatus, err := c.getPage(ctx, c.paths.ProductStatus)
	if err != nil {
		return Snapshot{}, err
	}
	usageStatus, err := c.getPage(ctx, c.paths.UsageStatus)
	if err != nil {
		return Snapshot{}, err
	}
	networkStatus, err := c.getPage(ctx, c.paths.NetworkStatus)
	if err != nil {
		return Snapshot{}, err
	}
	hardwareStatus, err := c.getPage(ctx, c.paths.HardwareStatus)
	if err != nil {
		return Snapshot{}, err
	}

	return ParseSnapshot(productStatus, usageStatus, networkStatus, hardwareStatus)
}

func (c *Client) getPage(ctx context.Context, pagePath string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(pagePath).String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("GET %s returned %s", req.URL.String(), resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", req.URL.String(), err)
	}
	return data, nil
}

func (c *Client) endpoint(pagePath string) *url.URL {
	u := *c.baseURL
	u.Path = path.Join("/", strings.TrimSpace(pagePath))
	return &u
}
