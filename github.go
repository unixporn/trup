package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v31/github"
)

func githubIssueLoop(s *discordgo.Session, client *github.Client) {
	since := time.Now().Truncate(time.Hour * 48)
	postedIssues := map[int64]struct{}{}

	for {
		time.Sleep(2 * time.Minute)

		issues, _, err := client.Issues.ListByRepo(context.Background(), "unixporn", "trup", nil)
		if err != nil {
			log.Println("Error on ListByRepo", err)
			continue
		}

		var issuesToPost []*github.Issue
		for _, issue := range issues {
			if _, posted := postedIssues[*issue.ID]; !posted && issue.CreatedAt.After(since) {
				issuesToPost = append(issuesToPost, issue)
			}
		}
		since = time.Now()

		for _, issue := range issuesToPost {
			embed := discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name: "New Github Issue",
				},
				URL:         *issue.URL,
				Title:       *issue.User.Login + ": " + *issue.Title,
				Description: *issue.Body,
			}

			_, err = s.ChannelMessageSendEmbed(env.ChannelBot, &embed)
			if err != nil {
				log.Printf("Failed to post embed with issue %d; Error: %s\n", *issue.ID, err)
				continue
			}

			postedIssues[*issue.ID] = struct{}{}
		}
	}
}

func githubLoop(s *discordgo.Session) {
	since := time.Now()
	thankedUsers := map[string]struct{}{}
	for {
		time.Sleep(2 * time.Minute)

		client := github.NewClient(nil)
		stargazers, _, err := client.Activity.ListStargazers(context.Background(), "unixporn", "trup", nil)
		if err != nil {
			log.Println("Error on ListStargazers", err.Error())
			continue
		}

		var usersToThank []string
		for _, star := range stargazers {
			if _, thanked := thankedUsers[*star.User.Login]; !thanked && star.StarredAt.Time.After(since) {
				usersToThank = append(usersToThank, *star.User.Login)
			}
		}
		since = time.Now()
		if len(usersToThank) == 0 {
			continue
		}

		_, err = s.ChannelMessageSend(env.ChannelBot, "Thank you "+strings.Join(usersToThank, ",")+" for staring the repository! ❤️")
		if err != nil {
			log.Println("Failed to send message", err)
			continue
		}

		for _, u := range usersToThank {
			thankedUsers[u] = struct{}{}
		}
	}
}
