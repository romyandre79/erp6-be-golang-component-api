package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Input struct {
	Params []struct {
		InputName string `json:"inputname"`
		CompValue string `json:"compvalue"`
	} `json:"params"`
}

type Output struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

func main() {
	var input Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to decode input: %v", err)})
		return
	}

	var (
		url         string
		method      string
		headers     string
		body        string
		timeout     = 30
		contentType = "application/json"
	)

	for _, p := range input.Params {
		switch p.InputName {
		case "url":
			url = strings.TrimSpace(p.CompValue)
		case "method":
			method = strings.ToUpper(strings.TrimSpace(p.CompValue))
		case "headers":
			headers = strings.TrimSpace(p.CompValue)
		case "body":
			body = strings.TrimSpace(p.CompValue)
		case "timeout":
			fmt.Sscanf(p.CompValue, "%d", &timeout)
		case "contenttype":
			if ct := strings.TrimSpace(p.CompValue); ct != "" {
				contentType = ct
			}
		}
	}

	if url == "" {
		json.NewEncoder(os.Stdout).Encode(Output{Error: "url parameter is required"})
		return
	}

	if method == "" {
		method = "GET"
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Create request
	var req *http.Request
	var err error

	if method == "GET" || method == "DELETE" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, bytes.NewBufferString(body))
	}

	if err != nil {
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to create request: %v", err)})
		return
	}

	// Set content type
	if method == "POST" || method == "PUT" || method == "PATCH" {
		req.Header.Set("Content-Type", contentType)
	}

	// Parse and set custom headers
	if headers != "" {
		headerPairs := strings.Split(headers, ",")
		for _, pair := range headerPairs {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				req.Header.Set(key, value)
			}
		}
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("request failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to read response: %v", err)})
		return
	}

	// Try to parse as JSON
	var jsonResult interface{}
	if err := json.Unmarshal(respBody, &jsonResult); err == nil {
		// Valid JSON response
		result := map[string]interface{}{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
			"headers":     resp.Header,
			"body":        jsonResult,
		}
		json.NewEncoder(os.Stdout).Encode(Output{Result: result})
	} else {
		// Plain text response
		result := map[string]interface{}{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
			"headers":     resp.Header,
			"body":        string(respBody),
		}
		json.NewEncoder(os.Stdout).Encode(Output{Result: result})
	}
}
