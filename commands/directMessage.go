package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// dmCommands is invoked when receiving a direct message from a user. If they are in the
// process of registering, they cannot execute any other commands that they may normally
// be able to invoke, and must complete the registering flow first
func dmCommands(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	if commandStr, _ := extractCommand(m.Content); commandStr == "register" {
		delete(registering, m.Author.ID)
	} else if state, ok := registering[m.Author.ID]; ok {
		state = state(ctx, s, m)
		if state != nil {
			registering[m.Author.ID] = state
		}
		return
	}
	callCommand(s, m)
}
