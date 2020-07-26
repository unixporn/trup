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

const (
	fileExpiration    = time.Hour * 3
	maximumFilesizeMB = 10
)

func StoreAttachment(message *discordgo.Message, attachment *discordgo.MessageAttachment) error {
	if !isAttachment(attachment) || attachment.Size > maximumFilesizeMB*1000000 {
		return nil
	}

	resp, err := http.Get(attachment.ProxyURL)
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
		"INSERT INTO attachment_log_cache (channel_id, message_id, attachment_id, filename, create_date, object_id) VALUES ($1, $2, $3, $4, $5, $6)",
		message.ChannelID, message.ID, attachment.ID, attachment.Filename, messageSendTime, objectId,
	)
	if err != nil {
		return err
	}
	tx.Commit(context.Background())
	return nil
}

type StoredAttachment struct {
	Filename string
	Reader   io.Reader
}

func (storedAttachment *StoredAttachment) GetContentType() string {
	contentTypeBuf := make([]byte, 512)
	storedAttachment.Reader.Read(contentTypeBuf)
	storedAttachment.Reader = io.MultiReader(bytes.NewReader(contentTypeBuf), storedAttachment.Reader)
	return http.DetectContentType(contentTypeBuf)
}

func GetStoredAttachments(channelId string, messageId string) ([]*StoredAttachment, func() error, error) {
	rows, err := db.Query(
		context.Background(),
		"SELECT object_id, filename FROM attachment_log_cache WHERE channel_id = $1 AND message_id = $2",
		channelId, messageId,
	)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	tx, err := db.Begin(context.Background())
	if err != nil {
		return nil, nil, err
	}

	files := []*StoredAttachment{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, nil, err
		}
		objectId := values[0].(uint32)
		filename := values[1].(string)

		reader, err := loadBytesFromLargeObject(tx, objectId)
		if err != nil {
			log.Printf("Error loading reader from largeobject: %s\n", err)
			continue
		}

		files = append(files, &StoredAttachment{filename, reader})
	}
	onFinish := func() error {
		return tx.Commit(context.Background())
	}
	return files, onFinish, nil
}

func PruneExpiredAttachments() error {
	rows, err := db.Query(
		context.Background(),
		"SELECT object_id FROM attachment_log_cache WHERE CURRENT_TIMESTAMP - create_date > $1",
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
		"DELETE FROM attachment_log_cache WHERE CURRENT_TIMESTAMP - create_date > $1",
		fileExpiration,
	)
	if err != nil {
		return err
	}
	return nil
}

func loadBytesFromLargeObject(tx pgx.Tx, objectId uint32) (io.ReadCloser, error) {
	lo := tx.LargeObjects()
	object, err := lo.Open(context.Background(), objectId, pgx.LargeObjectModeRead)
	if err != nil {
		return nil, err
	}
	return object, nil
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

func isAttachment(attachment *discordgo.MessageAttachment) bool {
	return attachment.Height != 0 && attachment.Width != 0
}
