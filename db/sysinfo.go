package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx"
)

type SysinfoData struct {
	Distro          string
	Kernel          string
	Terminal        string
	Editor          string
	DeWm            string
	Bar             string
	Resolution      string
	DisplayProtocol string
	Gtk3Theme       string
	GtkIconTheme    string
	Cpu             string
	Gpu             string
	Memory          uint64
	Image           string
}

type Sysinfo struct {
	User       string
	Info       SysinfoData
	ModifyDate time.Time
	CreateDate time.Time
}

func NewSysinfo(user string, info SysinfoData) *Sysinfo {
	now := time.Now()
	return &Sysinfo{
		User:       user,
		Info:       info,
		ModifyDate: now,
		CreateDate: now,
	}
}

func (sysinfo *Sysinfo) Save() error {
	_, err := db.Exec(context.Background(), "call sysinfo_set($1, $2, $3, $4)", sysinfo.User, &sysinfo.Info, sysinfo.ModifyDate, sysinfo.CreateDate)
	return err
}

func GetSysinfo(user string) (*Sysinfo, error) {
	row := db.QueryRow(context.Background(), "SELECT usr, info, modify_date, create_date FROM sysinfo WHERE usr = $1", user)
	var sysinfo Sysinfo
	err := row.Scan(&sysinfo.User, &sysinfo.Info, &sysinfo.ModifyDate, &sysinfo.CreateDate)
	if err != nil {
		return nil, err
	}

	return &sysinfo, nil
}

func UpdateSysinfoImage(userId, image string) {
	info, err := GetSysinfo(userId)
	if err != nil && err.Error() != pgx.ErrNoRows.Error() {
		log.Printf("Failed to fetch system info for %s; Error: %s\n", userId, err)
		return
	}

	if info == nil {
		info = NewSysinfo(userId, SysinfoData{
			Image: image,
		})
	} else {
		info.Info.Image = image
	}

	err = info.Save()
	if err != nil {
		log.Printf("Failed to save info %#v; Error: %s\n", info, err)
		return
	}
}

type TopSysinfo struct {
	Field      string
	Name       string
	Percentage int
}

func TopSysinfoFields() ([]TopSysinfo, error) {
	rows, err := db.Query(context.Background(), `
	SELECT fields.field, fields.name, FLOOR((fields.count::float / fields.total_count::float) * 100) FROM (
		SELECT 1 as order, * FROM top_field('Distro')
		UNION SELECT 2 as order, * FROM top_field('DeWm')
		UNION SELECT 3 as order, * FROM top_field('DisplayProtocol')
		UNION SELECT 4 as order, * FROM top_field('Terminal')
		UNION SELECT 5 as order, * FROM top_field('Bar')
		UNION SELECT 6 as order, * FROM top_field('Gtk3Theme')
		UNION SELECT 7 as order, * FROM top_field('GtkIconTheme')
		UNION SELECT 8 as order, * FROM top_field('Editor')
		UNION SELECT 9 as order, * FROM top_field('Cpu')
	) fields WHERE total_count != 0 ORDER BY fields.order ASC;
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []TopSysinfo
	for rows.Next() {
		var info TopSysinfo
		err := rows.Scan(&info.Field, &info.Name, &info.Percentage)
		if err != nil {
			return nil, err
		}
		results = append(results, info)
	}

	return results, nil
}
