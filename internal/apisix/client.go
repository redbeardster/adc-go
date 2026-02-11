package apisix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/api7/adc-go/internal/config"
)

type Client struct {
	baseURL    string
	adminKey   string
	keyName    string
	httpClient *http.Client
}

func NewClient(config *config.ADCConfig) *Client {
	return &Client{
		baseURL:  config.APISIX.BaseURL,
		adminKey: config.APISIX.AdminKey,
		keyName:  config.APISIX.AdminKeyName,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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

// Методы для работы с ресурсами
func (c *Client) GetRoutes() (map[string]interface{}, error) {
	var response struct {
		Node struct {
			Nodes []struct {
				Value map[string]interface{} `json:"value"`
			} `json:"nodes"`
		} `json:"node"`
	}

	err := c.doJSONRequest("GET", "/apisix/admin/routes", nil, &response)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, node := range response.Node.Nodes {
		if value, ok := node.Value["value"].(map[string]interface{}); ok {
			if id, ok := value["id"].(string); ok {
				result[id] = value
			}
		}
	}

	return result, nil
}

func (c *Client) CreateRoute(route map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/routes/"+route["id"].(string), route, nil)
}

func (c *Client) UpdateRoute(id string, route map[string]interface{}) error {
	return c.doJSONRequest("PUT", "/apisix/admin/routes/"+id, route, nil)
}

func (c *Client) DeleteRoute(id string) error {
	return c.doJSONRequest("DELETE", "/apisix/admin/routes/"+id, nil, nil)
}

// Аналогичные методы для других ресурсов (Services, Upstreams, Consumers, SSL и т.д.)
// ...
