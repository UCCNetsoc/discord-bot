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
	First  int
	Second int
	Date   time.Time
}

// Population of Ireland
const Population = 4900000

func (v *Vaccines) Embed(prev *Vaccines) *discordgo.MessageEmbed {
	p := message.NewPrinter(language.English)
	firstPercentage := (float64(v.First) / float64(Population)) * 100
	secondPercentage := (float64(v.Second) / float64(Population)) * 100
	var description string
	if prev == nil {
		description = p.Sprintf(`
				**First installment**: %d (%.2f%% of population)
				**Second installment**: %d (%.2f%% of population)
			`, v.First, firstPercentage, v.Second, secondPercentage)
	} else {
		description = p.Sprintf(`
				__**New**__
				**First installment**: %d (+%d) (%.2f%% of population)
				**Second installment**: %d (+%d) (%.2f%% of population)

				__**Previously**__
				**First installment**: %d
				**Second installment**: %d
			`,
			v.First,
			int64(math.Abs(float64(v.First-prev.First))),
			firstPercentage,
			v.Second,
			int64(math.Abs(float64(v.Second-prev.Second))),
			secondPercentage,
			prev.First,
			prev.Second,
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
				First  int `json:"total_number_of_1st_dose_admini"`
				Second int `json:"total_number_of_2nd_dose_admini"`
				Date   int `json:"data_relevent_up_to_date"`
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
		First:  vaccines.Features[0].Attributes.First,
		Second: vaccines.Features[0].Attributes.Second,
		Date:   time.Unix(int64(vaccines.Features[0].Attributes.Date)/1000, 0),
	}, nil
}
