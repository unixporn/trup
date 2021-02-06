package db

import (
	"context"
	"time"
)

type ShowcaseEntry struct {
	MessageID  string    `json:"message_id"`
	UserID     string    `json:"user_id"`
	Score      int       `json:"score"`
	CreateDate time.Time `json:"create_date"`
}

func AddShowcaseEntries(entries []ShowcaseEntry) error {
	_, err := db.Exec(context.Background(), `
	INSERT INTO showcase_entries (
		message_id,
		user_id,
		score,
		create_date
	)
	SELECT
		*
	FROM
		jsonb_to_recordset($1) AS t (
			message_id text,
			user_id text,
			score integer,
			create_date timestamptz
		)
	ON CONFLICT (message_id) DO UPDATE SET score = EXCLUDED.score
	`, &entries)
	return err
}
