package apisix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/api7/adc-go/internal/config"
	"github.com/api7/adc-go/internal/retry"
)

type Client struct {
	baseURL     string
	adminKey    string
	keyName     string
	httpClient  *http.Client
	retryConfig *retry.Config
}

func NewClient(config *config.ADCConfig) *Client {
	return &Client{
		baseURL:  config.APISIX.BaseURL,
		adminKey: config.APISIX.AdminKey,
		keyName:  config.APISIX.AdminKeyName,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		retryConfig: retry.DefaultConfig(),
	}
}

// SetRetryConfig sets custom retry configuration
func (c *Client) SetRetryConfig(config *retry.Config) {
	c.retryConfig = config
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.adminKey != "" {
		req.Header.Set(c.keyName, c.adminKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) doJSONRequest(method, path string, reqBody, respBody interface{}) error {
	resp, err := c.doRequest(method, path, reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Routes
func (c *Client) GetRoutes() ([]map[string]interface{}, error) {
	return c.getResources("/apisix/admin/routes")
}

func (c *Client) CreateRoute(id string, route map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/routes/"+id, route, nil)
}

func (c *Client) UpdateRoute(id string, route map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/routes/"+id, route, nil)
}

func (c *Client) DeleteRoute(id string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/routes/"+id, nil, nil)
}

// Services
func (c *Client) GetServices() ([]map[string]interface{}, error) {
	return c.getResources("/apisix/admin/services")
}

func (c *Client) CreateService(id string, service map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/services/"+id, service, nil)
}

func (c *Client) UpdateService(id string, service map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/services/"+id, service, nil)
}

func (c *Client) DeleteService(id string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/services/"+id, nil, nil)
}

// Upstreams
func (c *Client) GetUpstreams() ([]map[string]interface{}, error) {
	return c.getResources("/apisix/admin/upstreams")
}

func (c *Client) CreateUpstream(id string, upstream map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/upstreams/"+id, upstream, nil)
}

func (c *Client) UpdateUpstream(id string, upstream map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/upstreams/"+id, upstream, nil)
}

func (c *Client) DeleteUpstream(id string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/upstreams/"+id, nil, nil)
}

// Consumers
func (c *Client) GetConsumers() ([]map[string]interface{}, error) {
	return c.getResources("/apisix/admin/consumers")
}

func (c *Client) CreateConsumer(username string, consumer map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/consumers/"+username, consumer, nil)
}

func (c *Client) UpdateConsumer(username string, consumer map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/consumers/"+username, consumer, nil)
}

func (c *Client) DeleteConsumer(username string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/consumers/"+username, nil, nil)
}

// SSLs
func (c *Client) GetSSLs() ([]map[string]interface{}, error) {
	return c.getResources("/apisix/admin/ssls")
}

func (c *Client) CreateSSL(id string, ssl map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/ssls/"+id, ssl, nil)
}

func (c *Client) UpdateSSL(id string, ssl map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/ssls/"+id, ssl, nil)
}

func (c *Client) DeleteSSL(id string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/ssls/"+id, nil, nil)
}

// Global Rules
func (c *Client) GetGlobalRules() ([]map[string]interface{}, error) {
	return c.getResources("/apisix/admin/global_rules")
}

func (c *Client) CreateGlobalRule(id string, rule map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/global_rules/"+id, rule, nil)
}

func (c *Client) UpdateGlobalRule(id string, rule map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/global_rules/"+id, rule, nil)
}

func (c *Client) DeleteGlobalRule(id string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/global_rules/"+id, nil, nil)
}

// Plugin Configs
func (c *Client) GetPluginConfigs() ([]map[string]interface{}, error) {
	return c.getResources("/apisix/admin/plugin_configs")
}

func (c *Client) CreatePluginConfig(id string, config map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/plugin_configs/"+id, config, nil)
}

func (c *Client) UpdatePluginConfig(id string, config map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/plugin_configs/"+id, config, nil)
}

func (c *Client) DeletePluginConfig(id string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/plugin_configs/"+id, nil, nil)
}

// Stream Routes
func (c *Client) GetStreamRoutes() ([]map[string]interface{}, error) {
	return c.getResources("/apisix/admin/stream_routes")
}

func (c *Client) CreateStreamRoute(id string, route map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/stream_routes/"+id, route, nil)
}

func (c *Client) UpdateStreamRoute(id string, route map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/stream_routes/"+id, route, nil)
}

func (c *Client) DeleteStreamRoute(id string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/stream_routes/"+id, nil, nil)
}

// Plugin Metadata
func (c *Client) GetPluginMetadata(pluginName string) (map[string]interface{}, error) {
	var response struct {
		Node struct {
			Value map[string]interface{} `json:"value"`
		} `json:"node"`
	}

	err := c.doJSONRequest("GET", "/apisix/admin/plugin_metadata/"+pluginName, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Node.Value, nil
}

func (c *Client) UpdatePluginMetadata(pluginName string, metadata map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/plugin_metadata/"+pluginName, metadata, nil)
}

func (c *Client) DeletePluginMetadata(pluginName string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/plugin_metadata/"+pluginName, nil, nil)
}

// Helper method to get resources list
func (c *Client) getResources(path string) ([]map[string]interface{}, error) {
	var response struct {
		List []struct {
			Value map[string]interface{} `json:"value"`
		} `json:"list"`
	}

	err := c.doJSONRequest("GET", path, nil, &response)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(response.List))
	for _, item := range response.List {
		result = append(result, item.Value)
	}

	return result, nil
}

// Ping checks connection to APISIX
func (c *Client) Ping() error {
	resp, err := c.doRequest("GET", "/apisix/admin/routes", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to APISIX: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("APISIX returned error status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
