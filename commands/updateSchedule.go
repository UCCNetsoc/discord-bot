package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/UCCNetsoc/discord-bot/api"
	"github.com/UCCNetsoc/discord-bot/config"
	"github.com/UCCNetsoc/discord-bot/prometheus"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

// Allows Captains to create a scheduled item that can be requested
// Command formate !addSchedule "Game" "Opponent" "time-date"

type entry struct {
	Game string
	Opponent string
	Date string
}

func updateSchedule(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	channels := viper.Get("discord.channels").(*config.Channels)
}