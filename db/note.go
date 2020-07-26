package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type NoteType int

const (
	ManualNote         NoteType = 0
	BlocklistViolation NoteType = 1
)

type Note struct {
	Id         pgtype.UUID
	Taker      string
	About      string
	Content    string
	NoteType   NoteType
	CreateDate time.Time
}

func NewNote(taker, about, content string, noteType NoteType) *Note {
	return &Note{
		pgtype.UUID{},
		taker,
		about,
		content,
		noteType,
		time.Now(),
	}
}

func (note *Note) Save() error {
	_, err := db.Exec(context.Background(), "INSERT INTO note(id, taker, about, content, note_type, create_date) VALUES(uuid_generate_v4(), $1, $2, $3, $4, $5)", note.Taker, note.About, note.Content, note.NoteType, note.CreateDate)
	return err
}

func GetNotes(about string) ([]Note, error) {
	var res []Note
	rows, err := db.Query(context.Background(), "SELECT id, taker, about, content, note_type, create_date FROM note WHERE about=$1 ORDER BY create_date DESC LIMIT 25", &about)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var note Note
		err = rows.Scan(&note.Id, &note.Taker, &note.About, &note.Content, &note.NoteType, &note.CreateDate)
		if err != nil {
			return nil, err
		}
		res = append(res, note)
	}

	return res, nil
}
