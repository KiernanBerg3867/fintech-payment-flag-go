package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDecideRiskFlag(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		amount   int64
		country  string
		approved bool
		path     string
	}{
		{"legacy threshold", false, 90000, "US", true, "legacy"},
		{"new policy declines large payment", true, 90000, "US", false, "risk-v2"},
		{"new policy blocks country", true, 1000, "blocked", false, "risk-v2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decide(PaymentEvent{ID: "evt-1", AmountCents: tt.amount, Country: tt.country}, tt.enabled, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
			if got.Approved != tt.approved || got.Path != tt.path || got.Audit.Action == "" {
				t.Fatalf("decision=%+v", got)
			}
		})
	}
}

func TestIsEnabledDecodesFlagValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/flags/is_enabled/"+paymentRiskFlag {
			t.Fatalf("path=%q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{"key":"kb_checkout_v2","enabled":true,"value":true}}`))
	}))
	defer server.Close()

	client := &FlagsClient{BaseURL: server.URL, Key: "test-key", HTTP: server.Client()}
	enabled, err := client.IsEnabled(paymentRiskFlag)
	if err != nil {
		t.Fatal(err)
	}
	if !enabled {
		t.Fatal("flag should be enabled")
	}
}
