package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Mute struct {
	Id        pgtype.UUID
	GuildId   string
	Moderator string
	User      string
	Reason    string
	StartTime time.Time
	EndTime   time.Time
}

func NewMute(guildId, mod, user, reason string, start, end time.Time) *Mute {
	return &Mute{
		pgtype.UUID{},
		guildId,
		mod,
		user,
		reason,
		start,
		end,
	}
}

func (mute *Mute) Save() error {
	_, err := db.Exec(context.Background(), "INSERT INTO mute(id, guildid, moderator, usr, end_time, start_time, active, reason) VALUES(uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7)", mute.GuildId, mute.Moderator, mute.User, mute.EndTime, mute.StartTime, true, mute.Reason)
	return err
}

func SetMuteInactive(id pgtype.UUID) error {
	_, err := db.Exec(context.Background(), "UPDATE mute SET active=false WHERE id=$1", &id)
	return err
}

// Careful: Returned mutes do not include all fields
func GetExpiredMutes() ([]Mute, error) {
	rows, err := db.Query(context.Background(), "SELECT id, guildid, usr FROM mute WHERE active=true AND end_time < CURRENT_TIMESTAMP")
	if err != nil {
		return nil, err
	}

	var res []Mute
	for rows.Next() {
		var m Mute
		err = rows.Scan(&m.Id, &m.GuildId, &m.User)
		if err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, nil
}
