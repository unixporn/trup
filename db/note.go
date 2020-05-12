package db

import (
	"context"
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Note struct {
	Id         pgtype.UUID
	Taker      string
	About      string
	Content    string
	CreateDate time.Time
}

func NewNote(taker, about, content string) *Note {
	return &Note{
		pgtype.UUID{},
		taker,
		about,
		content,
		time.Now(),
	}
}

func (note *Note) Save() error {
	_, err := db.Exec(context.Background(), "insert into note(id, taker, about, content, create_date) values(uuid_generate_v4(), $1, $2, $3, $4)", note.Taker, note.About, note.Content, note.CreateDate)
	return err
}

func GetNotes(about string) ([]Note, error) {
	var res []Note
	rows, err := db.Query(context.Background(), "select id, taker, about, content, create_date from note where about=$1 order by create_date desc limit 25", &about)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var note Note
		err = rows.Scan(&note.Id, &note.Taker, &note.About, &note.Content, &note.CreateDate)
		if err != nil {
			return nil, err
		}
		res = append(res, note)
	}

	return res, nil
}
