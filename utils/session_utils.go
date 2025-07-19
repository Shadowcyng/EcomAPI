package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"time"
)

// sessions is a simple in-memory map to store session IDs mapped to user IDs.
// THIS IS FOR DEMO PURPOSES ONLY AND IS NOT SUITABLE FOR PRODUCTION.
// In production, use a persistent store like Redis or a database.

var sessions = make(map[string]string) // sessionID -> userID

// GenerateSessionID creates a simple, unique (for demo) session ID.
func GenerateSessionID() string {
	b := make([]byte, 32)                   
	if _, err := rand.Read(b); err != nil {
		log.Printf("ERROR: Failed to generate random bytes for session ID: %v", err)
		return "fallback_session_id_" + time.Now().Format("20060102150405") // Fallback for demo
	}
	// Encode to Base64 to make it a URL-safe string
	return base64.URLEncoding.EncodeToString(b)
}

// StoreSession stores a session ID and its associated user ID in our in-memory map.
func StoreSession(sessionID, userID string) {
	sessions[sessionID] = userID
	log.Printf("Session stored: %s for user: %s (In-memory, NOT PRODUCTION)", sessionID, userID)
}

// GetUserIDFromSession retrieves a user ID given a session ID.
func GetUserIDFromSession(sessionID string) (string, bool) {
	userID, ok := sessions[sessionID]
	return userID, ok
}

// DeleteSession removes a session ID from our in-memory map.
func DeleteSession(sessionID string) {
	delete(sessions, sessionID)
	log.Printf("Session deleted: %s (In-memory, NOT PRODUCTION)", sessionID)
}
