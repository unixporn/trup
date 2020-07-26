package db

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
)

var (
	fileExpiration   = time.Minute
	imageLogFilePath = "./imageLog"
)

func StoreImage(message *discordgo.Message, attachment *discordgo.MessageAttachment) error {
	if !isImage(attachment) {
		return nil
	}

	resp, err := http.Get(attachment.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	messageSendTime, err := message.Timestamp.Parse()
	if err != nil {
		return err
	}

	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}

	lo := tx.LargeObjects()
	objectId, err := lo.Create(context.Background(), 0)
	if err != nil {
		return err
	}
	obj, err := lo.Open(context.Background(), objectId, pgx.LargeObjectModeWrite)
	if err != nil {
		return err
	}
	_, err = io.Copy(obj, resp.Body)
	if err != nil {
		return err
	}
	err = obj.Close()
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		context.Background(),
		"INSERT INTO image_log_files (channel_id, message_id, attachment_id, filename, create_date, object_id) VALUES ($1, $2, $3, $4, $5, $6)",
		message.ChannelID, message.ID, attachment.ID, attachment.Filename, messageSendTime, objectId,
	)
	if err != nil {
		return err
	}
	tx.Commit(context.Background())
	return nil
}

func GetStoredImages(channelId string, messageId string) ([]*discordgo.File, error) {
	rows, err := db.Query(
		context.Background(),
		"SELECT object_id, filename FROM image_log_files WHERE channel_id = $1 AND message_id = $2",
		channelId, messageId,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	files := []*discordgo.File{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		objectId := values[0].(uint32)
		filename := values[1].(string)

		fileBytes, err := loadBytesFromLargeObject(objectId)
		if err != nil {
			log.Printf("Error loading bytes from largeobject: %s\n", err)
			continue
		}

		files = append(files, &discordgo.File{
			Name:   filename,
			Reader: fileBytes,
		})
	}
	return files, nil
}

func PruneExpiredImageLogs() error {
	rows, err := db.Query(
		context.Background(),
		"SELECT object_id FROM image_log_files WHERE CURRENT_TIMESTAMP - create_date > $1",
		fileExpiration,
	)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return err
		}

		err = deleteLargeObject(values[0].(uint32))
		if err != nil {
			return nil
		}
	}

	_, err = db.Exec(
		context.Background(),
		"DELETE FROM image_log_files WHERE CURRENT_TIMESTAMP - create_date > $1",
		fileExpiration,
	)
	if err != nil {
		return err
	}
	return nil
}

func loadBytesFromLargeObject(objectId uint32) (*bytes.Buffer, error) {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	lo := tx.LargeObjects()
	object, err := lo.Open(context.Background(), objectId, pgx.LargeObjectModeRead)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(object)

	object.Close()
	tx.Commit(context.Background())
	return buf, nil
}

func deleteLargeObject(objectId uint32) error {
	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}

	lo := tx.LargeObjects()
	err = lo.Unlink(context.Background(), objectId)
	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func isImage(attachment *discordgo.MessageAttachment) bool {
	return attachment.Height != 0 && attachment.Width != 0
}
