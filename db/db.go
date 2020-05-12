package db

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
)

var db *pgxpool.Pool

func init() {
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	db = conn
}
