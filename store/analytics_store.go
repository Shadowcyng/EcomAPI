// api/internal/store/analytics_store.go
package store

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"mabletask/api/database"
	"mabletask/api/models"
	"mabletask/api/utils"
)

type AnalyticsStore struct {
	DB *database.ClickHouseClient
}

type EventTypeCountByTime struct {
	Time      time.Time `json:"time"`
	EventType *string   `json:"eventType,omitempty"`
	Count     uint64    `json:"count"`
}

func NewAnalyticsStore(chClient *database.ClickHouseClient) *AnalyticsStore {
	return &AnalyticsStore{
		DB: chClient,
	}
}

func (s *AnalyticsStore) InsertAnalyticsEvents(ctx context.Context, events []models.AnalyticsEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Prepare a batch insert statement.
	// Ensure these column names and their order exactly match your ClickHouse table schema.
	batch, err := s.DB.Conn.PrepareBatch(ctx, `
		INSERT INTO analytics_events (
			event_id, event_type, user_id, session_id, timestamp, page_path, referrer, user_agent,
			ip_address, duration_ms, products, location, event_data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch insert: %w", err)
	}

	for _, event := range events {
		err := batch.Append(
			event.EventID,
			event.EventType,
			event.UserID,
			event.SessionID,
			event.Timestamp,
			event.PagePath,
			event.Referrer,
			event.UserAgent,
			event.IPAddress,
			event.DurationMs,
			event.Products,
			event.Location,
			event.EventData,
		)
		if err != nil {
			log.Printf("Error appending event to batch (EventID: %s): %v", event.EventID, err)
		}
	}

	err = batch.Send()
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	log.Printf("Successfully inserted %d analytics events.", len(events))
	return nil
}

func (s *AnalyticsStore) GetEventCountsOverTime(ctx context.Context, interval string, start, end time.Time, eventTypeFilter string) ([]EventTypeCountByTime, error) {
	var query string
	var args []interface{}
	args = append(args, start, end)

	if !utils.IsValidInterval(interval) {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	// Dynamically build SELECT, GROUP BY, and WHERE clauses
	selectCols := fmt.Sprintf("toStartOf%s(timestamp) as time_bucket, count() as total_events", interval)
	groupByCols := "time_bucket"
	whereClause := "WHERE timestamp >= ? AND timestamp <= ?"
	orderByCols := "time_bucket ASC"
	isFilteringByType := eventTypeFilter != ""

	if isFilteringByType {
		selectCols += ", event_type"
		groupByCols += ", event_type"
		whereClause += " AND event_type = ?"
		args = append(args, eventTypeFilter)
		orderByCols += ", event_type ASC"
	}

	query = fmt.Sprintf(`
		SELECT %s
		FROM analytics_events
		%s
		GROUP BY %s
		ORDER BY %s
	`, selectCols, whereClause, groupByCols, orderByCols)

	rows, err := s.DB.Conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query event counts over time: %w", err)
	}
	defer rows.Close()

	var results []EventTypeCountByTime
	for rows.Next() {
		var (
			timeBucket    time.Time
			count         uint64
			eventTypeDB   string
			currentResult EventTypeCountByTime
		)

		if isFilteringByType {
			// Scan into all three variables if filtering by type
			if err := rows.Scan(&timeBucket, &count, &eventTypeDB); err != nil {
				log.Printf("Error scanning row for event counts over time (with type filter): %v", err)
				continue
			}
			currentResult.EventType = &eventTypeDB // Assign pointer to string
		} else {
			// Scan only time and count if not filtering by type
			if err := rows.Scan(&timeBucket, &count); err != nil {
				log.Printf("Error scanning row for event counts over time (no type filter): %v", err)
				continue
			}
			currentResult.EventType = nil // Explicitly set to nil for total counts
		}

		currentResult.Time = timeBucket
		currentResult.Count = count
		results = append(results, currentResult)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error during event counts over time query: %w", err)
	}

	return results, nil
}

func (s *AnalyticsStore) GetAverageEventDuration(ctx context.Context, eventTypeFilter string, start, end time.Time) (float64, error) {
	var query string
	var args []interface{}

	// Base query to calculate average durationMs
	query = `SELECT avg(duration_ms) FROM analytics_events WHERE timestamp >= ? AND timestamp <= ?`
	args = append(args, start, end)

	if eventTypeFilter != "" {
		query += ` AND event_type = ?`
		args = append(args, eventTypeFilter)
	}

	var avgDuration float64
	err := s.DB.Conn.QueryRow(ctx, query, args...).Scan(&avgDuration)
	if err != nil {
		// Check if no rows were returned (e.g., no events found)
		if err.Error() == "sql: no rows in result set" { // Specific error for no rows
			return 0.0, nil // Return 0.0 for average if no events, no error
		}
		return 0.0, fmt.Errorf("failed to query average event duration: %w", err)
	}

	return avgDuration, nil
}

func (s *AnalyticsStore) GetAverageCustomEventParameter(ctx context.Context, eventTypeFilter, paramName string, start, end time.Time) (float64, error) {
	if paramName == "" {
		return 0.0, fmt.Errorf("parameter name for average calculation cannot be empty")
	}

	// Construct the query string dynamically to extract the specific parameter
	// We cast event_data to String for JSONExtractFloat, as required by ClickHouse for JSON columns.
	query := fmt.Sprintf(`
		SELECT avg(JSONExtractFloat(toString(event_data), '%s'))
		FROM analytics_events
		WHERE event_type = ? AND timestamp >= ? AND timestamp <= ?
	`, paramName)

	args := []interface{}{eventTypeFilter, start, end}

	var avgValue float64
	err := s.DB.Conn.QueryRow(ctx, query, args...).Scan(&avgValue)
	if err != nil {
		// If no rows are found, ClickHouse might return a "sql: no rows in result set" error.
		// In this case, we return 0.0 as the average.
		if err.Error() == "sql: no rows in result set" {
			return 0.0, nil
		}
		// For other errors, return a detailed error.
		return 0.0, fmt.Errorf("failed to query average of custom event parameter '%s': %w", paramName, err)
	}

	// ClickHouse's avg() function returns NaN if there are no matching rows,
	// which is not supported by standard JSON marshalling.
	// We check for NaN and convert it to 0.0 for consistent API responses.
	if math.IsNaN(avgValue) {
		return 0.0, nil
	}

	return avgValue, nil
}

func (s *AnalyticsStore) GetUniqueUsersOverTime(ctx context.Context, interval string, start, end time.Time) ([]EventTypeCountByTime, error) {
	if !utils.IsValidInterval(interval) {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	query := fmt.Sprintf(`
		SELECT toStartOf%s(timestamp) AS time_bucket, uniq(user_id) AS unique_users
		FROM analytics_events
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY time_bucket
		ORDER BY time_bucket ASC
	`, interval)

	rows, err := s.DB.Conn.Query(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query unique users over time: %w", err)
	}
	defer rows.Close()

	var results []EventTypeCountByTime
	for rows.Next() {
		var timeBucket time.Time
		var uniqueUsers uint64
		if err := rows.Scan(&timeBucket, &uniqueUsers); err != nil {
			log.Printf("Error scanning row for unique users: %v", err)
			continue
		}
		results = append(results, EventTypeCountByTime{
			Time:  timeBucket,
			Count: uniqueUsers,
			// EventType will be nil/omitted, as this is a total count
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows for unique users: %w", err)
	}

	return results, nil
}

func (s *AnalyticsStore) GetTopNPagePaths(ctx context.Context, start, end time.Time, limit uint64) ([]models.TopPathResult, error) {
	if limit == 0 {
		limit = 10 // Default limit if 0 is passed
	}

	query := `
		SELECT page_path, count() as view_count
		FROM analytics_events
		WHERE event_type = 'page_view' AND timestamp >= ? AND timestamp <= ?
		GROUP BY page_path
		ORDER BY view_count DESC
		LIMIT ?
	`
	rows, err := s.DB.Conn.Query(ctx, query, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top page paths: %w", err)
	}
	defer rows.Close()

	var results []models.TopPathResult
	for rows.Next() {
		var pagePath string
		var count uint64
		if err := rows.Scan(&pagePath, &count); err != nil {
			log.Printf("Error scanning row for top page paths: %v", err)
			continue
		}
		results = append(results, models.TopPathResult{
			PagePath: pagePath,
			Count:    count,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows for top page paths: %w", err)
	}

	return results, nil
}
