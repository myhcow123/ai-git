package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(port int) *APIClient {
	return &APIClient{
		baseURL: fmt.Sprintf("http://localhost:%d/api/v1", port),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *APIClient) IsAvailable() bool {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/health", c.baseURL))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (c *APIClient) Get(endpoint string) (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s%s", c.baseURL, endpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *APIClient) Post(endpoint string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s%s", c.baseURL, endpoint),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *APIClient) Init(path string) (map[string]interface{}, error) {
	return c.Post("/init", map[string]string{"path": path})
}

func (c *APIClient) Search(query string) (map[string]interface{}, error) {
	return c.Get(fmt.Sprintf("/search?q=%s", query))
}

func (c *APIClient) Read(symbol string) (map[string]interface{}, error) {
	return c.Get(fmt.Sprintf("/read/%s", symbol))
}

func (c *APIClient) Insert(target, code, position string) (map[string]interface{}, error) {
	return c.Post("/insert", map[string]interface{}{
		"target":   target,
		"code":     code,
		"position": position,
	})
}

func (c *APIClient) Replace(target, code string) (map[string]interface{}, error) {
	return c.Post("/replace", map[string]interface{}{
		"target": target,
		"code":   code,
	})
}

func (c *APIClient) Delete(target string) (map[string]interface{}, error) {
	return c.Post("/delete", map[string]string{"target": target})
}

func (c *APIClient) Status() (map[string]interface{}, error) {
	return c.Get("/status")
}

func getAPIClient() *APIClient {
	return NewAPIClient(8080)
}

func shouldUseAPI() bool {
	client := getAPIClient()
	return client.IsAvailable()
}

func ensureAPIRunning() error {
	if shouldUseAPI() {
		return nil
	}
	
	return fmt.Errorf("AI-Git daemon is not running. Please start it first:\n  ai-git daemon start")
}
