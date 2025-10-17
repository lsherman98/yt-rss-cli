package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIClient struct {
	client  *http.Client
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (c *APIClient) do(method, path string, body io.Reader, v any) error {
	apiKey, err := GetApiKey()
	if err != nil {
		return fmt.Errorf("API key not set. Please run 'ytrss auth'")
	}

	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed: %s - %s", resp.Status, string(bodyBytes))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return err
		}
	}

	return nil
}

func (c *APIClient) download(method, path string, body io.Reader) (*http.Response, error) {
	apiKey, err := GetApiKey()
	if err != nil {
		return nil, fmt.Errorf("API key not set. Please run 'ytrss auth'")
	}

	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s - %s", resp.Status, string(bodyBytes))
	}

	return resp, nil
}
