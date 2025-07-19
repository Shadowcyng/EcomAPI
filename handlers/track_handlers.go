// api/handlers/analytics_handlers.go
package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"mabletask/api/models" // Your updated models package
	"mabletask/api/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // For generating EventID
)

type AnalyticsHandlers struct {
	AnalyticsStore *store.AnalyticsStore
}

func NewAnalyticsHandlers(s *store.AnalyticsStore) *AnalyticsHandlers {
	return &AnalyticsHandlers{
		AnalyticsStore: s,
	}
}

func (h *AnalyticsHandlers) TrackEvent(c *gin.Context) {
	// The frontend sends an array of AnalyticsEvent objects.
	var incomingEvents []models.AnalyticsEvent
	if err := c.ShouldBindJSON(&incomingEvents); err != nil {
		log.Printf("Error binding incoming analytics JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(incomingEvents) == 0 {
		c.Status(http.StatusOK)
		return
	}

	var eventsToInsert []models.AnalyticsEvent

	for _, event := range incomingEvents {
		event.EventID = uuid.New().String() // Generate a unique ID for this event record
		event.IPAddress = c.ClientIP()      // Capture IP address from the request context

		eventsToInsert = append(eventsToInsert, event)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second) // Set a timeout for DB operation
	defer cancel()

	if err := h.AnalyticsStore.InsertAnalyticsEvents(ctx, eventsToInsert); err != nil {
		log.Printf("Error inserting analytics events into ClickHouse: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record analytics events"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *AnalyticsHandlers) GetEventCountsOverTime(c *gin.Context) {
	interval := c.Query("interval")
	if interval == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interval query parameter is required (e.g., 'day', 'hour')"})
		return
	}

	// Optional eventType filter
	eventTypeFilter := c.Query("eventType") // Will be "" if not provided

	// Parse start and end times
	var start, end time.Time
	var err error

	startParam := c.Query("start")
	if startParam != "" {
		start, err = time.Parse(time.RFC3339, startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'start' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		start = time.Now().UTC().Add(-7 * 24 * time.Hour) // Default to 7 days ago if not provided
		log.Printf("No 'start' timestamp provided, defaulting to 7 days ago: %s", start.Format(time.RFC3339))
	}

	endParam := c.Query("end")
	if endParam != "" {
		end, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'end' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		end = time.Now().UTC() // Default to now
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	results, err := h.AnalyticsStore.GetEventCountsOverTime(ctx, interval, start, end, eventTypeFilter)
	if err != nil {
		log.Printf("Error getting event counts over time: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event statistics"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *AnalyticsHandlers) GetAverageEventDuration(c *gin.Context) {
	eventTypeFilter := c.Query("eventType") // Will be "" if not provided

	// Parse start and end times (re-use logic from GetEventCountsOverTime)
	var start, end time.Time
	var err error

	startParam := c.Query("start")
	if startParam != "" {
		start, err = time.Parse(time.RFC3339, startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'start' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		start = time.Now().UTC().Add(-7 * 24 * time.Hour) // Default to 7 days ago
	}

	endParam := c.Query("end")
	if endParam != "" {
		end, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'end' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		end = time.Now().UTC() // Default to now
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	avgDuration, err := h.AnalyticsStore.GetAverageEventDuration(ctx, eventTypeFilter, start, end)
	if err != nil {
		log.Printf("Error getting average event duration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve average event duration statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"eventType":         eventTypeFilter, // Echo back the filter used
		"startDate":         start.Format(time.RFC3339),
		"endDate":           end.Format(time.RFC3339),
		"averageDurationMs": avgDuration,
	})
}

func (h *AnalyticsHandlers) GetAverageCustomEventParameter(c *gin.Context) {
	eventTypeFilter := c.Query("eventType")
	paramName := c.Query("paramName")

	if eventTypeFilter == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventType query parameter is required"})
		return
	}
	if paramName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paramName query parameter is required (e.g., 'revenue', 'score')"})
		return
	}

	// Parse start and end times (re-use logic)
	var start, end time.Time
	var err error

	startParam := c.Query("start")
	if startParam != "" {
		start, err = time.Parse(time.RFC3339, startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'start' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		start = time.Now().UTC().Add(-7 * 24 * time.Hour) // Default to 7 days ago
	}

	endParam := c.Query("end")
	if endParam != "" {
		end, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'end' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		end = time.Now().UTC() // Default to now
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	avgValue, err := h.AnalyticsStore.GetAverageCustomEventParameter(ctx, eventTypeFilter, paramName, start, end)
	if err != nil {
		log.Printf("Error getting average of custom event parameter '%s' for eventType '%s': %v", paramName, eventTypeFilter, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve average custom event parameter statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"eventType":    eventTypeFilter,
		"paramName":    paramName,
		"startDate":    start.Format(time.RFC3339),
		"endDate":      end.Format(time.RFC3339),
		"averageValue": avgValue,
	})
}

func (h *AnalyticsHandlers) GetUniqueUsersOverTime(c *gin.Context) {
	interval := c.Query("interval")
	if interval == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interval query parameter is required (e.g., 'Day', 'Hour')"})
		return
	}

	var start, end time.Time
	var err error
	startParam := c.Query("start")
	if startParam != "" {
		start, err = time.Parse(time.RFC3339, startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'start' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		start = time.Now().UTC().Add(-7 * 24 * time.Hour)
	}

	endParam := c.Query("end")
	if endParam != "" {
		end, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'end' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		end = time.Now().UTC()
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	results, err := h.AnalyticsStore.GetUniqueUsersOverTime(ctx, interval, start, end)
	if err != nil {
		log.Printf("Error getting unique users over time: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve unique user statistics"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *AnalyticsHandlers) GetTopNPagePaths(c *gin.Context) {
	var start, end time.Time
	var err error

	// Parse start and end times (re-use logic)
	startParam := c.Query("start")
	if startParam != "" {
		start, err = time.Parse(time.RFC3339, startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'start' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		start = time.Now().UTC().Add(-7 * 24 * time.Hour) // Default to 7 days ago
	}

	endParam := c.Query("end")
	if endParam != "" {
		end, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'end' timestamp format. Use RFC3339 (e.g., 2006-01-02T15:04:05Z)"})
			return
		}
	} else {
		end = time.Now().UTC() // Default to now
	}

	var limit uint64 = 10 // Default limit
	limitParam := c.Query("limit")
	if limitParam != "" {
		parsedLimit, err := strconv.ParseUint(limitParam, 10, 64)
		if err != nil || parsedLimit == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'limit' parameter. Must be a positive integer."})
			return
		}
		limit = parsedLimit
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	results, err := h.AnalyticsStore.GetTopNPagePaths(ctx, start, end, limit)
	if err != nil {
		log.Printf("Error getting top page paths: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve top page paths statistics"})
		return
	}

	c.JSON(http.StatusOK, results)
}
