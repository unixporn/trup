package db

import (
	"context"

	"github.com/jackc/pgx/v4"
)

func SetID(messageID string) error {
	_, err := db.Exec(context.Background(), "delete from psa")
	if err != nil {
		return err
	}
	_, err = db.Exec(context.Background(), "insert into psa(id) values($1)", messageID)
	return err
}

func GetID() (messageID string, err error) {
	row := db.QueryRow(context.Background(), "SELECT id FROM psa")
	err = row.Scan(&messageID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return messageID, err
	}
	return messageID, nil
}
