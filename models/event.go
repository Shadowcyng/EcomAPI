// api/internal/models/analytics.go
package models

import (
	"encoding/json"
	"time"
)

// AnalyticsEvent represents a single analytics event.
type AnalyticsEvent struct {
	EventID    string          `json:"eventId"`
	EventType  string          `json:"eventType"`
	UserID     string          `json:"userId"`
	SessionID  string          `json:"sessionId"`
	Timestamp  time.Time       `json:"timestamp"`
	PagePath   string          `json:"pagePath"`
	Referrer   string          `json:"referrer"`
	UserAgent  string          `json:"userAgent"`
	IPAddress  string          `json:"ipAddress"`
	DurationMs int64           `json:"durationMs"`
	Products   json.RawMessage `json:"products,omitempty"`
	Location   string          `json:"location,omitempty"`
	EventData  json.RawMessage `json:"eventData,omitempty"`
}

type TopPathResult struct {
	PagePath string `json:"pagePath"`
	Count    uint64 `json:"count"`
}
