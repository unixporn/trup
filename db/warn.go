package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Warn struct {
	Id         pgtype.UUID
	Moderator  string
	User       string
	Reason     string
	CreateDate time.Time
}

func NewWarn(mod, user, reason string) *Warn {
	return &Warn{
		pgtype.UUID{},
		mod,
		user,
		reason,
		time.Now(),
	}
}

func (warn *Warn) Save() error {
	_, err := db.Exec(context.Background(), "insert into warn(id, moderator, usr, reason, create_date) values(uuid_generate_v4(), $1, $2, $3, $4)", warn.Moderator, warn.User, warn.Reason, warn.CreateDate)
	return err
}

func GetWarns(user string) ([]Warn, error) {
	rows, err := db.Query(context.Background(), "select id, moderator, usr, reason, create_date from warn where usr=$1", user)
	if err != nil {
		return nil, err
	}

	var res []Warn
	for rows.Next() {
		var w Warn
		err = rows.Scan(&w.Id, &w.Moderator, &w.User, &w.Reason, &w.CreateDate)
		if err != nil {
			return nil, err
		}
		res = append(res, w)
	}

	return res, nil
}

func CountWarns(user string) (int, error) {
	row := db.QueryRow(context.Background(), "select count(*) from warn where usr=$1", user)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}
