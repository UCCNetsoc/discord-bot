package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/corona"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func coronaCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	total, err, _ := corona.GetCorona()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error occured parsing covid stats")
		log.WithError(err).WithContext(ctx).Error("covid summary invalid output")
		return
	}
	cm, slug := extractCommand(m.Content)
	p := message.NewPrinter(language.English)
	if slug == cm {
		slug = viper.GetString("corona.default")
		// Also send global stats
		body := "**New**\n"
		body += p.Sprintf("Cases: %d\n", total.Global["NewConfirmed"].(int))
		body += p.Sprintf("Deaths: %d\n", total.Global["NewDeaths"].(int))
		body += p.Sprintf("Recoveries: %d\n", total.Global["NewRecovered"].(int))
		body += "\n**Total**\n"
		body += p.Sprintf("Cases: %d\n", total.Global["TotalConfirmed"].(int))
		body += p.Sprintf("Deaths: %d\n", total.Global["TotalDeaths"].(int))
		body += p.Sprintf("Recoveries: %d\n", total.Global["TotalRecovered"].(int))

		emb := embed.NewEmbed()
		emb.SetTitle("Covid-19 Global Stats")
		emb.SetColor(0x128af1)
		emb.SetDescription(body)
		s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
	} else {
		slug = strings.ToLower(
			strings.ReplaceAll(
				strings.TrimSpace(strings.TrimPrefix(slug, cm)),
				" ", "-",
			),
		)
	}
	var country *corona.CountrySummary
	if slug == viper.GetString("corona.default") {
		_, country, err = corona.GetArcgis()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error occured parsing covid stats")
			log.WithError(err).WithContext(ctx).Error("covid arcgis invalid output")
			return
		}
	} else {
		country = total.GetCountry(slug)
	}
	if country != nil {
		corona.CreateEmbed(country, s, m.ChannelID, ctx)
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Couldn't find a country called %s", slug))
	}
}
