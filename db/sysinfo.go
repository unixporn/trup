package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx"
)

type SysinfoData struct {
	Distro          string
	Host            string
    Kernel          string
	Terminal        string
	Editor          string
	DeWm            string
	Bar             string
	Resolution      string
	DisplayProtocol string
	Shell           string
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
	    UNION SELECT 2 as order, * FROM top_field('Host')
        UNION SELECT 3 as order, * FROM top_field('DeWm')
		UNION SELECT 4 as order, * FROM top_field('DisplayProtocol')
		UNION SELECT 5 as order, * FROM top_field('Shell')
		UNION SELECT 6 as order, * FROM top_field('Terminal')
		UNION SELECT 7 as order, * FROM top_field('Bar')
		UNION SELECT 8 as order, * FROM top_field('Gtk3Theme')
		UNION SELECT 9 as order, * FROM top_field('GtkIconTheme')
		UNION SELECT 10 as order, * FROM top_field('Editor')
		UNION SELECT 11 as order, * FROM top_field('Cpu')
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

type TopFieldValue struct {
	Name       string
	Percentage int
}

func TopFieldValues(field string) ([]TopFieldValue, error) {
	rows, err := db.Query(context.Background(), `
WITH total_count AS
  (SELECT count(*)
   FROM sysinfo
   WHERE info->>$1 != ''),
top_names AS
  (SELECT info->>$1 AS name,
                 count(*)
   FROM sysinfo
   WHERE info->>$1 != ''
   GROUP BY info->>$1
   ORDER BY COUNT DESC
   LIMIT 5)
SELECT name,
       FLOOR((count::float /
                (SELECT *
                 FROM total_count)::float) * 100) AS percentage
FROM top_names;
	`, &field)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []TopFieldValue
	for rows.Next() {
		var info TopFieldValue
		err := rows.Scan(&info.Name, &info.Percentage)
		if err != nil {
			return nil, err
		}
		results = append(results, info)
	}
	return results, nil
}

type FieldValueStat struct {
	Count      int
	Percentage int
}

func FetchFieldValueStat(field string, value string) (FieldValueStat, error) {
	row := db.QueryRow(context.Background(), `
SELECT
  count,
  ROUND((count::float/total::float)*100) AS percentage
FROM (
  SELECT
    count(*),
    (SELECT count(*) FROM sysinfo WHERE info->>$1 != '') AS total
  FROM sysinfo WHERE info->>$1 ~* $2
) t;
`, &field, &value)

	var stat FieldValueStat
	err := row.Scan(&stat.Count, &stat.Percentage)
	if err != nil {
		return FieldValueStat{}, err
	}

	return stat, nil
}
