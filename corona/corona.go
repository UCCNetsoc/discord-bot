package corona

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	covidSummary      = "https://api.covid19api.com/summary"
	covidDayOne       = "https://api.covid19api.com/total/dayone/country/%s/status/confirmed"
	imgHost           = "https://freeimage.host/api/1/upload"
	arcgis            = "https://services1.arcgis.com/eNO7HHeQ3rUcBllm/arcgis/rest/services/CovidStatisticsProfileHPSCIrelandOpenData/FeatureServer/0/query?f=json&where=1%3D1&returnGeometry=false&spatialRel=esriSpatialRelIntersects&outFields=*&orderByFields=Date%20asc&resultOffset=0&resultRecordCount=32000&resultType=standard&cacheHint=true"
	covidVaccines     = "https://services-eu1.arcgis.com/z6bHNio59iTqqSUY/arcgis/rest/services/Covid19_Vaccine_Administration_Hosted_View/FeatureServer/0/query?f=json&where=1=1&outFields=*&returnGeometry=false"
	covidVaccinesType = "https://services-eu1.arcgis.com/z6bHNio59iTqqSUY/arcgis/rest/services/Covid19_Vaccine_Administration_VaccineTypeHostedView_V2/FeatureServer/0/query?f=json&where=1%3D1&outFields=*&returnGeometry=false"
	layoutIE          = "02/01/06"
	sleepTime         = time.Duration(3 * time.Minute)
)

