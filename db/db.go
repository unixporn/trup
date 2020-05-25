package db

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
)

var db *pgxpool.Pool

func DatabaseUsable() bool {
	return db != nil
}

func init() {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Println("Database disabled")
		return
	}
	conn, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		log.Printf("Unable to connect to the database, the database features will be disabled:\n%s\n", err)
		return
	}
	db = conn
}
