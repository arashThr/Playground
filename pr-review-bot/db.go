package main

import (
	"database/sql"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS pr_reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pr_url TEXT NOT NULL,
    description TEXT,
    channel_id TEXT NOT NULL,
    message_ts TEXT NOT NULL,
    reviewers TEXT NOT NULL,  -- Comma-separated user IDs
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP,
    approved_by TEXT
);`

type PRReview struct {
	ID          int64
	PRUrl       string
	Description string
	ChannelID   string
	// message_ts (timestamp) serves as the unique identifier for a message
	MessageTS  string
	Reviewers  []string
	Status     string
	CreatedAt  time.Time
	ApprovedAt sql.NullTime
	ApprovedBy sql.NullString
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./pr_reviews.db")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, err
	}

	return db, nil
}

func storePRReview(db *sql.DB, pr *PRReview) error {
	query := `
        INSERT INTO pr_reviews (
            pr_url, description, channel_id, message_ts, reviewers, status
        ) VALUES (?, ?, ?, ?, ?, ?)`

	reviewers := strings.Join(pr.Reviewers, ",")

	_, err := db.Exec(query,
		pr.PRUrl,
		pr.Description,
		pr.ChannelID,
		pr.MessageTS,
		reviewers,
		"pending",
	)

	return err
}

func updatePRStatus(db *sql.DB, messageTS string, status string, approvedBy string) error {
	query := `
        UPDATE pr_reviews 
        SET status = ?, 
            approved_at = CASE WHEN ? = 'approved' THEN CURRENT_TIMESTAMP ELSE NULL END,
            approved_by = CASE WHEN ? = 'approved' THEN ? ELSE NULL END
        WHERE message_ts = ?`

	_, err := db.Exec(query, status, status, status, approvedBy, messageTS)
	return err
}

// Add this new database function
func addReviewer(db *sql.DB, messageTS string, reviewerID string) error {
	// First get current reviewers
	var reviewersStr string
	err := db.QueryRow("SELECT reviewers FROM pr_reviews WHERE message_ts = ?", messageTS).Scan(&reviewersStr)
	if err != nil {
		return err
	}

	// Convert to slice
	reviewers := strings.Split(reviewersStr, ",")

	// Check if reviewer already exists
	for _, r := range reviewers {
		if r == reviewerID {
			return nil // Already a reviewer
		}
	}

	// Add new reviewer
	reviewers = append(reviewers, reviewerID)
	newReviewersStr := strings.Join(reviewers, ",")

	// Update database
	_, err = db.Exec("UPDATE pr_reviews SET reviewers = ? WHERE message_ts = ?",
		newReviewersStr, messageTS)
	return err
}

// Add these database functions
func getPendingPRs(db *sql.DB) ([]PRReview, error) {
	query := `
        SELECT id, pr_url, description, channel_id, message_ts, reviewers, created_at 
        FROM pr_reviews 
        WHERE status = 'pending'`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []PRReview
	for rows.Next() {
		var pr PRReview
		var reviewersStr string
		err := rows.Scan(
			&pr.ID,
			&pr.PRUrl,
			&pr.Description,
			&pr.ChannelID,
			&pr.MessageTS,
			&reviewersStr,
			&pr.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		pr.Reviewers = strings.Split(reviewersStr, ",")
		prs = append(prs, pr)
	}
	return prs, nil
}
