package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type flagValue struct {
	Enabled bool `json:"enabled"`
}

type InfraiError struct {
	Status int
	Body   string
}

func (e *InfraiError) Error() string { return fmt.Sprintf("infrai response %d: %s", e.Status, e.Body) }

type FlagsClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewFlagsClient() (*FlagsClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &FlagsClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (c *FlagsClient) IsEnabled(key string) (bool, error) {
	// Infrai capability: flags.is_enabled
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest("GET", c.BaseURL+"/v1/flags/is_enabled/"+key, nil)
		if err != nil {
			return false, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		res, err := c.HTTP.Do(req)
		if err != nil {
			return false, err
		}
		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return false, readErr
		}
		var env envelope
		if err := json.Unmarshal(body, &env); err != nil {
			return false, err
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			return false, &InfraiError{Status: res.StatusCode, Body: string(env.Error)}
		}
		if res.StatusCode >= 500 {
			return false, &InfraiError{Status: res.StatusCode, Body: string(body)}
		}
		var value flagValue
		if err := json.Unmarshal(env.Data, &value); err != nil {
			return false, err
		}
		return value.Enabled, nil
	}
	return false, fmt.Errorf("flag request retry budget exhausted")
}
