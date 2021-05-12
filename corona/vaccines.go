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
	Janssen     int
	Date        time.Time
}

// Population of Ireland.
const Population = 4900000

func (v *Vaccines) Embed(prev *Vaccines) *discordgo.MessageEmbed {
	p := message.NewPrinter(language.English)
	firstPercentage := (float64(v.First) / float64(Population)) * 100
	secondPercentage := (float64(v.Second) / float64(Population)) * 100
	totalPercentage := (float64(v.Total) / float64(Population)) * 100

	var description string
	if prev == nil {
		description = p.Sprintf(`
**First installment**: %d (%.2f%% of population)
**Second installment**: %d (%.2f%% of population)
**Fully vaccinated**: %d (%.2f%% of population)
**Total Administered**: %d 

***Pfizer***: %d 
***AstraZeneca***: %d 
***Moderna***: %d 
***J&J/Janssen***: %d 
			`,
			v.First, firstPercentage,
			v.Second, secondPercentage,
			v.Total, totalPercentage,
			v.TotalAdmin,
			v.Pfizer,
			v.AstraZeneca,
			v.Moderna,
			v.Janssen,
		)
	} else {
		description = p.Sprintf(`
**First installment**: %d (+%d) (%.2f%% of population)
**Second installment**: %d (+%d) (%.2f%% of population)
**Fully vaccinated**: %d (+%d) (%.2f%% of population)
**Total Administered**: %d (+%d) 

***Pfizer***: %d (+%d) 
***AstraZeneca***: %d (+%d) 
***Moderna***: %d (+%d) 
***J&J/Janssen***: %d (+%d) 
                `,
			v.First, int64(math.Abs(float64(v.First-prev.First))), firstPercentage,
			v.Second, int64(math.Abs(float64(v.Second-prev.Second))), secondPercentage,
			v.Total, int64(math.Abs(float64(v.Total-prev.Total))), totalPercentage,
			v.TotalAdmin, int64(math.Abs(float64(v.TotalAdmin-prev.TotalAdmin))),
			v.Pfizer, int64(math.Abs(float64(v.Pfizer-prev.Pfizer))),
			v.AstraZeneca, int64(math.Abs(float64(v.AstraZeneca-prev.AstraZeneca))),
			v.Moderna, int64(math.Abs(float64(v.Moderna-prev.Moderna))),
			v.Janssen, int64(math.Abs(float64(v.Janssen-prev.Janssen))),
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
	respTypes, err := http.Get(covidVaccinesType)
	if err != nil {
		return nil, err
	}
	vaccineTypes := &struct {
		Features []struct {
			Attributes struct {
				Date    int `json:"relDate"`
				Pfizer  int `json:"pf"`
				Moderna int `json:"modern"`
				Az      int `json:"az"`
				Janssen int `json:"janssen"`
			} `json:"attributes"`
		} `json:"features"`
	}{}
	if err = json.NewDecoder(respTypes.Body).Decode(vaccineTypes); err != nil {
		return nil, err
	}

	if len(vaccines.Features) < 1 || len(vaccineTypes.Features) < 1 {
		return nil, errors.New("no features")
	}
	return &Vaccines{
		First:       vaccines.Features[0].Attributes.First,
		Second:      vaccines.Features[0].Attributes.Second,
		Total:       vaccines.Features[0].Attributes.Second + vaccineTypes.Features[0].Attributes.Janssen,
		TotalAdmin:  vaccines.Features[0].Attributes.Total,
		Pfizer:      vaccineTypes.Features[0].Attributes.Pfizer,
		Moderna:     vaccineTypes.Features[0].Attributes.Moderna,
		AstraZeneca: vaccineTypes.Features[0].Attributes.Az,
		Janssen:     vaccineTypes.Features[0].Attributes.Janssen,
		Date:        time.Unix(int64(vaccines.Features[0].Attributes.Date)/1000, 0),
	}, nil
}
