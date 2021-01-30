package job

import (
	"trup/command"

	"github.com/bwmarrin/discordgo"
)

func messageCreate(env *command.Env) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
	}
}
