package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
)

var (
	db         *pgxpool.Pool
	databaseOk = false
)

func DatabaseUsable() bool {
	return databaseOk
}

func init() {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		fmt.Println("Database disabled")
		return
	}
	conn, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Printf("Unable to connect to the database, the database features will be disabled:\n%s\n", err)
		return
	}
	databaseOk = true
	db = conn
}
