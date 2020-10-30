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
	"github.com/vincent-petithory/dataurl"
	"github.com/wcharczuk/go-chart"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	covidSummary = "https://api.covid19api.com/summary"
	covidDayOne  = "https://api.covid19api.com/dayone/country/%s"
)

// CountryBase is the basic country stats.
type CountryBase struct {
	Country     string
	CountryCode string
	Date        time.Time
}

// CountryDaily contains daily confirmed cases.
type CountryDaily struct {
	CountryBase
	Confirmed int
}

// CountrySummary contains covid data.
type CountrySummary struct {
	CountryBase
	Slug           string
	NewConfirmed   int
	TotalConfirmed int
	NewDeaths      int
	TotalDeaths    int
	NewRecovered   int
	TotalRecovered int
}

func (c *CountrySummary) getHistory() ([]CountryDaily, error) {
	resp, err := http.Get(fmt.Sprintf(covidDayOne, c.Slug))
	if err != nil {
		return nil, err
	}
	country := []CountryDaily{}
	if err = json.NewDecoder(resp.Body).Decode(&country); err != nil {
		return nil, err
	}
	return country, nil
}

// Graph generated a graph of historic cases.
func (c *CountrySummary) Graph() (*bytes.Buffer, error) {
	history, err := c.getHistory()
	if err != nil {
		return nil, err
	}
	dates := []time.Time{}
	totalCases := []float64{}
	for _, cases := range history {
		totalCases = append([]float64{float64(cases.Confirmed)}, totalCases...)
		dates = append([]time.Time{cases.Date}, dates...)
	}
	graph := chart.Chart{
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dates,
				YValues: totalCases,
				Name:    fmt.Sprintf("Cases per day for %s", c.Country),
			},
		},
	}
	buffer := bytes.NewBuffer([]byte{})
	if err = graph.Render(chart.PNG, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
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
		return
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

	graph, err := country.Graph()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error occured generating graph")
		log.WithError(err).WithContext(ctx).Error("Error occured generating graph")
		return
	}
	img := dataurl.EncodeBytes(graph.Bytes())
	s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
}