var (
	currentDate     *time.Time
	currentVaccines *Vaccines
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
	Cases  int
	Deaths int
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

// Get Arcgis data.
func GetArcgis() (daily []CountryDaily, summary *CountrySummary, err error) {
	var resp *http.Response
	if resp, err = http.Get(arcgis); err != nil {
		return
	}
	data := struct {
		Features []struct {
			Attributes struct {
				Date                     int64
				ConfirmedCovidCases      int
				TotalConfirmedCovidCases int
				ConfirmedCovidDeaths     int
				TotalCovidDeaths         int
			} `json:"attributes"`
		} `json:"features"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}
	for _, attrs := range data.Features {
		country := attrs.Attributes
		daily = append(daily, CountryDaily{
			CountryBase: CountryBase{
				Date:        time.Unix(country.Date/1000, 0),
				Country:     "Ireland",
				CountryCode: "IE",
			},
			Cases:  country.TotalConfirmedCovidCases,
			Deaths: country.TotalCovidDeaths,
		})
	}
	if len(data.Features) > 0 {
		last := data.Features[len(data.Features)-1].Attributes
		summary = &CountrySummary{
			CountryBase: CountryBase{
				Country:     "Ireland",
				CountryCode: "IE",
				Date:        time.Unix(last.Date/1000, 0),
			},
			Slug:           "ireland",
			NewConfirmed:   last.ConfirmedCovidCases,
			NewDeaths:      last.ConfirmedCovidDeaths,
			TotalConfirmed: last.TotalConfirmedCovidCases,
			TotalDeaths:    last.TotalCovidDeaths,
		}
	} else {
		err = errors.New("No data received from arcgis")
	}
	return
}

func (c *CountrySummary) getHistory() ([]CountryDaily, error) {
	if c.Slug == viper.GetString("corona.default") {
		daily, _, err := GetArcgis()
		if err != nil {
			return nil, err
		}
		return daily, nil
	}
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
func (c *CountrySummary) Graph(month bool) (*bytes.Buffer, error) {
	history, err := c.getHistory()
	if err != nil {
		return nil, err
	}
	var first CountryDaily
	if month {
		first = history[len(history)-33]
		history = history[len(history)-32:]
	}
	dates := []time.Time{}
	totalCases := []float64{}
	totalDeaths := []float64{}
	aggregate := 0
	aggregateDeaths := 0
	if month {
		aggregate = first.Cases
		aggregateDeaths = first.Deaths
	}
	for _, cases := range history {
		newCases := float64(cases.Cases - aggregate)
		if newCases < 0 {
			continue
		}
		if !month && aggregate > 0 && newCases > 1000 && newCases > float64(aggregate)*5 {
			continue
		}
		newDeaths := float64(cases.Deaths - aggregateDeaths)
		if newDeaths < 0 {
			continue
		}
		totalDeaths = append([]float64{newDeaths}, totalDeaths...)
		aggregateDeaths = cases.Deaths
		totalCases = append([]float64{newCases}, totalCases...)
		aggregate = cases.Cases
		dates = append([]time.Time{cases.Date}, dates...)
	}
	graph := chart.Chart{
		Canvas: chart.Style{
			StrokeWidth: 0,
			FillColor:   drawing.ColorFromHex("2f3136"),
			Show:        true,
		},

		Background: chart.Style{
			StrokeWidth: 0,
			FillColor:   drawing.ColorFromHex("2f3136"),
			Show:        true,
		},
		Title: fmt.Sprintf("Cases per day for %s", c.Country),
		XAxis: chart.XAxis{
			Name: "Date",
			TickStyle: chart.Style{
				StrokeWidth: 1,
				StrokeColor: drawing.ColorFromHex("ffffff"),
				FillColor:   drawing.ColorFromHex("2f3136"),
				Show:        true,
			},
			Style: chart.Style{
				FontColor: drawing.ColorFromHex("ffffff"),
				FillColor: drawing.ColorFromHex("2f3136"),
				Show:      true,
			},
			ValueFormatter: chart.TimeDateValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:           "New Cases",
			ValueFormatter: func(v interface{}) string { return chart.FloatValueFormatterWithFormat(v, "%.0f") },
			TickStyle: chart.Style{
				StrokeWidth: 1,
				StrokeColor: drawing.ColorFromHex("ffffff"),
				FillColor:   drawing.ColorFromHex("2f3136"),
				Show:        true,
			},
			Style: chart.Style{
				FontColor: drawing.ColorFromHex("ffffff"),
				FillColor: drawing.ColorFromHex("2f3136"),
				Show:      true,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Style: chart.Style{
					StrokeWidth: 5,
					StrokeColor: drawing.ColorFromHex("ffffff"),
					FillColor:   drawing.ColorFromHex("2f3136"),
					Show:        true,
				},
				XValues: dates,
				YValues: totalCases,
			},
			chart.TimeSeries{
				Style: chart.Style{
					StrokeWidth: 2,
					StrokeColor: drawing.ColorFromHex("ffffff"),
					FillColor:   drawing.ColorFromHex("2f3136"),
					Show:        true,
				},
				XValues: dates,
				YValues: totalDeaths,
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
	Global    map[string]interface{}
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
	temp := make(map[string]interface{})
	for col, val := range total.Global {
		if f, ok := val.(float64); ok {
			temp[col] = int(f)
		}
	}
	total.Global = temp
	return
}

// CreateEmbed sends a corona embed to the specified channel.
func CreateEmbed(country *CountrySummary, s *discordgo.Session, ctx context.Context) ([]*discordgo.MessageEmbed, error) {
	var embeds []*discordgo.MessageEmbed
	title := "Covid-19 Stats for"
	p := message.NewPrinter(language.English)
	body := "**New**\n"
	body += p.Sprintf("Cases: %d\n", country.NewConfirmed)
	body += p.Sprintf("Deaths: %d\n", country.NewDeaths)
	body += "\n**Total**\n"
	body += p.Sprintf("Cases: %d\n", country.TotalConfirmed)
	body += p.Sprintf("Deaths: %d\n", country.TotalDeaths)

	emb := embed.NewEmbed()
	emb.SetTitle(strings.Join([]string{title, strings.Title(strings.ReplaceAll(country.Slug, "-", " "))}, " "))
	emb.SetDescription(body)
	emb.SetFooter(fmt.Sprintf("As of %s", country.Date.Format(layoutIE)))
	emb.SetColor(0x9b12f1)
	// Upload graph to freeimage.
	graph, err := country.Graph(false)
	if err != nil {
		log.WithError(err).WithContext(ctx).Error("Error occured generating graph")
		return nil, errors.New("error occured generating graph")
	}
	img, err := upload(graph)
	if err != nil {
		log.WithError(err).WithContext(ctx).Error("Error occured uploading graph")
		return nil, errors.New("error occured uploading graph")
	}
	emb.SetImage(img)
	embeds = append(embeds, emb.MessageEmbed)

	// monthly graph embed
	monthlyGraph, err := country.Graph(true)
	if err != nil {
		log.WithError(err).WithContext(ctx).Error("Error occured generating graph")
		return nil, errors.New("error occured generating graph")
	}
	monthlyImg, err := upload(monthlyGraph)
	if err != nil {
		log.WithError(err).WithContext(ctx).Error("Error occured uploading graph")
		return nil, errors.New("error occured uploading graph")
	}
	monthlyEmb := embed.NewEmbed()
	monthlyEmb.SetImage(monthlyImg)
	monthlyEmb.SetTitle(fmt.Sprintf("Last 31 days cases for %s", strings.Title(strings.ReplaceAll(country.Slug, "-", " "))))
	monthlyEmb.SetColor(0x9b12f1)
	embeds = append(embeds, monthlyEmb.MessageEmbed)
	return embeds, nil
}

func upload(b *bytes.Buffer) (string, error) {
	img := dataurl.EncodeBytes(b.Bytes())
	img = img[22:]
	var apiKey string
	if apiKey = viper.GetString("freeimage.key"); apiKey == "" {
		return "", errors.New("No freeimage key")
	}
	resp, err := http.PostForm(imgHost, map[string][]string{
		"key":    {apiKey},
		"source": {img},
	})
	if err != nil {
		return "", err
	}
	imgResp := struct {
		Image struct {
			URL string `json:"url"`
		} `json:"image"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&imgResp); err != nil {
		return "", err
	}
	return imgResp.Image.URL, nil
}

// Listen for new covid cases.
func Listen(s *discordgo.Session) error {
	channelID := viper.GetString("discord.public.corona")
	for {
		_, data, err := GetArcgis()
		if err != nil {
			log.WithError(err).Error("error occured listening for HSE corona updates")
			time.Sleep(30 * time.Second)
			continue
		}
		if currentDate == nil {
			currentDate = &data.Date
		} else if data.Date.Unix() > currentDate.Unix() {
			// New data found.
			currentDate = &data.Date
			log.WithFields(log.Fields{"arcGis": *data}).Info("Found new COVID cases from the HSE")
			s.ChannelMessageSend(channelID, "The HSE has released new case numbers for Ireland:")
			coronaEmbeds, err := CreateEmbed(data, s, context.Background())
			if err != nil {
				s.ChannelMessageSend(channelID, err.Error())
			} else {
				for _, coronaEmbed := range coronaEmbeds {
					s.ChannelMessageSendEmbed(channelID, coronaEmbed)
				}
			}
		}
		vaccines, err := GetVaccines()
		if err != nil {
			log.WithError(err).Error("error occured listening for HSE vaccine updates")
			time.Sleep(30 * time.Second)
			continue
		}
		if currentVaccines == nil {
			currentVaccines = vaccines
		} else if vaccines.Date.Unix() > currentVaccines.Date.Unix() {
			// New vaccines found.
			s.ChannelMessageSend(channelID, "The HSE has released new vaccine numbers for Ireland:")
			s.ChannelMessageSendEmbed(channelID, vaccines.Embed(currentVaccines))
			currentVaccines = vaccines
		}
		<-time.After(sleepTime)
	}
}
