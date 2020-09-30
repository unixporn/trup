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

func UpdateSysinfoImage(userId string, image string) {
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
