package corona

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Vaccines from the HSE arcGIS.
type Vaccines struct {
	First       int
	Second      int
	Total       int
	TotalAdmin  int
	Pfizer      int
	Moderna     int
	AstraZeneca int
	Date        time.Time
}

// Population of Ireland.
const Population = 4900000

func (v *Vaccines) Embed(prev *Vaccines) *discordgo.MessageEmbed {
	p := message.NewPrinter(language.English)
	firstPercentage := (float64(v.First) / float64(Population)) * 100
	secondPercentage := (float64(v.Second) / float64(Population)) * 100
	totalPercentage := (float64(v.Total) / float64(Population)) * 100
	adminPercentage := (float64(v.TotalAdmin) / float64(Population)) * 100

	pfPercentage := (float64(v.Pfizer) / float64(Population)) * 100
	azPercentage := (float64(v.AstraZeneca) / float64(Population)) * 100
	mdPercentage := (float64(v.Moderna) / float64(Population)) * 100

	var description string
	if prev == nil {
		description = p.Sprintf(`
				**First installment**: %d (%.2f%% of population)
				**Second installment**: %d (%.2f%% of population)
				**Fully vaccinated**: %d (%.2f%% of population)
				**Total Administered**: %d (%.2f%% of population)

				***Pfizer***: %d (%.2f%% of population)
				***AstraZeneca***: %d (%.2f%% of population)
				***Moderna***: %d (%.2f%% of population)
			`,
			v.First,
			firstPercentage,
			v.Second,
			secondPercentage,
			v.Total,
			totalPercentage,
			v.TotalAdmin,
			adminPercentage,
			v.Pfizer,
			pfPercentage,
			v.AstraZeneca,
			azPercentage,
			v.Moderna,
			mdPercentage,
		)
	} else {
		description = p.Sprintf(`
				__**New**__
				**First installment**: %d (+%d) (%.2f%% of population)
				**Second installment**: %d (+%d) (%.2f%% of population)
				**Fully vaccinated**: %d (+%d) (%.2f%% of population)
				**Total Administered**: %d (+%d) (%.2f%% of population)

				***Pfizer***: %d (+%d) (%.2f%% of population)
				***AstraZeneca***: %d (+%d) (%.2f%% of population)
				***Moderna***: %d (+%d) (%.2f%% of population)

				__**Previously**__
				**First installment**: %d
				**Second installment**: %d
				**Fully vaccinated**: %d
				**Total Administered**: %d 

				***Pfizer***: %d
				***AstraZeneca***: %d
				***Moderna***: %d 
			`,
			v.First, int64(math.Abs(float64(v.First-prev.First))), firstPercentage,
			v.Second, int64(math.Abs(float64(v.Second-prev.Second))), secondPercentage,
			v.Total, int64(math.Abs(float64(v.Total-prev.Total))), totalPercentage,
			v.TotalAdmin, int64(math.Abs(float64(v.TotalAdmin-prev.TotalAdmin))), adminPercentage,
			v.Pfizer, int64(math.Abs(float64(v.Pfizer-prev.Pfizer))), pfPercentage,
			v.AstraZeneca, int64(math.Abs(float64(v.AstraZeneca-prev.AstraZeneca))), azPercentage,
			v.Moderna, int64(math.Abs(float64(v.Moderna-prev.Moderna))), mdPercentage,

			prev.First,
			prev.Second,
			prev.Total,
			prev.TotalAdmin,
			prev.Pfizer,
			prev.AstraZeneca,
			prev.Moderna,
		)
	}
	return embed.NewEmbed().SetTitle("Vaccines Rollout in Ireland").SetDescription(description).SetFooter(fmt.Sprintf("As of %s", v.Date.Format(layoutIE))).MessageEmbed
}

// GetVaccines will query HSE arcGIS and return vaccine stats.
func GetVaccines() (*Vaccines, error) {
	resp, err := http.Get(covidVaccines)
	if err != nil {
		return nil, err
	}
	vaccines := &struct {
		Features []struct {
			Attributes struct {
				First   int `json:"firstDose"`
				Second  int `json:"secondDose"`
				Total   int `json:"totalAdministered"`
				Date    int `json:"relDate"`
				Pfizer  int `json:"pf"`
				Moderna int `json:"modern"`
				Az      int `json:"az"`
			} `json:"attributes"`
		} `json:"features"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(vaccines); err != nil {
		return nil, err
	}
	if len(vaccines.Features) < 1 {
		return nil, errors.New("no features")
	}
	return &Vaccines{
		First:       vaccines.Features[0].Attributes.First,
		Second:      vaccines.Features[0].Attributes.Second,
		Total:       vaccines.Features[0].Attributes.Second,
		TotalAdmin:  vaccines.Features[0].Attributes.Total,
		Pfizer:      vaccines.Features[0].Attributes.Pfizer,
		Moderna:     vaccines.Features[0].Attributes.Moderna,
		AstraZeneca: vaccines.Features[0].Attributes.Az,
		Date:        time.Unix(int64(vaccines.Features[0].Attributes.Date)/1000, 0),
	}, nil
}
