package db

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

var (
	fileExpirationMinutes = 5
	imageLogFilePath      = "./imageLog"
)

func StoreImage(message *discordgo.Message, attachment *discordgo.MessageAttachment) error {
	if !isImage(attachment) {
		return nil
	}

	err := DownloadFile(attachment.URL, imageLogFilePath+"/"+attachment.ID)
	if err != nil {
		return err
	}

	messageSendTime, err := message.Timestamp.Parse()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		context.Background(),
		"INSERT INTO image_log_files (channel_id, message_id, attachment_id, filename, create_date) VALUES ($1, $2, $3, $4, $5)",
		message.ChannelID, message.ID, attachment.ID, attachment.Filename, messageSendTime,
	)
	if err != nil {
		return err
	}
	return nil
}

type CachedFile struct {
	Filepath     string
	Filename     string
	AttachmentId int64
}

func GetStoredImages(channelId string, messageId string) ([]*discordgo.File, error) {
	rows, err := db.Query(
		context.Background(),
		"SELECT filename, attachment_id FROM image_log_files WHERE channel_id = $1 AND message_id = $2",
		channelId, messageId,
	)
	if err != nil {
		return nil, err
	}
	files := []*discordgo.File{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		attachmentId := values[1].(int64)
		cachedFile := CachedFile{Filepath: fmt.Sprintf("%s/%d", imageLogFilePath, attachmentId), Filename: values[0].(string), AttachmentId: attachmentId}
		reader, err := os.Open(cachedFile.Filepath)
		if err != nil {
			return nil, err
		}
		files = append(files, &discordgo.File{
			Name:   cachedFile.Filename,
			Reader: reader,
		})
	}
	return files, nil
}

func PopExpiredImageLogs() ([]CachedFile, error) {
	rows, err := db.Query(
		context.Background(),
		`SELECT filename, attachment_id FROM image_log_files 
     WHERE 
      ( DATE_PART('day',    CURRENT_TIMESTAMP - create_date) * 60 * 24 
      + DATE_PART('hour',   CURRENT_TIMESTAMP - create_date) * 60 
      + DATE_PART('minute', CURRENT_TIMESTAMP - create_date)
      ) > $1`,
		fileExpirationMinutes,
	)
	if err != nil {
		return nil, err
	}
	files := []CachedFile{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		attachmentId := values[1].(int64)
		cachedFile := CachedFile{Filepath: fmt.Sprintf("%s/%d", imageLogFilePath, attachmentId), Filename: values[0].(string), AttachmentId: attachmentId}
		files = append(files, cachedFile)
	}

	_, err = db.Exec(
		context.Background(),
		`DELETE FROM image_log_files 
     WHERE 
      ( DATE_PART('day',    CURRENT_TIMESTAMP - create_date) * 60 * 24 
      + DATE_PART('hour',   CURRENT_TIMESTAMP - create_date) * 60 
      + DATE_PART('minute', CURRENT_TIMESTAMP - create_date)
      ) > $1`,
		fileExpirationMinutes,
	)
	if err != nil {
		log.Printf("Failed to pop expired images from database: %s", err)
	}
	return files, nil
}

func isImage(attachment *discordgo.MessageAttachment) bool {
	return attachment.Height != 0 && attachment.Width != 0
}

func DownloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
