package commands

import (
	"context"
	"fmt"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/corona"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func vaccines(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	vaccines, err := corona.GetVaccines()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Error querying vaccines from arcgis")
		return
	}
	p := message.NewPrinter(language.English)
	s.ChannelMessageSendEmbed(m.ChannelID, embed.NewEmbed().SetTitle("Vaccines Rollout in Ireland").SetDescription(p.Sprintf(`
		**First installment**: %d
	`, vaccines.First)).SetFooter(fmt.Sprintf("As of %s", vaccines.Date.Format(layoutIE))).MessageEmbed)
}
