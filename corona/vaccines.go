package corona

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (v *Vaccines) Embed() *discordgo.MessageEmbed {
	p := message.NewPrinter(language.English)
	return embed.NewEmbed().SetTitle("Vaccines Rollout in Ireland").SetDescription(p.Sprintf(`
				**First installment**: %d
				**Second installment**: %d
			`, v.First, v.Second)).SetFooter(fmt.Sprintf("As of %s", v.Date.Format(layoutIE))).MessageEmbed
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
