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

func coronaCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	var embeds []*discordgo.MessageEmbed
	var countryInput string
	args := i.ApplicationCommandData().Options
	total, err, _ := corona.GetCorona()
	if err != nil {
		log.WithError(err).WithContext(ctx).Error("covid summary invalid output")
		InteractionResponseError(s, i, "Unable to parse covid stats", true)
		return
	}
	p := message.NewPrinter(language.English)
	if len(args) < 1 {
		countryInput = viper.GetString("corona.default")
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
		embeds = append(embeds, emb.MessageEmbed)
	} else {
		countryInput = strings.ToLower(
			strings.ReplaceAll(
				strings.TrimSpace(args[0].StringValue()),
				" ", "-",
			),
		)
	}
	var country *corona.CountrySummary
	if countryInput == viper.GetString("corona.default") {
		_, country, err = corona.GetArcgis()
		if err != nil {
			log.WithError(err).WithContext(ctx).Error("covid arcgis invalid output")
			InteractionResponseError(s, i, "Unable to parse covid stats", true)
			return
		}
	} else {
		country = total.GetCountry(countryInput)
	}
	if country != nil {
		var coronaEmbeds []*discordgo.MessageEmbed
		coronaEmbeds, err = corona.CreateEmbed(country, s, ctx)
		if err != nil {
			InteractionResponseError(s, i, err.Error(), true)
			return
		}
		embeds = append(embeds, coronaEmbeds...)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: embeds,
			},
		})
		if err != nil {
			log.WithContext(ctx).WithError(err)
		}
	} else {
		InteractionResponseError(s, i, fmt.Sprintf("Couldn't find a country called %s", countryInput), false)
	}
}
