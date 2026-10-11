package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type PaymentEvent struct {
	ID          string `json:"id"`
	CustomerID  string `json:"customer_id"`
	AmountCents int64  `json:"amount_cents"`
	Country     string `json:"country"`
}

type AuditNotification struct {
	EventID string `json:"event_id"`
	Action  string `json:"action"`
	Reason  string `json:"reason"`
	At      string `json:"at"`
}

type Decision struct {
	Approved bool              `json:"approved"`
	Path     string            `json:"path"`
	Audit    AuditNotification `json:"audit"`
}

const paymentRiskFlag = "kb_checkout_v2"

func decide(event PaymentEvent, enabled bool, now time.Time) Decision {
	path, reason := "legacy", "flag disabled"
	approved := event.AmountCents <= 100000
	if enabled {
		path, reason = "risk-v2", "risk policy evaluated"
		approved = event.AmountCents <= 50000 && event.Country != "blocked"
	}
	action := "declined"
	if approved {
		action = "approved"
	}
	return Decision{Approved: approved, Path: path, Audit: AuditNotification{EventID: event.ID, Action: action, Reason: reason, At: now.UTC().Format(time.RFC3339)}}
}

func authorizeHandler(flags *FlagsClient, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var event PaymentEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid payment event", http.StatusBadRequest)
			return
		}
		enabled, err := flags.IsEnabled(paymentRiskFlag)
		if err != nil {
			http.Error(w, "flag service unavailable", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(decide(event, enabled, now()))
	})
}
