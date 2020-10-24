package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	covidSummary = "https://api.covid19api.com/summary"
)

// CountrySummary contains covid data.
type CountrySummary struct {
	Country        string
	CountryCode    string
	Slug           string
	NewConfirmed   int
	TotalConfirmed int
	NewDeaths      int
	TotalDeaths    int
	NewRecovered   int
	TotalRecovered int
	Date           time.Time
}

// TotalSummary contains global data.
type TotalSummary struct {
	Global    map[string]int
	Countries []CountrySummary
}

// GetCountry returns a country.
func (t *TotalSummary) GetCountry(slug string) *CountrySummary {
	for _, country := range t.Countries {
		if country.Slug == slug {
			return &country
		}
	}
	return nil
}

// GetCorona give you corona.
func GetCorona() (total *TotalSummary, err error, raw bytes.Buffer) {
	var resp *http.Response
	resp, err = http.Get(covidSummary)
	if err != nil {
		return
	}
	total = &TotalSummary{}
	if err = json.NewDecoder(io.TeeReader(resp.Body, &raw)).Decode(total); err != nil {
		return
	}
	return
}

func coronaCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	total, err, _ := GetCorona()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error occured parsing covid stats")
		log.WithError(err).WithContext(ctx).Error("covid summary invalid output")
	}
	cm, slug := extractCommand(m.Content)
	title := "Covid-19 Stats for"
	p := message.NewPrinter(language.English)
	if slug == cm {
		slug = viper.GetString("corona.default")
		// Also send global stats
		body := "**New**\n"
		body += p.Sprintf("Cases: %d\n", total.Global["NewConfirmed"])
		body += p.Sprintf("Deaths: %d\n", total.Global["NewDeaths"])
		body += p.Sprintf("Recoveries: %d\n", total.Global["NewRecovered"])
		body += "\n**Total**\n"
		body += p.Sprintf("Cases: %d\n", total.Global["TotalConfirmed"])
		body += p.Sprintf("Deaths: %d\n", total.Global["TotalDeaths"])
		body += p.Sprintf("Recoveries: %d\n", total.Global["TotalRecovered"])

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
	country := total.GetCountry(slug)
	body := "**New**\n"
	body += p.Sprintf("Cases: %d\n", country.NewConfirmed)
	body += p.Sprintf("Deaths: %d\n", country.NewDeaths)
	body += p.Sprintf("Recoveries: %d\n", country.NewRecovered)
	body += "\n**Total**\n"
	body += p.Sprintf("Cases: %d\n", country.TotalConfirmed)
	body += p.Sprintf("Deaths: %d\n", country.TotalDeaths)
	body += p.Sprintf("Recoveries: %d\n", country.TotalRecovered)

	emb := embed.NewEmbed()
	emb.SetTitle(strings.Join([]string{title, strings.Title(strings.ReplaceAll(slug, "-", " "))}, " "))
	emb.SetDescription(body)
	emb.SetFooter(fmt.Sprintf("As of %s", country.Date.Format(layoutIE)))
	emb.SetColor(0x9b12f1)
	s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
}
