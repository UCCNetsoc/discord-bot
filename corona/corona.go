package corona

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
	"github.com/wcharczuk/go-chart/drawing"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	covidSummary = "https://api.covid19api.com/summary"
	covidDayOne  = "https://api.covid19api.com/dayone/country/%s"
	imgHost      = "https://freeimage.host/api/1/upload"
	layoutIE     = "02/01/06"
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
	aggregate := 0
	for _, cases := range history {
		totalCases = append([]float64{float64(cases.Confirmed - aggregate)}, totalCases...)
		aggregate = cases.Confirmed
		dates = append([]time.Time{cases.Date}, dates...)
	}
	graph := chart.Chart{
		Title: fmt.Sprintf("Cases per day for %s", c.Country),
		XAxis: chart.XAxis{
			Name:           "Date",
			Style:          chart.StyleShow(),
			ValueFormatter: chart.TimeDateValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:           "New Cases",
			ValueFormatter: func(v interface{}) string { return chart.FloatValueFormatterWithFormat(v, "%.0f") },
			Style:          chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Style: chart.Style{
					StrokeWidth: 5,
					StrokeColor: drawing.ColorFromHex("9b12f1"),
					Show:        true,
				},
				XValues: dates,
				YValues: totalCases,
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

// CreateEmbed sends a corona embed to the specified channel.
func CreateEmbed(country *CountrySummary, s *discordgo.Session, channelID string, ctx context.Context) {
	title := "Covid-19 Stats for"
	p := message.NewPrinter(language.English)
	body := "**New**\n"
	body += p.Sprintf("Cases: %d\n", country.NewConfirmed)
	body += p.Sprintf("Deaths: %d\n", country.NewDeaths)
	body += p.Sprintf("Recoveries: %d\n", country.NewRecovered)
	body += "\n**Total**\n"
	body += p.Sprintf("Cases: %d\n", country.TotalConfirmed)
	body += p.Sprintf("Deaths: %d\n", country.TotalDeaths)
	body += p.Sprintf("Recoveries: %d\n", country.TotalRecovered)

	emb := embed.NewEmbed()
	emb.SetTitle(strings.Join([]string{title, strings.Title(strings.ReplaceAll(country.Slug, "-", " "))}, " "))
	emb.SetDescription(body)
	emb.SetFooter(fmt.Sprintf("As of %s", country.Date.Format(layoutIE)))
	emb.SetColor(0x9b12f1)
	// Upload graph to freeimage.
	if apiKey := viper.GetString("freeimage.key"); apiKey != "" {
		graph, err := country.Graph()
		if err != nil {
			s.ChannelMessageSend(channelID, "Error occured generating graph")
			log.WithError(err).WithContext(ctx).Error("Error occured generating graph")
			return
		}
		img := dataurl.EncodeBytes(graph.Bytes())
		img = img[22:]
		resp, err := http.PostForm(imgHost, map[string][]string{
			"key":    {apiKey},
			"source": {img},
		})
		if err != nil {
			s.ChannelMessageSend(channelID, "Error occured uploading graph")
			log.WithError(err).WithContext(ctx).Error("Error occured uploading graph")
			return
		}
		imgResp := struct {
			Image struct {
				URL string `json:"url"`
			} `json:"image"`
		}{}
		if err = json.NewDecoder(resp.Body).Decode(&imgResp); err != nil {
			s.ChannelMessageSend(channelID, "Error occured parsing graph url")
			log.WithError(err).WithContext(ctx).Error("Error occured parsing graph url")
			return
		}
		emb.SetImage(imgResp.Image.URL)
	}
	s.ChannelMessageSendEmbed(channelID, emb.MessageEmbed)
}
